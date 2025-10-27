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

### Check Version

```bash
# Get Kyverno CLI version
dagger call -m kyverno version \
  --progress plain
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
