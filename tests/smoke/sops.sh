#!/usr/bin/env bash
# Smoke test for the sops module (#404). Runs against the real container with
# throwaway age keys generated here -- never real keys.
#
# Key handling: the private keys only live in a 0700 temp dir that is removed
# on exit, and reach dagger as file: secrets. Nothing secret is echoed; checks
# compare values inside the shell.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

MODULE=./sops
FIXTURE=tests/sops/secret.yaml

TMP=$(mktemp -d)
chmod 700 "${TMP}"
trap 'rm -rf "${TMP}"' EXIT

call() { dagger call -m "${MODULE}" "$@"; }

fail() { echo "FAIL: $*"; exit 1; }

# Assert that a top-level or nested YAML line is plaintext / encrypted.
expect_plain() { grep -qE "^$2$" "$1" || fail "$3: expected plaintext line '$2'"; }
expect_enc() { grep -qE "^$2 ENC\[AES256_GCM" "$1" || fail "$3: expected '$2' to be encrypted"; }

# --- throwaway keys ------------------------------------------------------
call generate-age-key export --path "${TMP}/key1.txt" >/dev/null
call generate-age-key export --path "${TMP}/key2.txt" >/dev/null
chmod 600 "${TMP}"/key*.txt
# generate-age-key used to be served from Dagger's cache: same key every call.
cmp -s "${TMP}/key1.txt" "${TMP}/key2.txt" && fail "generate-age-key returned the same key twice"
echo "OK: generate-age-key returns a fresh key per call"

PUB1=$(call age-public-key --age-key "file:${TMP}/key1.txt")
PUB2=$(call age-public-key --age-key "file:${TMP}/key2.txt")
[[ "${PUB1}" == age1* && "${PUB2}" == age1* ]] || fail "age-public-key did not return age1... keys"
# The derived key must match the one age-keygen wrote into the key file.
grep -q "# public key: ${PUB1}$" "${TMP}/key1.txt" || fail "age-public-key does not match the key file"
echo "OK: age-public-key"

# Public keys are not secret; dagger takes them as secrets only because
# --age-key is a Secret.
export PUB1 PUB2

# --- generate-sops-config ------------------------------------------------
call generate-sops-config --age-public-key "${PUB1}" \
  --encrypted-regex '^(data|stringData)$' \
  export --path "${TMP}/flux.sops.yaml" >/dev/null
grep -q "encrypted_regex: '^(data|stringData)\$'" "${TMP}/flux.sops.yaml" || fail "generate-sops-config: no encrypted_regex"
grep -q "path_regex: '.\*\\\\.yaml'" "${TMP}/flux.sops.yaml" || fail "generate-sops-config: no path_regex per extension"
echo "OK: generate-sops-config --encrypted-regex"

# --- repro from #404: only data/stringData encrypted ---------------------
check_flux_secret() {
  local f=$1 what=$2
  expect_plain "${f}" "apiVersion: v1" "${what}"
  expect_plain "${f}" "kind: Secret" "${what}"
  expect_plain "${f}" "    name: dummy" "${what}"
  expect_plain "${f}" "type: Opaque" "${what}"
  expect_enc "${f}" "    foo:" "${what}"
  echo "OK: ${what} keeps apiVersion/kind/metadata readable, stringData encrypted"
}

call encrypt --age-key env:PUB1 --plaintext-file "${FIXTURE}" \
  --sops-config "${TMP}/flux.sops.yaml" \
  export --path "${TMP}/via-config.enc.yaml" >/dev/null
check_flux_secret "${TMP}/via-config.enc.yaml" "encrypt --sops-config"

call encrypt --age-key env:PUB1 --plaintext-file "${FIXTURE}" \
  --encrypted-regex '^(data|stringData)$' \
  export --path "${TMP}/via-regex.enc.yaml" >/dev/null
check_flux_secret "${TMP}/via-regex.enc.yaml" "encrypt --encrypted-regex"

# --- decrypt --extract ---------------------------------------------------
got=$(call decrypt --age-key "file:${TMP}/key1.txt" \
  --encrypted-file "${TMP}/via-regex.enc.yaml" \
  --extract '["stringData"]["foo"]' contents)
[ "${got}" = "bar" ] || fail "decrypt --extract: unexpected value"
echo "OK: decrypt --extract"

# --- round trip: encrypt -> set -> decrypt -------------------------------
NEW_VALUE="rotated-$(date +%s)"
export NEW_VALUE
call set --encrypted-file "${TMP}/via-regex.enc.yaml" \
  --path '["stringData"]["foo"]' --value env:NEW_VALUE \
  --age-key "file:${TMP}/key1.txt" \
  export --path "${TMP}/set.enc.yaml" >/dev/null
expect_plain "${TMP}/set.enc.yaml" "kind: Secret" "set"
expect_enc "${TMP}/set.enc.yaml" "    foo:" "set"
grep -q "${NEW_VALUE}" "${TMP}/set.enc.yaml" && fail "set: the new value is in the file in plaintext"

call decrypt --age-key "file:${TMP}/key1.txt" --encrypted-file "${TMP}/set.enc.yaml" \
  export --path "${TMP}/set.dec.yaml" >/dev/null
grep -q "^    foo: ${NEW_VALUE}$" "${TMP}/set.dec.yaml" || fail "set: decrypted value is not the new one"
grep -q "^    other: unchanged$" "${TMP}/set.dec.yaml" || fail "set: other values changed"
echo "OK: encrypt -> set -> decrypt"

# --- update-keys: add a second recipient ---------------------------------
if call decrypt --age-key "file:${TMP}/key2.txt" --encrypted-file "${TMP}/via-config.enc.yaml" \
  contents >/dev/null 2>&1; then
  fail "update-keys: key2 can decrypt before it was added"
fi

call generate-sops-config --age-public-key "${PUB1},${PUB2}" \
  --encrypted-regex '^(data|stringData)$' \
  export --path "${TMP}/two.sops.yaml" >/dev/null

call update-keys --encrypted-file "${TMP}/via-config.enc.yaml" \
  --sops-config "${TMP}/two.sops.yaml" --age-key "file:${TMP}/key1.txt" \
  export --path "${TMP}/two.enc.yaml" >/dev/null
[ "$(grep -c 'recipient: age1' "${TMP}/two.enc.yaml")" -eq 2 ] || fail "update-keys: expected two recipients"

got=$(call decrypt --age-key "file:${TMP}/key2.txt" --encrypted-file "${TMP}/two.enc.yaml" \
  --extract '["stringData"]["foo"]' contents)
[ "${got}" = "bar" ] || fail "update-keys: key2 cannot decrypt after update"
echo "OK: update-keys adds a second recipient"

echo "ALL OK: sops"
