# Helm Dagger Module

This module provides Dagger functions for Helm chart operations including linting, packaging, rendering, and deployment.

## Features

- ✅ Helm chart linting and validation
- ✅ Chart packaging and registry publishing
- ✅ Template rendering with custom values
- ✅ Helmfile operations with Kubernetes integration
- ✅ Polaris-based security validation
- ✅ Kubeconform schema validation (with Datree CRDs-catalog fallback)
- ✅ Conftest / OPA policy evaluation (scaffold; policies caller-supplied)
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

### INSTALL CHART

```bash
dagger call -m helm execute \
--src "./" \
--release-name nginx \
--chart-path "oci://registry-1.docker.io/bitnamicharts/nginx" \
--operation install \
--namespace default \
--values "service.type=ClusterIP" \
--kube-config file://~/.kube/demo-infra
```

```bash
dagger call -m helm execute \
--src "./" \
--release-name nginx \
--operation uninstall \
--namespace default \
--kube-config file://~/.kube/demo-infra
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
# Apply Helmfile with local helmfile.yaml
dagger call -m helm helmfile-operation \
  --src tests/helm/ \
  --helmfile-ref helmfile.yaml \
  --kube-config file://~/.kube/config \
  --progress plain

# Apply from Git with state values
dagger call -m helm helmfile-operation \
  --helmfile-ref "git::https://github.com/stuttgart-things/helm.git@apps/nginx.yaml.gotmpl" \
  --operation apply \
  --state-values "replicas=3,namespace=webserver" \
  --kube-config file://~/.kube/kind-dev2 \
  --progress plain

# Destroy releases (use same state-values as apply)
dagger call -m helm helmfile-operation \
  --helmfile-ref "git::https://github.com/stuttgart-things/helm.git@apps/nginx.yaml.gotmpl" \
  --operation destroy \
  --state-values "namespace=webserver" \
  --kube-config file://~/.kube/kind-dev2 \
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

### Schema Validation (Kubeconform)

Render the chart and validate every resource against Kubernetes + CRD schemas.
Defaults to upstream Kubernetes schemas plus the Datree CRDs-catalog fallback,
which covers ArgoCD, Gateway-API, cert-manager, and cilium CRDs.

```bash
# Default schema locations (k8s + Datree CRDs-catalog)
dagger call -m helm kubeconform \
  --src tests/helm/test-chart

# With a values fixture
dagger call -m helm kubeconform \
  --src tests/helm/test-chart \
  --values-file tests/helm/test-values.yaml

# Override schema locations (e.g. internal CRD mirror)
dagger call -m helm kubeconform \
  --src tests/helm/test-chart \
  --schema-locations default \
  --schema-locations 'https://crds.internal/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json'
```

### Policy Evaluation (Conftest)

Render the chart and run `conftest test` against a caller-supplied Rego policy
directory. Policy authoring is deferred; the function is scaffolded so consumers
can wire the call now and add policies later without a signature change.

```bash
dagger call -m helm conftest \
  --src tests/helm/test-chart \
  --policy-dir ./policy \
  --values-file tests/helm/test-values.yaml
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
