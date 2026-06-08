package main

import (
	"context"
	"dagger/crossplane/internal/dagger"
)

const (
	kubeconformVersion = "v0.6.7"
	// openapi2jsonschema.py is pinned to the same kubeconform release so the
	// script and binary evolve together.
	openapi2jsonschemaURL = "https://raw.githubusercontent.com/yannh/kubeconform/" + kubeconformVersion + "/scripts/openapi2jsonschema.py"

	// crossplaneChannel selects the crossplane (crank) CLI release channel.
	// Used only as the bucket path and as the fallback when crossplaneVersion
	// is empty.
	crossplaneChannel = "stable"

	// crossplaneVersion pins the crank CLI to an exact release instead of
	// tracking the channel's floating `current` marker.
	//
	// Why this is pinned: crank v2.3.0 changed `crossplane render` to run the
	// pipeline by re-exec'ing `crossplane internal render` inside the function
	// Docker container. The crossplane-contrib function images still ship an
	// older crossplane binary that has no `internal` subcommand, so every
	// render fails with:
	//
	//   cannot run crossplane internal render in Docker: container exited with
	//   status 1: crossplane: error: unexpected argument internal
	//
	// When `current` advanced to v2.3.0 this silently broke Verify for every
	// function-based Configuration (it tracked the channel, so the same module
	// tag started failing without any code change). v2.2.2 is the latest
	// release whose render still runs functions the compatible way. Bump this
	// once the function images (or a later crank) resolve the skew.
	// See stuttgart-things/dagger#295.
	crossplaneVersion = "v2.2.2"
)

// crossplaneInstall downloads and installs the crossplane (crank) CLI with
// integrity checks. A bare `curl … --output crossplane` silently writes
// whatever the CDN returns — a truncated stream, a rate-limit page, or an
// HTML 4xx body — to /usr/bin/crossplane, which then executes as a shell
// script and dies with "line 2: syntax error: unexpected newline" (the
// classic "no ELF magic, fall back to /bin/sh" signature). This installer
// instead: (1) uses curl -f + --retry so HTTP errors and transient blips
// fail/retry rather than producing a bad file; (2) verifies the published
// SHA256 when present (catches truncation and HTML-200 bodies); (3) executes
// the installed binary as a final sanity gate; and (4) retries the whole
// install on any failure, so a single corrupt fetch can't poison the image.
const crossplaneInstall = `#!/bin/sh
set -u

CHANNEL="${CROSSPLANE_CHANNEL:-stable}"
BASE="https://releases.crossplane.io/${CHANNEL}"

# Prefer an explicit pin (CROSSPLANE_VERSION). Only fall back to the channel's
# floating 'current' marker when no pin is set — tracking 'current' is what let
# crank v2.3.0 silently break render (see crossplaneVersion comment above).
VERSION="${CROSSPLANE_VERSION:-}"
if [ -z "${VERSION}" ]; then
  # Resolve the channel to an exact version so the binary and its checksum are
  # fetched from the same immutable path.
  VERSION=$(curl -fsSL --retry 5 --retry-delay 2 --retry-all-errors "${BASE}/current/version" | tr -d '[:space:]')
fi
if [ -z "${VERSION}" ]; then
  echo "crossplane install: could not resolve ${CHANNEL} version" >&2
  exit 1
fi
URL="${BASE}/${VERSION}/bin/linux_amd64/crank"
echo "crossplane install: ${CHANNEL} ${VERSION}"

i=1
while [ "${i}" -le 3 ]; do
  rm -f /tmp/crossplane /tmp/crossplane.sha256
  if curl -fsSL --retry 5 --retry-delay 2 --retry-all-errors "${URL}" -o /tmp/crossplane; then
    # Integrity: compare against the published checksum when the bucket serves
    # one. NOTE: the crossplane release bucket currently serves crank.sha256
    # files that do NOT match the served binaries (observed for every version
    # via CloudFront, 2026-06). A mismatch is therefore treated as a WARNING,
    # not a hard failure: we fall through to the run-check below, which is the
    # authoritative gate. A truncated body or HTML error page cannot execute,
    # so it still fails the run-check and retries; only a valid binary with a
    # stale/wrong published checksum is allowed through.
    # See stuttgart-things/dagger#295.
    if curl -fsSL --retry 5 --retry-delay 2 --retry-all-errors "${URL}.sha256" -o /tmp/crossplane.sha256 2>/dev/null \
         && [ -s /tmp/crossplane.sha256 ]; then
      want=$(awk '{print $1}' /tmp/crossplane.sha256)
      got=$(sha256sum /tmp/crossplane | awk '{print $1}')
      if [ "${want}" != "${got}" ]; then
        echo "crossplane install: WARNING checksum mismatch (want ${want}, got ${got}); relying on run-check" >&2
      fi
    fi
    install -m 0755 /tmp/crossplane /usr/bin/crossplane
    # Authoritative gate: a real binary runs; an HTML/truncated/corrupt file
    # does not. This also covers the case where the bucket served no .sha256
    # (or a wrong one) to verify.
    if crossplane version --client >/dev/null 2>&1; then
      echo "crossplane install: ok"
      exit 0
    fi
    echo "crossplane install: installed binary failed to run, attempt ${i}" >&2
  else
    echo "crossplane install: download failed, attempt ${i}" >&2
  fi
  i=$((i + 1)); sleep 2
done

echo "crossplane install: failed after retries" >&2
exit 1
`

// GetXplaneContainer returns the default image for Crossplane with crossplane and kcl2xrd installed
func (m *Crossplane) GetXplaneContainer(ctx context.Context) *dagger.Container {
	return dag.Container().
		From("cgr.dev/chainguard/wolfi-base:latest").
		// Install dependencies. python + py3-pyyaml + crane back the Verify pipeline
		// (CRD schema extraction + provider image export); curl + yq are also used elsewhere.
		WithExec([]string{"apk", "add", "curl", "yq", "crane", "python-3.13", "py3-pyyaml"}).
		// Install crossplane (crank) via an integrity-checked installer. A bare
		// download writes any CDN error/truncation straight to the binary, which
		// then dies as a "line 2: syntax error" the next time it is invoked.
		WithEnvVariable("CROSSPLANE_CHANNEL", crossplaneChannel).
		WithEnvVariable("CROSSPLANE_VERSION", crossplaneVersion).
		WithNewFile("/tmp/install-crossplane.sh", crossplaneInstall).
		WithExec([]string{"sh", "/tmp/install-crossplane.sh"}).
		// Install kcl2xrd
		WithExec([]string{"curl", "-L", "https://github.com/ggkhrmv/kcl2xrd/releases/download/v0.8.0/kcl2xrd-linux-amd64", "--output", "kcl2xrd"}).
		WithExec([]string{"mv", "kcl2xrd", "/usr/bin/kcl2xrd"}).
		WithExec([]string{"chmod", "+x", "/usr/bin/kcl2xrd"}).
		// Install kubeconform (release binary; not in wolfi apk)
		WithExec([]string{"sh", "-c",
			"curl -sL https://github.com/yannh/kubeconform/releases/download/" + kubeconformVersion +
				"/kubeconform-linux-amd64.tar.gz | tar -xz -C /usr/bin kubeconform"}).
		WithExec([]string{"chmod", "+x", "/usr/bin/kubeconform"}).
		// Install openapi2jsonschema (CRD -> JSON Schema converter used by Verify).
		// Filenames are lowercase, which matches kubeconform's {{.ResourceKind}}
		// template (it lowercases the kind from the resource).
		WithExec([]string{"curl", "-sL", openapi2jsonschemaURL, "-o", "/usr/bin/openapi2jsonschema"}).
		WithExec([]string{"chmod", "+x", "/usr/bin/openapi2jsonschema"})
}
