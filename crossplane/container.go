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

	// crossplaneChannel selects the release channel. Used as the bucket path
	// for BOTH downloads below.
	crossplaneChannel = "stable"

	// ZWEI BINAERIES, ZWEI BUCKETS, ZWEI REPOS. Das ist der Kern hier.
	//
	// Seit CLI v2.3.0 sind CLI und Core GETRENNT:
	//
	//   CLI   crossplane/cli          cli.crossplane.io        `crossplane`
	//         kann render/resource trace/xpkg
	//   Core  crossplane/crossplane   releases.crossplane.io   `crossplane`
	//         kann `internal render` — die CLI kann das NICHT
	//
	// Vor v2.3.0 lagen beide zusammen und die CLI hiess im alten Bucket
	// `crank`. Deshalb ergibt ein reiner Versions-Bump ohne Bucket- und
	// Namenswechsel einen 404.
	//
	// WARUM DER CORE UEBERHAUPT GEBRAUCHT WIRD.
	//
	// Ab CLI v2.3.0 fuehrt `crossplane render` die Pipeline nicht mehr selbst
	// aus, sondern ruft `crossplane internal render` auf — per Default in
	// einem Docker-Container aus dem FLOATING Tag
	// xpkg.crossplane.io/crossplane/crossplane:stable. Beides ist hier falsch:
	//
	//   1. Floating Tag. Genau so ist Verify im Juni ohne eine einzige
	//      Code-Aenderung umgefallen (#295), damals ueber den CLI-Kanal.
	//   2. Docker im Sandbox. Verschachteltes Docker funktioniert in diesem
	//      Modul nicht — deshalb laufen die Functions seit #300 als
	//      Dagger-Services. Ein Core-CONTAINER wuerde genau das zurueckholen.
	//
	// Gemessen (2026-08-21): die Development-Runtime schuetzt NICHT davor —
	// auch mit allen Functions als Services startet die CLI den Core-Container.
	//
	// Loesung: den Core als BINARY mitinstallieren und render per
	// --crossplane-binary darauf zeigen lassen. Damit entfaellt der Container
	// vollstaendig und der Tag ist gepinnt statt floating.
	//
	// crossplaneVersion pinnt die CLI, crossplaneCoreVersion den Core. Beide
	// exakt, nie ein floating `current`.
	crossplaneVersion     = "v2.4.1"
	crossplaneCoreVersion = "v2.4.0"
)

// crossplaneInstall downloads and installs BOTH binaries the render path
// needs — the CLI (/usr/bin/crossplane) and the Crossplane core
// (/usr/bin/crossplane-core, which owns `internal render`). See the constants
// above for why these are two artifacts from two buckets.
//
// Integrity, and why the two are treated differently. A bare
// `curl … --output crossplane` silently writes whatever the CDN returns — a
// truncated stream, a rate-limit page, an HTML 4xx body — to a file that then
// executes as a shell script and dies with "line 2: syntax error: unexpected
// newline" (the "no ELF magic, fall back to /bin/sh" signature). So every
// download here uses curl -f + --retry, is checksum-checked where the bucket
// publishes a usable checksum, is executed as a final gate, and is retried as
// a whole on any failure.
//
//	core  STRICT   releases.crossplane.io publishes a matching sha256 for the
//	              pinned release (verified 2026-08-21: dfe07a50… published ==
//	              served). NOTE: older core releases do NOT verify — v2.2.2
//	              serves df012171… while publishing 336aabd0…, so pinning
//	              crossplaneCoreVersion back below v2.3 makes this install
//	              refuse with a checksum error. That is the right outcome for
//	              the wrong-looking reason: those cores have no `internal
//	              render` either, so they could never serve this purpose.
//	CLI   WARN     cli.crossplane.io publishes a sha256 that does NOT match
//	              the served binary — and not through CDN mangling: the
//	              checksum shipped INSIDE the official bundle tarball
//	              (ad0661ee…) disagrees with the binary next to it in the same
//	              tarball (963ae072…). Measured 2026-08-21, same symptom the
//	              old crank.sha256 had (#295). A mismatch is therefore a
//	              WARNING and the run-check is the authoritative gate: a
//	              truncated or HTML body cannot execute, so it still fails and
//	              retries; only a valid binary with a wrong published checksum
//	              is allowed through.
const crossplaneInstall = `#!/bin/sh
set -u

CHANNEL="${CROSSPLANE_CHANNEL:-stable}"
CLI_VERSION="${CROSSPLANE_VERSION:?CROSSPLANE_VERSION must be set}"
CORE_VERSION="${CROSSPLANE_CORE_VERSION:?CROSSPLANE_CORE_VERSION must be set}"

# Bucket and binary name follow the CLI version, exactly as the upstream
# cli.crossplane.io/install.sh does: v2.3.0 is where the CLI moved to its own
# repo and bucket and lost the "crank" name. Encoded here rather than hardcoded
# so pinning CROSSPLANE_VERSION back to a 2.2.x release keeps working.
_v=$(echo "${CLI_VERSION}" | sed -e 's/^v//' -e 's/-.*//')
_maj=$(echo "${_v}" | cut -d. -f1)
_min=$(echo "${_v}" | cut -d. -f2)
if [ "${_maj}" -lt 2 ] || { [ "${_maj}" -eq 2 ] && [ "${_min}" -lt 3 ]; }; then
  CLI_URL="https://releases.crossplane.io/${CHANNEL}/${CLI_VERSION}/bin/linux_amd64/crank"
else
  CLI_URL="https://cli.crossplane.io/${CHANNEL}/${CLI_VERSION}/bin/linux_amd64/crossplane"
fi
CORE_URL="https://releases.crossplane.io/${CHANNEL}/${CORE_VERSION}/bin/linux_amd64/crossplane"

# fetch_install <url> <dest> <strict|warn> <run-check command...>
fetch_install() {
  _url="$1"; _dest="$2"; _mode="$3"; shift 3
  _i=1
  while [ "${_i}" -le 3 ]; do
    rm -f /tmp/dl /tmp/dl.sha256
    if curl -fsSL --retry 5 --retry-delay 2 --retry-all-errors "${_url}" -o /tmp/dl; then
      if curl -fsSL --retry 3 --retry-delay 2 "${_url}.sha256" -o /tmp/dl.sha256 2>/dev/null \
           && [ -s /tmp/dl.sha256 ]; then
        _want=$(awk '{print $1}' /tmp/dl.sha256)
        _got=$(sha256sum /tmp/dl | awk '{print $1}')
        if [ "${_want}" != "${_got}" ]; then
          if [ "${_mode}" = "strict" ]; then
            echo "install ${_dest}: checksum MISMATCH (want ${_want}, got ${_got})" >&2
            _i=$((_i + 1)); sleep 2; continue
          fi
          echo "install ${_dest}: WARNING checksum mismatch (want ${_want}, got ${_got}); relying on run-check" >&2
        fi
      fi
      install -m 0755 /tmp/dl "${_dest}"
      if "$@" >/dev/null 2>&1; then
        echo "install ${_dest}: ok"
        return 0
      fi
      echo "install ${_dest}: installed binary failed to run, attempt ${_i}" >&2
    else
      echo "install ${_dest}: download failed, attempt ${_i}" >&2
    fi
    _i=$((_i + 1)); sleep 2
  done
  echo "install ${_dest}: failed after retries" >&2
  return 1
}

echo "crossplane install: CLI ${CHANNEL} ${CLI_VERSION} from ${CLI_URL}"
fetch_install "${CLI_URL}" /usr/bin/crossplane warn /usr/bin/crossplane version --client || exit 1

# The run-check is deliberately "internal render --help": that subcommand is the
# ONLY reason this binary is here, and it is exactly what the CLI cannot do. A
# core image/binary without it would otherwise install cleanly and fail later,
# per XR, with "unexpected argument internal".
echo "crossplane install: core ${CHANNEL} ${CORE_VERSION} from ${CORE_URL}"
fetch_install "${CORE_URL}" /usr/bin/crossplane-core strict /usr/bin/crossplane-core internal render --help || exit 1

exit 0
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
		WithEnvVariable("CROSSPLANE_CORE_VERSION", crossplaneCoreVersion).
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
