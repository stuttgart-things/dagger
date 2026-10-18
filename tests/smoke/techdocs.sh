#!/usr/bin/env bash
# Smoke test for the techdocs module.
#
# `dagger functions` and `dagger develop` pass whether or not the container can
# actually build docs: node is downloaded and checked against SHASUMS256.txt,
# and mkdocs-techdocs-core and @techdocs/cli are installed at pinned versions,
# all at runtime. This builds the fixture under tests/techdocs and reads its
# entity, which needs neither secrets nor a bucket. publish is not covered
# here: it needs S3 credentials.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

MODULE=techdocs
FIXTURE=tests/techdocs

fail=0

# expect <description> <needle> <function and args...>
expect() {
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

# The entity Backstage would read the docs under: the Component carrying
# techdocs-ref, not the System before it.
expect "entity from a multi-document catalog-info.yaml" \
  "default/component/techdocs-smoke" \
  entity --src "${FIXTURE}"

# techdocs_metadata.json is what Backstage reads first; without it the page
# stays empty even if the HTML is there.
expect "generate writes techdocs_metadata.json" \
  '"site_name": "techdocs-smoke"' \
  generate --src "${FIXTURE}" file --path techdocs_metadata.json contents

expect "generate renders the page" \
  "techdocs smoke test" \
  generate --src "${FIXTURE}" file --path index.html contents

exit "${fail}"
