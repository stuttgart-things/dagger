# Helm Dagger Module

This module provides Dagger functions for Helm chart operations including linting, packaging, rendering, and deployment.

## Features

- ✅ Helm chart linting and validation
- ✅ Chart packaging and registry publishing
- ✅ Template rendering with custom values
- ✅ Helmfile operations (apply / destroy / template / diff) with Kubernetes integration
- ✅ Polaris-based security validation (report-only)
- ✅ Kubeconform schema validation (with the community-maintained CRDs-catalog at `github.com/datreeio/CRDs-catalog` as fallback)
- ✅ Conftest / OPA policy evaluation (scaffold; policies caller-supplied)
- ✅ Registry authentication and multi-registry support
- ✅ Vault-backed kubeconfig retrieval via vals
- ✅ `helm test` post-install verification
- ✅ Bundled tool-version reporting (`version`)

## Bundled tool versions

Pinned in `helm/container.go`. Run `dagger call -m helm version` to report the live values from a built container.

| Tool        | Version |
|-------------|---------|
| helm        | v4.1.4  |
| helmfile    | 1.4.4   |
| polaris     | 10.1.8  |
| kubeconform | 0.7.0   |
| conftest    | 0.68.2  |
| vals        | 0.43.9  |
| kubectl     | 1.35.x (Wolfi `kubectl-1.35`) |

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

### Install Chart

```bash
dagger call -m helm execute \
--src "./" \
--release-name nginx \
--chart-path "oci://registry-1.docker.io/bitnamicharts/nginx" \
--operation install \
--namespace default \
--values "service.type=ClusterIP" \
--wait --timeout 5m --atomic \
--kube-config file://~/.kube/demo-infra
```

`--wait` blocks until resources are Ready, `--timeout` bounds the wait, and `--atomic` rolls back on failure. `--dry-run` is also available for render-and-validate-only.

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

### Render Helmfile

Renders all releases declared in a helmfile to YAML without contacting a cluster. Useful for diffs against a checked-in golden manifest.

```bash
# Render local helmfile
dagger call -m helm render-helmfile \
  --src tests/helm/ \
  --helmfile-name helmfile.yaml

# Render with state-values overrides
dagger call -m helm render-helmfile \
  --src tests/helm/ \
  --helmfile-name helmfile.yaml \
  --state-values "namespace=demo,replicas=3"
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

### Helmfile Diff

Read-only comparison of declared releases against the cluster. Wraps `helmfile diff`.

```bash
dagger call -m helm diff \
  --src tests/helm/ \
  --helmfile-ref helmfile.yaml \
  --kube-config file://~/.kube/config
```

### Helm Test

Post-install verification — runs `helm test <release>` against an existing release.

```bash
dagger call -m helm test \
  --release-name nginx \
  --namespace default \
  --kube-config file://~/.kube/config \
  --timeout 5m \
  --logs
```

### Bundled Tool Versions

Print the version of every binary in the container. No args.

```bash
dagger call -m helm version
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

Polaris runs in **report-only** mode — severity violations land in the returned JSON file but do not fail the call. Callers parse the report to enforce policy.

```bash
# Validate with Polaris
dagger call -m helm validate-chart \
  --severity danger \
  --src tests/helm/test-chart/ \
  export --path=/tmp/polaris.json
```

### Schema Validation (Kubeconform)

Render the chart and validate every resource against Kubernetes + CRD schemas.
Defaults to upstream Kubernetes schemas plus the community-maintained
[CRDs-catalog](https://github.com/datreeio/CRDs-catalog) at
`github.com/datreeio/CRDs-catalog`, which covers ArgoCD, Gateway-API,
cert-manager, and cilium CRDs.

```bash
# Default schema locations (k8s + CRDs-catalog)
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

> **Rego v1 required.** The bundled conftest 0.68.x defaults to Rego v1, which
> requires the `if` and `contains` keywords. Old-style policies fail to parse —
> add `import rego.v1` and rewrite rules to `deny contains msg if { ... }`.
> See `tests/helm/test-policy/smoke.rego` for a minimal example.

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
- [Kubeconform](https://github.com/yannh/kubeconform)
- [Conftest](https://www.conftest.dev/)
- [vals](https://github.com/helmfile/vals)
- [CRDs-catalog](https://github.com/datreeio/CRDs-catalog)
