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
# FROM defaultTerraformVersion SO THE PIN STAYS SINGLE-SOURCED (Renovate keeps
# it and New's +default in step).
expected=$(sed -n 's/.*const defaultTerraformVersion = "\(.*\)".*/\1/p' "${MODULE}/container.go")
if [ -z "${expected}" ]; then
  echo "FAIL: could not read defaultTerraformVersion from ${MODULE}/container.go"
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

# RESOLV-CONF: THE MOUNTED FILE REPLACES THE ENGINE'S RESOLVER. A resolv.conf
# POINTING NOWHERE MUST BREAK RESOLUTION (PROVES IT IS USED), A PUBLIC ONE
# MUST RESOLVE example.com AGAIN.
printf 'nameserver 127.0.0.1\noptions timeout:1 attempts:1\n' > /tmp/resolv-nowhere.conf
printf 'nameserver 1.1.1.1\nnameserver 8.8.8.8\n' > /tmp/resolv-public.conf
if dagger call -m "${MODULE}" execute \
  --terraform-dir tests/terraform-bind \
  --operation apply \
  --variables "url=http://example.com" \
  --resolv-conf /tmp/resolv-nowhere.conf \
  --progress plain >/tmp/resolv-nowhere.log 2>&1; then
  echo "FAIL: example.com resolved with a resolv.conf that points nowhere"
  exit 1
fi
status=$(dagger call -m "${MODULE}" execute \
  --terraform-dir tests/terraform-bind \
  --operation apply \
  --variables "url=http://example.com" \
  --resolv-conf /tmp/resolv-public.conf \
  --export-tf-output \
  --progress plain \
  file --path output.json contents | jq -r .status_code.value)
if [ -z "${status}" ] || [ "${status}" = "null" ]; then
  echo "FAIL: example.com not reachable with a public --resolv-conf"
  exit 1
fi
echo "OK: resolv-conf (HTTP ${status})"

# RESOLV-CONF AND BIND-SERVICE ARE EXCLUSIVE
if dagger call -m "${MODULE}" execute \
  --terraform-dir tests/terraform-bind \
  --resolv-conf /tmp/resolv-public.conf \
  --bind-service "tcp://${BIND_IP}:80" \
  --bind-service-alias pinned.dagger-smoke.test \
  --progress plain >/tmp/resolv-bind.log 2>&1; then
  echo "FAIL: --resolv-conf together with --bind-service was accepted"
  exit 1
fi
grep -q "resolvConf and bindService are exclusive" /tmp/resolv-bind.log || { echo "FAIL: rejected for the wrong reason"; exit 1; }
echo "OK: resolv-conf/bind-service exclusive"

# EXTRA-FILES: A FILE FROM OUTSIDE terraform-dir IS READ AS ${path.module}/<name>,
# IS NOT HANDED BACK IN THE RETURNED DIRECTORY, AND A NAME CLASH IS AN ERROR
EXTRA_FOLDER=${OUTPUT_STATE_FOLDER}-extra
rm -rf "${EXTRA_FOLDER}"
dagger call -m "${MODULE}" execute \
  --terraform-dir tests/terraform-extra-files/code \
  --operation apply \
  --refuse-destroy \
  --extra-files tests/terraform-extra-files/extra/extra.txt \
  --export-tf-output \
  --progress plain \
  export --path="${EXTRA_FOLDER}"
value=$(jq -r .extra.value "${EXTRA_FOLDER}/output.json")
if [ "${value}" != "hello from an extra file" ]; then
  echo "FAIL: --extra-files content not readable at \${path.module}/extra.txt (got '${value}')"
  exit 1
fi
if [ -e "${EXTRA_FOLDER}/extra.txt" ]; then
  echo "FAIL: --extra-files leaked extra.txt into the returned directory"
  exit 1
fi
out=$(dagger call -m "${MODULE}" output \
  --terraform-dir "${EXTRA_FOLDER}" \
  --extra-files tests/terraform-extra-files/extra/extra.txt \
  --progress plain)
echo "${out}" | jq -e '.extra.value == "hello from an extra file"' >/dev/null \
  || { echo "FAIL: output with --extra-files"; exit 1; }
if dagger call -m "${MODULE}" execute \
  --terraform-dir tests/terraform-extra-files/code \
  --operation init \
  --extra-files tests/terraform-extra-files/extra/main.tf \
  --progress plain >/tmp/extra-clash.log 2>&1; then
  echo "FAIL: --extra-files overwrote main.tf in terraform-dir"
  exit 1
fi
grep -q 'already exists in terraformDir' /tmp/extra-clash.log || { echo "FAIL: rejected for the wrong reason"; exit 1; }
echo "OK: extra-files"
