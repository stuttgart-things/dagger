#!/usr/bin/env bash
# Smoke test for the cosign module.
#
# WHY THIS EXISTS. The module's CI (this-test-modules.yaml) runs `dagger
# functions`, `dagger develop` and golangci-lint -- all of which pass whether or
# not the cosign in the container is the pinned one, and whether or not the
# verification functions can tell a signature by the right identity from one by
# any other. Those are runtime properties:
#
#   - a verification that has never refused anything cannot be told apart from
#     one that cannot refuse, so VerifyRefuses has to pass on a wrong identity
#     and fail on the right one;
#   - a refusal only counts when cosign refused the identity, not when the
#     artefact carried no signature to refuse;
#   - Sign and Attest have to refuse a tag, and an empty predicate, before
#     anything is signed;
#   - the identity token must not show up in dagger's output (#318).
#
# Signing itself cannot run here: it needs an OIDC token from a job with
# `id-token: write`. What is checked instead is that a token reaches cosign
# and goes no further.
#
# Needs network (ghcr.io, Docker Hub, Sigstore's TUF root) but no secrets and no
# registry login, which is what makes it CI-eligible.
#
# Kept as a script rather than inline Taskfile cmds so CI can run it without
# Task, whose remote `includes:` needs an interactive trust prompt.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

MODULE=cosign

# FIXTURE: an image signed on main by schmetterpause's CI, with a CycloneDX
# attestation. It is pinned by digest, so it has to stay in its registry; if it
# is ever deleted, pick another .sig/.att pair from that package.
SIGNED=ghcr.io/stuttgart-things/schmetterpause@sha256:72be6d814478927f4894ac586fd88bd19cb0a570792e84dda3bfc4147a07c852
ISSUER=https://token.actions.githubusercontent.com
IDENTITY=https://github.com/stuttgart-things/schmetterpause/.github/workflows/ci.yml@refs/heads/main
IDENTITY_REGEXP='^https://github\.com/stuttgart-things/schmetterpause/\.github/workflows/ci\.yml@refs/'
OTHER_REGEXP='^https://github\.com/stuttgart-things/not-this-repository/'

# alpine:3.20's index, which carries no cosign signature.
UNSIGNED=alpine@sha256:beefdbd8a1da6d2915566fde36db9db0b524eb737fc57cd1367effd16dc0d06d

# EXPECTED VERSION IS READ FROM main.go SO THE PIN STAYS SINGLE-SOURCED.
want=$(sed -n 's/.*defaultCosignVersion *= *"\(.*\)".*/\1/p' "${MODULE}/main.go" | head -1)
if [ -z "${want}" ]; then
  echo "FAIL: could not read defaultCosignVersion from ${MODULE}/main.go"
  exit 1
fi
# No ".go: " in output: setup-go's problem matcher turns such a line into a
# failure annotation on a green job.
echo "cosign pin (read from ${MODULE}/main.go) is ${want}"

# Not a JWT, so cosign rejects it before it reaches Fulcio.
COSIGN_SMOKE_TOKEN="smoke-token-$(date +%s%N)-not-a-jwt"
export COSIGN_SMOKE_TOKEN

empty_predicate=$(mktemp)
predicate=$(mktemp)
trap 'rm -f "${empty_predicate}" "${predicate}"' EXIT
echo '{"bomFormat":"CycloneDX"}' > "${predicate}"

fail=0

# expect_pass <description> <needle> <function and args...>
expect_pass() {
  local desc=$1 needle=$2 out
  shift 2
  if ! out=$(dagger call -m "./${MODULE}" "$@" 2>&1); then
    echo "FAIL: ${desc} (call failed)"
    echo "${out}"
    fail=1
  elif ! grep -qF -- "${needle}" <<<"${out}"; then
    echo "FAIL: ${desc} (missing: ${needle})"
    echo "${out}"
    fail=1
  else
    echo "OK: ${desc}"
  fi
}

# expect_fail <description> <needle> <function and args...>
# The needle has to appear in the error, so a call that fails for an unrelated
# reason (a typo'd flag, a pull error) cannot pass as the refusal under test.
expect_fail() {
  local desc=$1 needle=$2 out
  shift 2
  if out=$(dagger call -m "./${MODULE}" "$@" 2>&1); then
    echo "FAIL: ${desc} (call succeeded)"
    echo "${out}"
    fail=1
  elif ! grep -qF -- "${needle}" <<<"${out}"; then
    echo "FAIL: ${desc} (failed, but without: ${needle})"
    echo "${out}"
    fail=1
  else
    echo "OK: ${desc}"
  fi
}

expect_pass "cosign defaults to the pinned ${want}" \
  "v${want}" \
  version
expect_pass "cosignVersion selects the release" \
  "v2.6.4" \
  version --cosign-version 2.6.4

# --- verify
expect_pass "verify: identity regexp" \
  "ci.yml@refs/heads/main" \
  verify --ref "${SIGNED}" --certificate-oidc-issuer "${ISSUER}" \
  --certificate-identity-regexp "${IDENTITY_REGEXP}"
expect_pass "verify: exact identity" \
  "ci.yml@refs/heads/main" \
  verify --ref "${SIGNED}" --certificate-oidc-issuer "${ISSUER}" \
  --certificate-identity "${IDENTITY}"
expect_fail "verify: a wrong identity is refused" \
  "none of the expected identities matched" \
  verify --ref "${SIGNED}" --certificate-oidc-issuer "${ISSUER}" \
  --certificate-identity-regexp "${OTHER_REGEXP}"
expect_fail "verify: no identity is not a default" \
  "pass certificateIdentity or certificateIdentityRegexp" \
  verify --ref "${SIGNED}" --certificate-oidc-issuer "${ISSUER}"
expect_fail "verify: identity and regexp together are refused" \
  "not both" \
  verify --ref "${SIGNED}" --certificate-oidc-issuer "${ISSUER}" \
  --certificate-identity "${IDENTITY}" --certificate-identity-regexp "${IDENTITY_REGEXP}"

# --- verify-refuses
expect_pass "verify-refuses: another repository is refused" \
  "refused, as it must" \
  verify-refuses --ref "${SIGNED}" --certificate-oidc-issuer "${ISSUER}" \
  --certificate-identity-regexp "${OTHER_REGEXP}"
expect_fail "verify-refuses: the real identity is not a refusal" \
  "is not a gate" \
  verify-refuses --ref "${SIGNED}" --certificate-oidc-issuer "${ISSUER}" \
  --certificate-identity-regexp "${IDENTITY_REGEXP}"
expect_fail "verify-refuses: no signature at all proves nothing" \
  "no signatures found" \
  verify-refuses --ref "${UNSIGNED}" --certificate-oidc-issuer "${ISSUER}" \
  --certificate-identity-regexp "${OTHER_REGEXP}"
expect_pass "verify-refuses: another repository's attestation is refused" \
  "refused, as it must" \
  verify-refuses --ref "${SIGNED}" --certificate-oidc-issuer "${ISSUER}" \
  --certificate-identity-regexp "${OTHER_REGEXP}" --predicate-type cyclonedx

# --- verify-attestation
expect_pass "verify-attestation: returns the CycloneDX predicate" \
  '"bomFormat":"CycloneDX"' \
  verify-attestation --ref "${SIGNED}" --predicate-type cyclonedx \
  --certificate-oidc-issuer "${ISSUER}" --certificate-identity-regexp "${IDENTITY_REGEXP}" \
  contents
expect_fail "verify-attestation: a type that is not attested fails" \
  "none of the attestations matched the predicate type" \
  verify-attestation --ref "${SIGNED}" --predicate-type spdxjson \
  --certificate-oidc-issuer "${ISSUER}" --certificate-identity-regexp "${IDENTITY_REGEXP}" \
  contents

# --- sign / attest refusals, before anything is signed
expect_fail "sign: a tag is refused" \
  "not pinned to a digest" \
  sign --ref alpine:3.20 --identity-token env:COSIGN_SMOKE_TOKEN
expect_fail "attest: a tag is refused" \
  "not pinned to a digest" \
  attest --ref alpine:3.20 --predicate "${predicate}" --predicate-type cyclonedx \
  --identity-token env:COSIGN_SMOKE_TOKEN
expect_fail "attest: an empty predicate is refused" \
  "is empty" \
  attest --ref "${UNSIGNED}" --predicate "${empty_predicate}" --predicate-type cyclonedx \
  --identity-token env:COSIGN_SMOKE_TOKEN

# --- the token reaches cosign, and nothing dagger prints carries it
if out=$(dagger --progress plain call -m "./${MODULE}" \
  sign --ref "${UNSIGNED}" --identity-token env:COSIGN_SMOKE_TOKEN 2>&1); then
  echo "FAIL: sign: a token that is not a JWT signed something"
  echo "${out}"
  fail=1
elif ! grep -qF "compact JWS format" <<<"${out}"; then
  echo "FAIL: sign: the token did not reach cosign as SIGSTORE_ID_TOKEN"
  echo "${out}"
  fail=1
elif grep -qF -- "${COSIGN_SMOKE_TOKEN}" <<<"${out}"; then
  echo "FAIL: sign: the identity token appears in --progress plain output"
  fail=1
else
  echo "OK: sign: the token reaches cosign and stays out of the output"
fi

exit "${fail}"
