package main

import rego.v1

# Smoke-test fixture for the conftest function. This policy never denies —
# it exists only so `dagger call conftest` has a valid --policy-dir to point
# at during `task test-helm`. Replace with real rules in downstream repos.

deny contains msg if {
	false
	msg := "unreachable"
}
