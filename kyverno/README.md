# Kyverno Dagger Module

This module provides Dagger functions for Kyverno policy validation and Kubernetes security policy management.

## Features

- ✅ Policy validation against Kubernetes resources
- ✅ Kyverno CLI version checking
- ✅ Resource compliance testing
- ✅ Multi-policy validation support
- ✅ Detailed validation reporting
- ✅ GitOps workflow integration

## Prerequisites

- Dagger CLI installed
- Docker runtime available
- Kyverno policy files
- Kubernetes resource manifests

## Quick Start

### Validate Resources

```bash
# Validate resources against policies
dagger call -m kyverno validate \
  --policy tests/kyverno/policies/ \
  --resource tests/kyverno/resource-good/ \
  --progress plain
```

### Test Policies

`validate` answers "do these resources pass these policies?". `test` answers
the question a policy change raises: does the policy still refuse what it must
refuse, and still admit what it must admit? It runs `kyverno test` over every
`kyverno-test.yaml` below `--path` and fails on any result that does not match
its expectation — and on a path without a single test.

```bash
dagger call -m kyverno test \
  --src tests/kyverno/test \
  --progress plain
```

`tests/kyverno/test/kyverno-test.yaml` is a minimal example covering both a
refusal and an admission.

Add `--warnings-as-errors` (CLI 1.19 or later) to fail on deprecations, such as
the warning every `kyverno.io/v1` ClusterPolicy draws from 1.19 on.

Policies that verify image signatures reach the registry and Rekor during the
test, so a signed fixture image has to stay in its registry. `test` is never
served from Dagger's cache for that reason.

### Check Version

```bash
# Get Kyverno CLI version
dagger call -m kyverno version \
  --progress plain
```

## CLI Version

Every function takes `--kyverno-version` (default `1.19.1`). Match it to the
Kyverno running on the cluster: 1.19 refuses policies earlier releases accepted.
The CLI is copied out of `ghcr.io/kyverno/kyverno-cli:v<version>` into
`--base-image`, so any release can be pinned and the result does not depend on
the day the container is built.

```bash
dagger call -m kyverno test \
  --src ./policy-tests \
  --kyverno-version 1.18.1
```

### Test Module

```bash
# Run comprehensive tests
task test-kyverno
```

## API Reference

### Policy Validation

```bash
# Validate single resource against policies
dagger call -m kyverno validate \
  --policy ./policies/ \
  --resource ./resources/ \
  --progress plain
```

### Version Information

```bash
dagger call -m kyverno version \
  --progress plain
```

## Policy Structure Example

**policies/require-labels.yaml:**
```yaml
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: require-labels
spec:
  validationFailureAction: enforce
  background: false
  rules:
  - name: check-team-label
    match:
      any:
      - resources:
          kinds:
          - Pod
    validate:
      message: "label 'team' is required"
      pattern:
        metadata:
          labels:
            team: "?*"
```

**resources/pod.yaml:**
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  labels:
    team: "backend"
spec:
  containers:
  - name: nginx
    image: nginx:1.21
```

## Validation Results

The validation provides detailed feedback:
- ✅ **Pass**: Resource complies with all policies
- ❌ **Fail**: Resource violates one or more policies
- ⚠️ **Warning**: Resource has potential issues

## Examples

See the [main README](../README.md#kyverno) for detailed usage examples.

## Testing

```bash
task test-kyverno
```

## Resources

- [Kyverno Documentation](https://kyverno.io/docs/)
- [Kyverno Policies](https://kyverno.io/policies/)
- [Policy Examples](https://github.com/kyverno/policies)
