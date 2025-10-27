# Helm Dagger Module

This module provides Dagger functions for Helm chart operations including linting, packaging, rendering, and deployment.

## Features

- ✅ Helm chart linting and validation
- ✅ Chart packaging and registry publishing
- ✅ Template rendering with custom values
- ✅ Helmfile operations with Kubernetes integration
- ✅ Polaris-based security validation
- ✅ Registry authentication and multi-registry support

## Prerequisites

- Dagger CLI installed
- Docker runtime available
- Kubernetes cluster access (for deployments)

## Quick Start

### Lint Chart

```bash
# Lint Helm chart
dagger call -m helm lint \
  --src tests/helm/test-chart \
  -vv --progress plain
```

### Package Chart

```bash
# Package chart as TGZ
dagger call -m helm package \
  --src tests/helm/test-chart \
  -vv --progress plain \
  export --path=/tmp/chart.tgz
```

### Push to Registry

```bash
# Push chart to OCI registry
dagger call -m helm push \
  --src tests/helm/test-chart \
  --registry ghcr.io \
  --repository stuttgart-things \
  --username patrick-hermann-sva \
  --password env:GITHUB_TOKEN \
  -vv --progress plain
```

### Test Module

```bash
# Run comprehensive tests
task test-helm
```

## API Reference

### Chart Linting

```bash
dagger call -m helm lint \
  --src tests/helm/test-chart \
  --progress plain
```

### Template Rendering

```bash
dagger call -m helm render \
  --src tests/helm/test-chart \
  --values-file tests/helm/test-values.yaml \
  --progress plain
```

### Chart Packaging

```bash
dagger call -m helm package \
  --src tests/helm/test-chart \
  export --path=/tmp/chart.tgz
```

### Registry Publishing

```bash
dagger call -m helm push \
  --src tests/helm/test-chart \
  --registry ghcr.io \
  --repository stuttgart-things \
  --username patrick-hermann-sva \
  --password env:GITHUB_TOKEN
```

### Helmfile Operations

```bash
# Apply Helmfile with kubeconfig
dagger call -m helm helmfile-operation \
  --src tests/helm/ \
  --kube-config file://~/.kube/config \
  --progress plain

# Destroy releases
dagger call -m helm helmfile-operation \
  --operation destroy \
  --src tests/helm/ \
  --kube-config file://~/.kube/config \
  --progress plain
```

### Vault Integration

```bash
# Use Vault for kubeconfig
dagger call -m helm helmfile-operation \
  --operation destroy \
  --src tests/helm/ \
  --vault-url env:VAULT_ADDR \
  --vault-secret-id env:VAULT_SECRET_ID \
  --vault-app-role-id env:VAULT_ROLE_ID \
  --secret-path-kubeconfig kubeconfigs/test/kubeconfig \
  --progress plain
```

### Security Validation

```bash
# Validate with Polaris
dagger call -m helm validate-chart \
  --severity danger \
  --src tests/helm/test-chart/ \
  export --path=/tmp/polaris.json
```

## Examples

See the [main README](../README.md#helm) for detailed usage examples.

## Testing

```bash
task test-helm
```

## Resources

- [Helm Documentation](https://helm.sh/docs/)
- [Helmfile](https://helmfile.readthedocs.io/)
- [Polaris](https://polaris.docs.fairwinds.com/)
