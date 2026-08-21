#!/usr/bin/env bash
# Smoke test for the crossplane module.
#
# WHY THIS EXISTS. The module's CI (this-test-modules.yaml) runs `dagger
# functions`, `dagger develop` and golangci-lint — all of which pass whether or
# not the container can actually be built and whether or not the binaries
# inside it work. Everything this module does at runtime hinges on two
# downloads, and both of them changed under us:
#
#   - the CLI moved repo + bucket + binary name at v2.3.0 (crank -> crossplane,
#     releases.crossplane.io -> cli.crossplane.io), so a version bump alone
#     yields a 404;
#   - render needs a SECOND binary, the Crossplane core, because CLI >= v2.3.0
#     shells out to `crossplane internal render` — which the CLI itself cannot
#     do (see container.go).
#
# None of that is visible to a compile-time check. This script builds the real
# container and asserts the four properties verify.go depends on.
#
# Runs without secrets, which is what makes it CI-eligible (unlike
# `task test-crossplane`, which pushes to a registry).
#
# Kept as a script rather than inline Taskfile cmds so CI can run it without
# Task, whose remote `includes:` needs an interactive trust prompt.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

MODULE=crossplane

# EXPECTED VERSIONS ARE READ FROM container.go SO THE PINS STAY SINGLE-SOURCED.
# A bump there without a matching bump here would otherwise pass silently.
read_pin() {
  sed -n "s/.*$1 *= *\"\(.*\)\".*/\1/p" "${MODULE}/container.go" | head -1
}
cli_want=$(read_pin crossplaneVersion)
core_want=$(read_pin crossplaneCoreVersion)

for pair in "crossplaneVersion:${cli_want}" "crossplaneCoreVersion:${core_want}"; do
  if [ -z "${pair#*:}" ]; then
    echo "FAIL: could not read ${pair%%:*} from ${MODULE}/container.go"
    exit 1
  fi
done
echo "pins from container.go: CLI ${cli_want}, core ${core_want}"

# One container build, one exec: cheaper than four round trips, and it keeps
# the assertions below reading against a single captured output.
# NOTE: dagger parses --args as CSV, so the probe below must contain neither a
# comma nor a double quote — both terminate the field and produce
# "bare \" in non-quoted-field". Hence `grep -o -- --crossplane-binary` with the
# pattern unquoted after `--` rather than the more obvious quoted form.
probe='crossplane version --client; crossplane-core --version; crossplane-core internal render --help 2>&1 | head -1; crossplane render --help 2>&1 | grep -o -- --crossplane-binary | head -1'

out=$(dagger call -m "./${MODULE}" get-xplane-container \
  with-exec --args sh,-c,"${probe}" \
  stdout)
echo "${out}"

fail=0
assert() { # assert <needle> <description>
  if echo "${out}" | grep -qF -- "$1"; then
    echo "OK: $2"
  else
    echo "FAIL: $2 (missing: $1)"
    fail=1
  fi
}

assert "Client Version: ${cli_want}" "CLI is ${cli_want}"
assert "${core_want}" "core binary is ${core_want}"
# The core's whole reason for being installed. Without it every render dies
# per-XR with "unexpected argument internal".
assert "crossplane internal render" "core provides 'internal render'"
# verify.go probes for this flag and silently renders without it if absent —
# which would put the floating crossplane:stable container back in the path.
assert "--crossplane-binary" "CLI supports --crossplane-binary"

exit "${fail}"
