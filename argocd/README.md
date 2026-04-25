# ArgoCD Dagger Module

Register Kubernetes clusters in ArgoCD from a Dagger pipeline.

| Function          | Purpose                                                                            |
| ----------------- | ---------------------------------------------------------------------------------- |
| `base-container`  | Wolfi container with `curl`, `git`, `kubectl` and the `argocd` CLI.                |
| `add-cluster`     | `argocd login` + `argocd cluster add` with optional label assignment.              |
| `verify-catalog`  | Catalog-wide gate: schema parse + Chart.yaml uniqueness + per-chart helm-lint/kubeconform + argocd#41 double-render. |

## base-container

Returns a ready-to-use container. The argocd CLI is downloaded from the upstream
GitHub release (override with `--argocd-download-url` to pin a version).

```bash
# DROP INTO A SHELL WITH argocd / kubectl AVAILABLE
dagger call -m github.com/stuttgart-things/dagger/argocd base-container terminal
```

```bash
# PIN argocd TO A SPECIFIC VERSION
dagger call -m github.com/stuttgart-things/dagger/argocd base-container \
  --argocd-download-url https://github.com/argoproj/argo-cd/releases/download/v2.14.0/argocd-linux-amd64 \
  terminal
```

## add-cluster

Logs in to ArgoCD and runs `argocd cluster add <cluster-name> --yes`. The
kubeconfig context is renamed to `cluster-name` first, so k3s/k3d kubeconfigs
(whose context is always `default`) register under a meaningful name.

Flow inside the container:

1. Mount the kubeconfig as a secret and copy it to a writable path.
2. If `--source-context` (default `default`) exists and differs from
   `--cluster-name`, run `kubectl config rename-context`.
3. `argocd login --grpc-web` (with `--plaintext` or `--insecure` per flags).
4. `argocd cluster add <cluster-name> --yes`.
5. If `--labels` were supplied, `argocd cluster set <cluster-name> --label k=v ...`.

### TLS flag precedence

| Scenario                                   | Flags                               |
| ------------------------------------------ | ----------------------------------- |
| ArgoCD served over plain HTTP              | `--plaintext=true` *(default)*      |
| Self-signed / skip verification            | `--insecure=true` *(default)*       |
| Public CA-signed cert                      | `--plaintext=false --insecure=false`|

### Examples

```bash
# K3S / K3D (kubeconfig context is "default"), plaintext ArgoCD
dagger call -m github.com/stuttgart-things/dagger/argocd add-cluster \
  --kube-config file://~/.kube/tpl-testvm \
  --argocd-server argocd.platform.example.com \
  --username admin \
  --password env:ARGOCD_PASSWORD \
  --cluster-name tpl-testvm \
  -vv --progress plain
```

```bash
# REGISTER AND LABEL IN ONE CALL
dagger call -m github.com/stuttgart-things/dagger/argocd add-cluster \
  --kube-config file://~/.kube/tpl-testvm \
  --argocd-server argocd.platform.example.com \
  --username admin \
  --password env:ARGOCD_PASSWORD \
  --cluster-name tpl-testvm \
  --labels auto-project=true,env=dev
```

```bash
# KUBECONFIG CONTEXT ALREADY MATCHES THE CLUSTER NAME — SKIP THE RENAME
dagger call -m github.com/stuttgart-things/dagger/argocd add-cluster \
  --kube-config file://~/.kube/cicd-mgmt-1 \
  --argocd-server argocd.platform.example.com \
  --username admin \
  --password env:ARGOCD_PASSWORD \
  --cluster-name cicd-mgmt-1 \
  --source-context "" \
  --plaintext=false --insecure=false
```

## verify-catalog

Catalog-wide verification for an ArgoCD app-of-apps tree. Runs four gates in order
and fails fast:

1. `linting.ValidateJsonSchema` over the whole tree — every `*.schema.json` parses.
2. Chart discovery + `Chart.yaml` `name:` uniqueness pre-flight (cheap fail-fast on
   duplicate chart names).
3. Per chart: `helm.Lint` followed by `helm.Kubeconform` — renders each chart and
   validates manifests against Kubernetes + CRD schemas. If
   `<chart>/ci/default-values.yaml` exists it is passed as `--values-file` so
   charts with required values still render.
4. argocd#41 regression guard — for each chart, render twice with different
   `destination.server` values and assert the rendered `argoproj.io` Application
   `metadata.name` set differs between the two renders. Charts that emit no
   Application resources are skipped. This is the direct regression guard for
   [stuttgart-things/argocd#41](https://github.com/stuttgart-things/argocd/issues/41).

Composed via `dagger install` from the `helm` and `linting` modules — no duplication.

```bash
# VERIFY AN EXTERNAL CATALOG
dagger call -m github.com/stuttgart-things/dagger/argocd verify-catalog \
  --src ../argocd
```

```bash
# CUSTOM SCHEMA LOCATIONS (e.g. a pinned CRDs-catalog revision)
dagger call -m github.com/stuttgart-things/dagger/argocd verify-catalog \
  --src ../argocd \
  --schema-locations default \
  --schema-locations https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json
```

### In-repo regression fixtures

`tests/argocd/verify-catalog-happy/` — unique chart names, schemas parse, manifests
render clean. Exercised by `task test-argocd`.

`tests/argocd/verify-catalog-dup/` — two charts declaring `name: collision`. The
task negates the call (`! dagger ...`) so a non-zero exit is the passing signal.

### Testing local changes

```bash
dagger call -m ./argocd add-cluster --kube-config file://~/.kube/tpl-testvm ...
```

Use `-m ./argocd` to exercise uncommitted code; `-m github.com/stuttgart-things/dagger/argocd@<ref>` pins to a published tag.
