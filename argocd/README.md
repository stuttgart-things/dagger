# ArgoCD Dagger Module

Register Kubernetes clusters in ArgoCD and render ArgoCD resources from OCI-hosted
KCL modules — without ever leaving a Dagger pipeline.

Three functions, three levels of coupling to the ArgoCD API:

| Function | How it reaches ArgoCD | Requires |
|---|---|---|
| `add-cluster-cli`    | `argocd login` + `argocd cluster add` (gRPC-Web) | ArgoCD username/password, reachable gRPC-Web ingress |
| `add-cluster-k-8-s`  | `kubectl apply` of a cluster Secret, no API call | kubeconfig of the ArgoCD-hosting cluster |
| `create-app-project` | Renders via shared KCL module → optional `kubectl apply` | kubeconfig of the ArgoCD-hosting cluster (only if applying) |

## add-cluster-cli

Logs in to ArgoCD and runs `argocd cluster add <context> --name <cluster-name>`
so the kubeconfig is registered untouched and appears in ArgoCD under the given
display name.

The source context is picked in this order:

1. `--source-context <name>` if supplied
2. `kubectl config current-context` of the kubeconfig
3. Fails with a list of available contexts if neither is set

TLS options (precedence: `--plaintext` > `--server-certs-dir` > `--server-cert` > `--insecure` > system trust store):

| Scenario                                          | Flags                                                        |
| ------------------------------------------------- | ------------------------------------------------------------ |
| ArgoCD server over plain HTTP (no TLS)            | `--plaintext=true`                                           |
| Public CA-signed cert (e.g. Let's Encrypt)        | `--insecure=false` (no extra flags)                          |
| Self-signed / private CA, single cert             | `--server-cert ./argocd-ca.pem`                              |
| Self-signed / private CA, whole directory         | `--server-certs-dir /usr/local/share/ca-certificates`        |
| Skip TLS verification (dev / trusted networks)    | `--insecure=true` *(default)*                                |

`--server-certs-dir` concatenates every `*.crt` and `*.pem` file in the directory
into a single PEM bundle inside the container and hands it to `argocd login --server-crt`.

Pin the CLI to the same major.minor as your server via `--cli-package` (default
`argo-cd-2.14`). The `argo-cd` meta-package tracks the latest major and can break
against older servers.

```bash
# SKIP TLS VERIFICATION (default)
dagger call -m github.com/stuttgart-things/dagger/argocd add-cluster-cli \
--kube-config file://~/.kube/ci-mgmt-1 \
--argocd-server argocd.example.com \
--username admin \
--password env:ARGOCD_PASSWORD \
--cluster-name cicd-mgmt-1 \
--insecure=true \
-vv --progress plain
```

```bash
# VERIFY TLS WITH A WHOLE DIRECTORY OF CA CERTS
dagger call -m github.com/stuttgart-things/dagger/argocd add-cluster-cli \
--kube-config file://~/.kube/ci-mgmt-1 \
--argocd-server argocd.example.com \
--username admin \
--password env:ARGOCD_PASSWORD \
--cluster-name cicd-mgmt-1 \
--server-certs-dir /usr/local/share/ca-certificates \
-vv --progress plain
```

```bash
# VERIFY TLS WITH A SINGLE CUSTOM CA
dagger call -m github.com/stuttgart-things/dagger/argocd add-cluster-cli \
--kube-config file://~/.kube/ci-mgmt-1 \
--argocd-server argocd.example.com \
--username admin \
--password env:ARGOCD_PASSWORD \
--cluster-name cicd-mgmt-1 \
--server-cert ./argocd-ca.pem \
-vv --progress plain
```

```bash
# PIN THE CLI TO v3.3 (when your server is v3.x)
dagger call -m github.com/stuttgart-things/dagger/argocd add-cluster-cli \
--kube-config file://~/.kube/ci-mgmt-1 \
--argocd-server argocd.example.com \
--username admin \
--password env:ARGOCD_PASSWORD \
--cluster-name cicd-mgmt-1 \
--cli-package argo-cd-3.3 \
-vv --progress plain
```

## add-cluster-k-8-s

Registers a Kubernetes cluster in ArgoCD **without calling the ArgoCD HTTP/gRPC API**.
Useful when the ArgoCD ingress blocks gRPC-Web, when you don't want to expose an
admin credential to the runner, or when you prefer a pure-Kubernetes path.

What it does, in order:

1. Reads the target kubeconfig, resolves context → cluster → `server` and
   `certificate-authority-data`.
2. Applies a `ServiceAccount`, `ClusterRole` (`*/*/*`) and `ClusterRoleBinding`
   (default `argocd-manager` in `kube-system`) in the **target** cluster.
3. Mints a token with `kubectl create token --duration=<token-duration>`.
4. Builds the ArgoCD cluster `Secret` (labelled
   `argocd.argoproj.io/secret-type=cluster`) and applies it in the **ArgoCD**
   cluster's `argocd` namespace.

`--argocd-kube-config` is optional — if omitted, the target kubeconfig is reused
(ArgoCD registers itself / is in the same cluster).

```bash
# REGISTER A REMOTE CLUSTER INTO ARGOCD
dagger call -m github.com/stuttgart-things/dagger/argocd add-cluster-k-8-s \
--kube-config file://~/.kube/ci-mgmt-1 \
--argocd-kube-config file://~/.kube/argocd-host \
--cluster-name cicd-mgmt-1 \
-vv --progress plain
```

```bash
# ARGOCD AND TARGET ARE THE SAME CLUSTER
dagger call -m github.com/stuttgart-things/dagger/argocd add-cluster-k-8-s \
--kube-config file://~/.kube/argocd-host \
--cluster-name cicd-mgmt-1 \
-vv --progress plain
```

```bash
# OVERRIDE SERVER URL (e.g. in-cluster DNS instead of the kubeconfig's external URL)
dagger call -m github.com/stuttgart-things/dagger/argocd add-cluster-k-8-s \
--kube-config file://~/.kube/ci-mgmt-1 \
--argocd-kube-config file://~/.kube/argocd-host \
--cluster-name cicd-mgmt-1 \
--server-url https://kubernetes.default.svc \
--token-duration 720h \
-vv --progress plain
```

Re-running the function re-mints the token and re-applies the `Secret`, which
is also how you'd rotate it before the previous one expires. Token duration is
capped by the target cluster's kube-apiserver
(`--service-account-max-token-expiration`).

## create-app-project

Renders an ArgoCD `AppProject` manifest from the
[`argocd-app-project`](https://github.com/stuttgart-things/stuttgart-things/tree/main/kubernetes/argocd-app-project)
KCL module (fetched from OCI via the shared `kcl` Dagger module) and optionally
applies it to the ArgoCD-hosting cluster.

Complex spec fields (`destinations`, whitelists, `labels`, `annotations`) take
JSON strings — wrap them in single quotes so the shell leaves the braces alone.

```bash
# RENDER ONLY (returns a Directory containing <name>.yaml)
dagger call -m github.com/stuttgart-things/dagger/argocd create-app-project \
--name xplane-test \
--destinations '[{"server":"https://10.100.136.192:34360","namespace":"*"}]' \
--progress plain \
export --path=/tmp/appproject
```

```bash
# RENDER + APPLY TO THE ARGOCD CLUSTER
dagger call -m github.com/stuttgart-things/dagger/argocd create-app-project \
--name xplane-test \
--destinations '[{"server":"https://10.100.136.192:34360","namespace":"*"}]' \
--apply-to-cluster=true \
--kube-config file://~/.kube/platform-sthings \
--progress plain
```

```bash
# LOCKED-DOWN PROJECT: one repo, one namespace, specific kinds
dagger call -m github.com/stuttgart-things/dagger/argocd create-app-project \
--name team-platform \
--description "Platform team project" \
--source-repos '["https://github.com/example/platform"]' \
--destinations '[{"server":"https://kubernetes.default.svc","namespace":"platform"}]' \
--cluster-resource-whitelist '[{"group":"","kind":"Namespace"}]' \
--namespace-resource-whitelist '[{"group":"apps","kind":"Deployment"},{"group":"","kind":"ConfigMap"}]' \
--labels '{"team":"platform"}' \
--apply-to-cluster=true \
--kube-config file://~/.kube/platform-sthings \
--progress plain
```

Pin the KCL module to a specific version with `--oci-source`:

```bash
--oci-source 'oci://ghcr.io/stuttgart-things/argocd-app-project?tag=0.1.0'
```

## Files

- [main.go](main.go) — module struct
- [cli.go](cli.go) — `AddClusterCli`
- [cluster.go](cluster.go) — `AddClusterK8s`
- [project.go](project.go) — `CreateAppProject`
