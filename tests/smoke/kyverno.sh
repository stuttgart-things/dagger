#!/usr/bin/env bash
# Smoke test for the kyverno module.
#
# WHY THIS EXISTS. The module's CI (this-test-modules.yaml) runs `dagger
# functions`, `dagger develop` and golangci-lint — all of which pass whether or
# not the CLI in the container is the pinned one, and whether or not Test can
# tell a policy that still refuses from one that admits everything. Both are
# runtime properties:
#
#   - the CLI used to be whatever Wolfi shipped on the day the container was
#     built, and 1.19 refuses policies that earlier releases accepted;
#   - a test run is only worth something if a wrong expectation fails it, and
#     if a directory without tests is not reported as a pass.
#
# Runs without secrets, a cluster or a registry login, which is what makes it
# CI-eligible. The fixtures under tests/kyverno/test* need no network.
#
# Kept as a script rather than inline Taskfile cmds so CI can run it without
# Task, whose remote `includes:` needs an interactive trust prompt.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

MODULE=kyverno
FIXTURES=tests/kyverno

# EXPECTED VERSION IS READ FROM main.go SO THE PIN STAYS SINGLE-SOURCED.
want=$(sed -n 's/.*defaultKyvernoVersion *= *"\(.*\)".*/\1/p' "${MODULE}/main.go" | head -1)
if [ -z "${want}" ]; then
  echo "FAIL: could not read defaultKyvernoVersion from ${MODULE}/main.go"
  exit 1
fi
echo "pin from main.go: kyverno CLI ${want}"

fail=0

# expect_pass <description> <needle> <function and args...>
# An empty needle only checks that the call succeeds, for functions such as
# validate that return nothing to look at.
expect_pass() {
  local desc=$1 needle=$2 out
  shift 2
  if ! out=$(dagger call -m "./${MODULE}" "$@" 2>&1); then
    echo "FAIL: ${desc} (call failed)"
    echo "${out}"
    fail=1
  elif [ -n "${needle}" ] && ! grep -qF -- "${needle}" <<<"${out}"; then
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

expect_pass "CLI defaults to the pinned ${want}" \
  "Version: v${want}" \
  version
expect_pass "kyvernoVersion selects the CLI" \
  "Version: v1.18.1" \
  version --kyverno-version 1.18.1

expect_pass "test: policy refuses latest-tag and admits pinned-tag" \
  "2 tests passed and 0 tests failed" \
  test --src "${FIXTURES}/test"
expect_fail "test: inverted expectations fail" \
  "2 tests failed" \
  test --src "${FIXTURES}/test-mismatch"
expect_pass "test: deprecated kind passes by default" \
  "2 tests passed and 0 tests failed" \
  test --src "${FIXTURES}/test-deprecated"
expect_fail "test: deprecated kind fails with --warnings-as-errors" \
  "deprecation warning" \
  test --src "${FIXTURES}/test-deprecated" --warnings-as-errors
expect_fail "test: a directory without tests is not a pass" \
  "no tests found" \
  test --src "${FIXTURES}/policies"

expect_pass "validate: good resource passes" \
  "" \
  validate --policy "${FIXTURES}/policies" --resource "${FIXTURES}/resource-good"
expect_fail "validate: bad resource fails" \
  "exit code: 1" \
  validate --policy "${FIXTURES}/policies" --resource "${FIXTURES}/resource-bad"

exit "${fail}"
