#!/usr/bin/env bash
# Smoke test for the ansible module. Runs the fixtures under tests/ansible
# against the real container.
#
# Kept as a script rather than inline Taskfile cmds so CI can run it without
# Task, whose remote `includes:` needs an interactive trust prompt.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

MODULE=ansible
TEST_COLLECTION=tests/ansible/collection
OUTPUT_COLLECTION_FOLDER=${OUTPUT_COLLECTION_FOLDER:-/tmp/dagger/ansible}

# ASSERT THE CONTAINER SHIPS THE PINNED ANSIBLE. THE EXPECTED VERSION IS READ
# FROM defaultAnsibleVersion SO THE PIN STAYS SINGLE-SOURCED.
expected=$(sed -n 's/.*defaultAnsibleVersion = "\(.*\)".*/\1/p' "${MODULE}/container.go")
if [ -z "${expected}" ]; then
  echo "FAIL: could not read defaultAnsibleVersion from ${MODULE}/container.go"
  exit 1
fi

actual=$(dagger call -m "${MODULE}" version)
echo "${actual}"
if ! echo "${actual}" | grep -q "^ansible package: ${expected}$"; then
  echo "FAIL: expected ansible package ${expected}"
  exit 1
fi
echo "OK: ansible ${expected}"

# EXECUTE PLAYBOOKS
dagger call -m "${MODULE}" execute \
  --src . \
  --playbooks tests/ansible/hello.yaml,tests/ansible/hello2.yaml \
  -vv --progress plain

# BUILD A COLLECTION
rm -rf "${OUTPUT_COLLECTION_FOLDER}" || true
dagger call -m "./${MODULE}" run-collection-build-pipeline \
  --src "${TEST_COLLECTION}" \
  --progress plain export \
  --path="${OUTPUT_COLLECTION_FOLDER}"
ls -lta "${OUTPUT_COLLECTION_FOLDER}"

if ! ls "${OUTPUT_COLLECTION_FOLDER}"/*.tar.gz >/dev/null 2>&1; then
  echo "FAIL: no collection artifact in ${OUTPUT_COLLECTION_FOLDER}"
  exit 1
fi
echo "OK: collection artifact built"
