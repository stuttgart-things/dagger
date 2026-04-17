# ArgoCD Dagger Module

Register Kubernetes clusters in ArgoCD and render ArgoCD resources from OCI-hosted
KCL modules — without ever leaving a Dagger pipeline.

Five functions, from direct ArgoCD-API calls to pure Kubernetes / KCL paths:

| Function | How it reaches ArgoCD | Requires |
|---|---|---|
| `add-cluster-cli`        | `argocd login` + `argocd cluster add` (gRPC-Web)        | ArgoCD username/password, reachable gRPC-Web ingress |
| `add-cluster-k-8-s`      | `kubectl apply` of a cluster Secret, no API call        | kubeconfig of the ArgoCD-hosting cluster |
| `create-app-project`     | Renders via shared KCL module → optional `kubectl apply` | kubeconfig of the ArgoCD-hosting cluster (only if applying) |
| `create-application`     | Renders via shared KCL module → optional `kubectl apply` | kubeconfig of the ArgoCD-hosting cluster (only if applying) |
| `create-application-set` | Renders via shared KCL module → optional `kubectl apply` | kubeconfig of the ArgoCD-hosting cluster (only if applying) |

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

The function returns a Directory containing `<cluster-name>.yaml` — the rendered
Secret. Pass `--apply-to-cluster=false` to skip the `kubectl apply` step and just
get the file back (the target cluster is still mutated — SA + RBAC + token — because
the Secret can't be built without a live token).

```bash
# RENDER ONLY — no ArgoCD cluster kubeconfig needed
dagger call -m github.com/stuttgart-things/dagger/argocd add-cluster-k-8-s \
--kube-config file://~/.kube/ci-mgmt-1 \
--cluster-name cicd-mgmt-1 \
--apply-to-cluster=false \
export --path /tmp/cluster-secret
```

## create-app-project

Renders an ArgoCD `AppProject` manifest from the
[`argocd-app-project`](https://github.com/stuttgart-things/kcl/tree/main/kubernetes/argocd-app-project)
KCL module (fetched from OCI via the shared `kcl` Dagger module) and optionally
applies it to the ArgoCD-hosting cluster.

Complex spec fields (`destinations`, whitelists, `labels`, `annotations`) take
JSON strings — wrap them in single quotes so the shell leaves the braces alone.

All three `create-*` functions also accept `--parameters-file <yaml>`, a YAML
(or JSON) file with the same keys as the CLI flags. CLI flags override
values from the file; values from the file override the KCL module's own
defaults. See [tests/argocd/](../../tests/argocd/) for working examples.

```bash
# FROM YAML ONLY
dagger call -m github.com/stuttgart-things/dagger/argocd create-app-project \
--parameters-file tests/argocd/appproject-xplane-test.yaml \
--apply-to-cluster=true --kube-config file://~/.kube/argocd-host

# YAML + CLI OVERRIDE
dagger call -m github.com/stuttgart-things/dagger/argocd create-application \
--parameters-file tests/argocd/application-xplane-test-guestbook.yaml \
--name xplane-test-guestbook-dev --dest-namespace guestbook-dev
```

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

## create-application

Renders an ArgoCD `Application` manifest from the
[`argocd-application`](https://github.com/stuttgart-things/kcl/tree/main/kubernetes/argocd-application)
KCL module and optionally applies it.

Scalar fields (`project`, `repoURL`, `path`, `chart`, `destServer`,
`destNamespace`, …) are exposed as typed params. Complex nested fields
(`helm`, `kustomize`, `sources`, `syncPolicy`, `retry`, `automated`,
`destination`, `info`, `labels`, `annotations`, `finalizers`) take JSON
strings — wrap them in single quotes.

```bash
# RECREATE THE xplane-test-guestbook APPLICATION
dagger call -m github.com/stuttgart-things/dagger/argocd create-application \
--name xplane-test-guestbook \
--project xplane-test \
--dest-server https://10.100.136.192:34360 \
--dest-namespace guestbook \
--progress plain \
export --path=/tmp/app.yaml
```

```bash
# HELM CHART FROM A HELM REPO
dagger call -m github.com/stuttgart-things/dagger/argocd create-application \
--name kube-prometheus-stack \
--project monitoring \
--repo-url https://prometheus-community.github.io/helm-charts \
--chart kube-prometheus-stack \
--target-revision 65.1.0 \
--dest-namespace monitoring \
--helm '{"releaseName":"kps","valueFiles":["values.yaml"]}' \
--apply-to-cluster=true \
--kube-config file://~/.kube/platform-sthings
```

```bash
# PIN A GIT REVISION, DISABLE AUTO-SYNC, ADD A FINALIZER
dagger call -m github.com/stuttgart-things/dagger/argocd create-application \
--name pinned-app \
--target-revision v1.2.3 \
--sync-policy '{}' \
--finalizers '["resources-finalizer.argocd.argoproj.io"]'
```

## create-application-set

Renders an ArgoCD `ApplicationSet` manifest from the
[`argocd-application-set`](https://github.com/stuttgart-things/kcl/tree/main/kubernetes/argocd-application-set)
KCL module and optionally applies it.

`generators` is the main lever. Argo Go-template expressions like
`{{ .cluster.name }}` survive the JSON/shell round-trip when they sit inside
quoted string values.

```bash
# FAN-OUT A SINGLE-SOURCE APP ACROSS A LIST OF CLUSTERS
dagger call -m github.com/stuttgart-things/dagger/argocd create-application-set \
--name multi-cluster-guestbook \
--generators '[{"list":{"elements":[
    {"cluster":"dev","url":"https://1.2.3.4","namespace":"guestbook-dev"},
    {"cluster":"prod","url":"https://5.6.7.8","namespace":"guestbook-prod"}
  ]}}]' \
--template-name '{{ .cluster }}-guestbook' \
--template-labels '{"cluster":"{{ .cluster }}"}' \
--source '{"repoURL":"https://github.com/argoproj/argocd-example-apps.git","path":"guestbook","targetRevision":"HEAD"}' \
--dest-server '{{ .url }}' \
--dest-namespace '{{ .namespace }}' \
--sync-options '["CreateNamespace=true"]'
```

```bash
# CLUSTER GENERATOR (ALL ARGO-REGISTERED CLUSTERS)
dagger call -m github.com/stuttgart-things/dagger/argocd create-application-set \
--name cluster-addons \
--generators '[{"clusters":{}}]' \
--template-name 'addons-{{ .name }}' \
--dest-server '{{ .server }}' \
--dest-namespace addons
```

```bash
# ROLLING SYNC STRATEGY
dagger call -m github.com/stuttgart-things/dagger/argocd create-application-set \
--name staged-rollout \
--strategy '{"type":"RollingSync","rollingSync":{"steps":[
    {"matchExpressions":[{"key":"env","operator":"In","values":["dev"]}]},
    {"matchExpressions":[{"key":"env","operator":"In","values":["prod"]}],"maxUpdate":"10%"}
  ]}}'
```

Pin any of the three KCL modules with `--oci-source 'oci://.../<module>?tag=<version>'`.

## Files

- [main.go](main.go) — module struct
- [cli.go](cli.go) — `AddClusterCli`
- [cluster.go](cluster.go) — `AddClusterK8s`
- [project.go](project.go) — `CreateAppProject`
- [application.go](application.go) — `CreateApplication`
- [applicationset.go](applicationset.go) — `CreateApplicationSet`
