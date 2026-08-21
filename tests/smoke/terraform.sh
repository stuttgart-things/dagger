#!/usr/bin/env bash
# Smoke test for the terraform module. Runs the fixture under tests/terraform
# through the full init/apply/output/destroy lifecycle.
#
# Kept as a script rather than inline Taskfile cmds so CI can run it without
# Task, whose remote `includes:` needs an interactive trust prompt.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

MODULE=terraform
TEST_TERRAFORM_CODE=tests/terraform
OUTPUT_STATE_FOLDER=${OUTPUT_STATE_FOLDER:-/tmp/dagger/terraform}

# ASSERT THE CONTAINER SHIPS THE PINNED TERRAFORM. THE EXPECTED VERSION IS READ
# FROM terraformVersion SO THE PIN STAYS SINGLE-SOURCED.
expected=$(sed -n 's/.*const terraformVersion = "\(.*\)".*/\1/p' "${MODULE}/container.go")
if [ -z "${expected}" ]; then
  echo "FAIL: could not read terraformVersion from ${MODULE}/container.go"
  exit 1
fi

actual=$(dagger call -m "${MODULE}" version)
echo "${actual}"
if ! echo "${actual}" | grep -q "^Terraform v${expected}$"; then
  echo "FAIL: expected Terraform v${expected}"
  exit 1
fi
echo "OK: terraform ${expected}"

# APPLY
dagger call -m "${MODULE}" execute \
  --terraform-dir "${TEST_TERRAFORM_CODE}" \
  --operation apply \
  -vv --progress plain \
  export --path="${OUTPUT_STATE_FOLDER}"

# OUTPUT
dagger call -m "${MODULE}" output \
  --terraform-dir "${TEST_TERRAFORM_CODE}" \
  -vv --progress plain

# DESTROY
dagger call -m "${MODULE}" execute --operation destroy \
  --terraform-dir "${OUTPUT_STATE_FOLDER}" \
  -vv --progress plain
