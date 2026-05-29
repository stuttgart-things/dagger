package main

import (
	"context"
	"dagger/crossplane/internal/dagger"
	"fmt"
)

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
// `crossplane render` is run inside the container via a docker-in-docker
// sidecar, so function images are pulled and executed without needing a
// Docker socket on the host.
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

	dockerd := dag.Container().
		From("docker:dind").
		WithMountedCache("/var/lib/docker", dag.CacheVolume("crossplane-verify-dind")).
		WithExposedPort(2375).
		AsService(dagger.ContainerAsServiceOpts{
			Args:                     []string{"dockerd", "--host=tcp://0.0.0.0:2375", "--tls=false"},
			InsecureRootCapabilities: true,
		})

	out, err := m.XplaneContainer.
		WithServiceBinding("dockerd", dockerd).
		WithEnvVariable("DOCKER_HOST", "tcp://dockerd:2375").
		WithEnvVariable("PROVIDER_K8S_VERSION", providerKubernetesVersion).
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithNewFile("/usr/local/bin/verify.sh", verifyScript).
		WithExec([]string{"sh", "/usr/local/bin/verify.sh"}).
		Stdout(ctx)

	if err != nil {
		return out, fmt.Errorf("verify failed: %w", err)
	}
	return out, nil
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

  # Render the Composition. Always read functions.yaml from examples/.
  # ${EXTRA_ARGS} (unquoted, may be empty) supplies EnvironmentConfigs.
  if RENDERED=$(crossplane render "${xr}" apis/composition.yaml examples/functions.yaml ${EXTRA_ARGS} 2>/tmp/render.err); then
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
