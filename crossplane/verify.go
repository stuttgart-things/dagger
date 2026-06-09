package main

import (
	"context"
	"dagger/crossplane/internal/dagger"
	"fmt"
	"strings"
)

// functionPort is the gRPC port a Composition Function listens on (crossplane's
// own Docker runtime maps the same port). Functions are served insecurely
// (plaintext h2c) via the --insecure flag, which the Development runtime dials.
const functionPort = 9443

// Verify performs offline, container-pinned verification of a Crossplane
// Configuration package. It runs four checks per example XR:
//
//  1. crossplane.yaml correctness (xpkg build, once per Configuration)
//  2. Layer 1 — each examples/xr*.yaml validated against the Configuration's own XRD
//  3. Layer 2 — each rendered kubernetes.m.crossplane.io Object validated
//     against the provider-kubernetes CRD schema
//  4. Layer 3 — each embedded `spec.forProvider.manifest` validated against
//     the built-in Kubernetes schemas (kubeconform default). Embedded
//     manifests whose kind is a CRD the harness has no schema for (e.g. a
//     Tekton PipelineRun) are skipped via -ignore-missing-schemas rather
//     than failing the layer — their correctness is owned by the renderer.
//
// Each Composition Function referenced by examples/functions.yaml is started as
// a Dagger service container and `crossplane render` is pointed at it via the
// "Development" runtime. crossplane render's default runtime instead starts each
// Function as a *nested* Docker container and dials its mapped gRPC port; inside
// this module's sandbox that nested Docker can't bring the container up (cgroup-v2
// subtree_control delegation fails on the runner), so render hangs for the full
// ~60s function-start grace and then fails with DeadlineExceeded / connection
// refused for any Docker-runtime Function (e.g. function-kcl). Running the
// Functions as Dagger services removes the nested Docker — and its per-XR 60s
// penalty — entirely. See stuttgart-things/dagger#300.
func (m *Crossplane) Verify(
	ctx context.Context,
	// a single Configuration directory (containing crossplane.yaml, apis/, examples/)
	src *dagger.Directory,
	// +optional
	// +default="v1.2.0"
	// provider-kubernetes version whose CRD schemas are used for Layer 2.
	// Should match the dependsOn entry in crossplane.yaml.
	providerKubernetesVersion string,
) (string, error) {

	if m.XplaneContainer == nil {
		m.XplaneContainer = m.GetXplaneContainer(ctx)
	}

	base := m.XplaneContainer.
		WithEnvVariable("PROVIDER_K8S_VERSION", providerKubernetesVersion).
		WithDirectory("/src", src).
		WithWorkdir("/src")

	// Discover the Functions the Configuration's pipeline uses and run each as a
	// Dagger service. verify.sh rewrites functions.yaml so render dials these
	// services instead of starting Docker containers (see verifyScript).
	fns, err := discoverFunctions(ctx, base)
	if err != nil {
		return "", fmt.Errorf("verify: discovering functions: %w", err)
	}

	container := base.WithNewFile("/usr/local/bin/verify.sh", verifyScript)
	for _, fn := range fns {
		svc := dag.Container().
			From(fn.Package).
			WithExposedPort(functionPort).
			// crossplane's own Docker runtime runs the function image with
			// exactly this arg; --insecure serves plaintext gRPC, which the
			// Development runtime dials. UseEntrypoint:true appends --insecure to
			// the image's entrypoint (the function server binary) — without it
			// Dagger would try to exec "--insecure" as the command. Dagger waits
			// for the exposed port to accept connections before running
			// verify.sh, so the function is up before render — no readiness race.
			AsService(dagger.ContainerAsServiceOpts{
				Args:          []string{"--insecure"},
				UseEntrypoint: true,
			})
		// Bind under the Function's own name so the in-container render target
		// "dns:///<name>:9443" (set by verify.sh) resolves to this service.
		container = container.WithServiceBinding(fn.Name, svc)
	}

	out, err := container.
		WithExec([]string{"sh", "/usr/local/bin/verify.sh"}).
		Stdout(ctx)

	if err != nil {
		return out, fmt.Errorf("verify failed: %w", err)
	}
	return out, nil
}

// composFunction is a Composition Function discovered in examples/functions.yaml.
type composFunction struct {
	Name    string // metadata.name — also the service binding hostname
	Package string // spec.package — the runnable function runtime image
}

// discoverFunctions reads examples/functions.yaml and returns the Functions it
// declares. A missing or Function-less file yields no functions (not an error):
// render then runs unchanged and surfaces any resulting error itself.
func discoverFunctions(ctx context.Context, base *dagger.Container) ([]composFunction, error) {
	// Per-document eval (not `ea`): with evaluate-all, `.metadata.name` and
	// `.spec.package` each iterate over every Function node and `+` cross-products
	// them, scrambling name↔package pairs. Default eval runs the expression once
	// per document, keeping each pair intact. `// ""` guards a Function with no
	// spec.package against yq's null-string concat error; `|| true` so a missing
	// functions.yaml (yq exits non-zero) yields empty output, not a failed exec.
	raw, err := base.
		WithExec([]string{"sh", "-c",
			`yq 'select(.kind == "Function") | .metadata.name + " " + (.spec.package // "")' examples/functions.yaml 2>/dev/null || true`}).
		Stdout(ctx)
	if err != nil {
		return nil, err
	}

	var fns []composFunction
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue // skip blanks and Functions missing a package
		}
		fns = append(fns, composFunction{Name: fields[0], Package: fields[1]})
	}
	return fns, nil
}

// verifyScript implements the three-layer pipeline described above. It is
// intentionally a shell script so the entire flow runs in a single container
// exec — cheaper than chaining one Dagger exec per step.
const verifyScript = `#!/bin/sh
set -u

OK=1
CONFIG=$(yq -r '.metadata.name // ""' crossplane.yaml 2>/dev/null)
[ -z "${CONFIG}" ] && CONFIG=$(basename "$(pwd)")
echo "Configuration: ${CONFIG}"

# ---- crossplane.yaml + Composition structure check -----------------------
if BUILD_ERR=$(crossplane xpkg build 2>&1); then
  echo "  ✓ xpkg build"
else
  echo "  ✗ xpkg build"
  echo "${BUILD_ERR}" | sed 's/^/      /'
  OK=0
fi

# ---- Generate JSON schemas for XRD (Layer 1) -----------------------------
mkdir -p /schemas/xrds /schemas/provider
if [ -f apis/definition.yaml ]; then
  # openapi2jsonschema only processes kind=CustomResourceDefinition; rewrite
  # the XRD to that kind so the same converter handles both XRDs and CRDs.
  yq '.kind = "CustomResourceDefinition"' apis/definition.yaml > /tmp/xrd-as-crd.yaml
  if ! XRDGEN=$(cd /schemas/xrds && openapi2jsonschema /tmp/xrd-as-crd.yaml 2>&1); then
    echo "  ! XRD schema generation failed:"
    printf '%s\n' "${XRDGEN}" | sed 's/^/      /'
  fi
fi

# ---- Generate JSON schemas for provider-kubernetes (Layer 2) -------------
# Cached per provider version via a marker file.
PROV_MARK="/schemas/provider/.ready-${PROVIDER_K8S_VERSION}"
if [ ! -f "${PROV_MARK}" ]; then
  rm -rf /schemas/provider && mkdir -p /schemas/provider
  crane export "xpkg.crossplane.io/crossplane-contrib/provider-kubernetes:${PROVIDER_K8S_VERSION}" /tmp/provider.tar >/dev/null 2>&1
  mkdir -p /tmp/provider-extract
  tar -xf /tmp/provider.tar -C /tmp/provider-extract package.yaml
  # Include the full group in the filename so the legacy and modern Object
  # CRDs (kubernetes.crossplane.io vs kubernetes.m.crossplane.io) don't collide.
  if ! PROVGEN=$(cd /schemas/provider && FILENAME_FORMAT="{kind}_{fullgroup}_{version}" openapi2jsonschema /tmp/provider-extract/package.yaml 2>&1); then
    echo "  ! provider-kubernetes schema generation failed:"
    printf '%s\n' "${PROVGEN}" | sed 's/^/      /'
  fi
  : > "${PROV_MARK}"
fi


# ---- Collect EnvironmentConfigs for render (Layer 0.5) -------------------
# An XR that sets spec.environmentConfig makes function-environment-configs
# select an EnvironmentConfig by label. Offline crossplane render has no
# cluster to select from, so it fails fatally with "expected exactly one
# required resource, got 0". Pass any EnvironmentConfig manifests shipped in
# examples/ as --extra-resources so the selector resolves. No-op when the
# Configuration ships none.
EXTRA_ARGS=""
yq ea 'select(.kind == "EnvironmentConfig")' examples/*.yaml > /tmp/extra-resources.yaml 2>/dev/null || true
if [ -s /tmp/extra-resources.yaml ]; then
  EXTRA_ARGS="--extra-resources /tmp/extra-resources.yaml"
fi

# ---- Point Functions at their Dagger service containers (Layer 0.6) ------
# The module starts each Function as a Dagger service reachable at
# "<function-name>:9443" (see Verify in verify.go). crossplane render's default
# runtime would instead start each Function as a nested Docker container, which
# can't work inside this sandbox (#300). Rewrite functions.yaml so every Function
# uses the "Development" runtime targeting its service endpoint; render then
# dials the already-running service instead of launching a container.
FUNCTIONS_FILE=examples/functions.yaml
if [ -f examples/functions.yaml ]; then
  if yq ea 'with(select(.kind == "Function");
        .metadata.annotations["render.crossplane.io/runtime"] = "Development" |
        .metadata.annotations["render.crossplane.io/runtime-development-target"] = ("dns:///" + .metadata.name + ":9443"))' \
      examples/functions.yaml > /tmp/functions.dev.yaml 2>/dev/null \
      && [ -s /tmp/functions.dev.yaml ]; then
    FUNCTIONS_FILE=/tmp/functions.dev.yaml
  fi
fi

# ---- Per-XR loop ---------------------------------------------------------
FOUND_XR=0
for xr in examples/xr*.yaml; do
  [ -f "${xr}" ] || continue
  FOUND_XR=1
  name=$(basename "${xr}")
  status=""
  details=""

  # Layer 1: XR ↔ XRD
  if ERR=$(kubeconform -strict \
      -schema-location default \
      -schema-location '/schemas/xrds/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json' \
      "${xr}" 2>&1); then
    status="XRD-valid"
  else
    status="XRD-INVALID"
    details="${details}\n${ERR}"
    OK=0
  fi

  # Render the Composition. ${FUNCTIONS_FILE} is the Development-runtime rewrite
  # of examples/functions.yaml (falls back to the original if the rewrite was a
  # no-op). ${EXTRA_ARGS} (unquoted, may be empty) supplies EnvironmentConfigs.
  if RENDERED=$(crossplane render "${xr}" apis/composition.yaml "${FUNCTIONS_FILE}" ${EXTRA_ARGS} 2>/tmp/render.err); then
    status="${status}, render-ok"

    # Layer 2: Object wrapper
    echo "${RENDERED}" | yq 'select(.kind == "Object")' > /tmp/objects.yaml
    if [ -s /tmp/objects.yaml ]; then
      if ERR=$(kubeconform -strict \
          -schema-location default \
          -schema-location '/schemas/provider/{{.ResourceKind}}_{{.Group}}_{{.ResourceAPIVersion}}.json' \
          /tmp/objects.yaml 2>&1); then
        status="${status}, object-valid"
      else
        status="${status}, object-INVALID"
        details="${details}\n${ERR}"
        OK=0
      fi
    fi

    # Layer 3: embedded manifests inside spec.forProvider.manifest
    # Embedded manifests may be arbitrary CRDs (e.g. Tekton PipelineRun) the
    # harness has no schema for. -ignore-missing-schemas skips those (so they
    # don't hard-fail the layer) while still strictly validating any embedded
    # core Kubernetes resources whose schema kubeconform can resolve.
    echo "${RENDERED}" | yq 'select(.kind == "Object") | .spec.forProvider.manifest' > /tmp/embedded.yaml
    if [ -s /tmp/embedded.yaml ]; then
      if ERR=$(kubeconform -strict -ignore-missing-schemas /tmp/embedded.yaml 2>&1); then
        status="${status}, embedded-valid"
      else
        status="${status}, embedded-INVALID"
        details="${details}\n${ERR}"
        OK=0
      fi
    fi
  else
    status="${status}, render-FAILED"
    details="${details}\n$(cat /tmp/render.err)"
    OK=0
  fi

  case "${status}" in
    *INVALID*|*FAILED*)
      printf "  ✗ %s: %s\n" "${name}" "${status}"
      printf "%b\n" "${details}" | sed 's/^/      /'
      ;;
    *)
      printf "  ✓ %s: %s\n" "${name}" "${status}"
      ;;
  esac
done

if [ "${FOUND_XR}" = "0" ]; then
  echo "  (no examples/xr*.yaml found — Layers 1-3 skipped)"
fi

[ "${OK}" = "1" ] || exit 1
`
