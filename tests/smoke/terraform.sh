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

# TARGETS: ONLY THE NAMED RESOURCE ENTERS THE STATE
TARGET_FOLDER=${OUTPUT_STATE_FOLDER}-targets
rm -rf "${TARGET_FOLDER}"
dagger call -m "${MODULE}" execute \
  --terraform-dir "${TEST_TERRAFORM_CODE}" \
  --operation apply \
  --targets null_resource.example \
  --refuse-destroy \
  --progress plain \
  export --path="${TARGET_FOLDER}"
if ! jq -e '[.resources[].name] == ["example"]' "${TARGET_FOLDER}/terraform.tfstate" >/dev/null; then
  echo "FAIL: --targets null_resource.example left more than that in the state"
  exit 1
fi
if ls "${TARGET_FOLDER}"/tfplan* >/dev/null 2>&1; then
  echo "FAIL: --refuse-destroy left its plan files (they hold values) in the output"
  exit 1
fi
echo "OK: targets"

# REFUSE-DESTROY: A CHANGED name REPLACES null_resource.example, SO THIS MUST FAIL
if dagger call -m "${MODULE}" execute \
  --terraform-dir "${TARGET_FOLDER}" \
  --operation apply \
  --variables "name=somebody-else" \
  --targets null_resource.example \
  --refuse-destroy \
  --progress plain \
  export --path="${TARGET_FOLDER}-refused" 2>&1 | tee /tmp/refuse.log; then
  echo "FAIL: --refuse-destroy applied a plan that replaces null_resource.example"
  exit 1
fi
grep -q "REFUSING TO APPLY" /tmp/refuse.log || { echo "FAIL: refused for the wrong reason"; exit 1; }
echo "OK: refuse-destroy"

# BIND-SERVICE: A NAME NO DNS KNOWS, REACHED THROUGH THE CALLER'S HOST SERVICE
BIND_IP=$(getent ahostsv4 example.com | awk 'NR==1 {print $1}')
status=$(dagger call -m "${MODULE}" execute \
  --terraform-dir tests/terraform-bind \
  --operation apply \
  --bind-service "tcp://${BIND_IP}:80" \
  --bind-service-alias pinned.dagger-smoke.test \
  --export-tf-output \
  --progress plain \
  file --path output.json contents | jq -r .status_code.value)
if [ -z "${status}" ] || [ "${status}" = "null" ]; then
  echo "FAIL: pinned.dagger-smoke.test was not reachable through --bind-service"
  exit 1
fi
echo "OK: bind-service (HTTP ${status} from ${BIND_IP})"
