# argocd test fixtures

YAML parameter files for the `argocd` Dagger module's `create-*` functions.
Pass them via `--parameters-file`; every CLI flag still wins over any value
set in the file.

| File | Target function | Reproduces |
|---|---|---|
| [appproject-xplane-test.yaml](appproject-xplane-test.yaml) | `create-app-project` | `xplane-test` AppProject pinned to a single cluster |
| [application-xplane-test-guestbook.yaml](application-xplane-test-guestbook.yaml) | `create-application` | `xplane-test-guestbook` Application (argocd-example-apps `guestbook`) |
| [applicationset-multi-cluster-guestbook.yaml](applicationset-multi-cluster-guestbook.yaml) | `create-application-set` | ApplicationSet with a list generator fanning guestbook across two clusters |

## Usage

```bash
# RENDER ONLY — all parameters come from the file
dagger call -m github.com/stuttgart-things/dagger/argocd create-app-project \
--parameters-file tests/argocd/appproject-xplane-test.yaml \
export --path=/tmp/appproject

# MIX FILE + CLI OVERRIDES — CLI wins
dagger call -m github.com/stuttgart-things/dagger/argocd create-application \
--parameters-file tests/argocd/application-xplane-test-guestbook.yaml \
--name xplane-test-guestbook-dev \
--dest-namespace guestbook-dev

# RENDER + APPLY
dagger call -m github.com/stuttgart-things/dagger/argocd create-application-set \
--parameters-file tests/argocd/applicationset-multi-cluster-guestbook.yaml \
--apply-to-cluster=true \
--kube-config file://~/.kube/argocd-host
```

## File format notes

- Scalars are plain YAML: `name: foo`, `namespace: argocd`.
- Arrays and objects are preserved — the KCL module's runner flattens them
  into JSON strings before handing them to `kcl run -D`.
- Multiline strings work; the runner escapes newlines.
- Go-template expressions inside quoted strings (`"{{ .cluster }}-guestbook"`)
  pass through untouched.
