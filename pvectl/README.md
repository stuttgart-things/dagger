# pvectl Dagger Module

Runs [pvectl](https://github.com/stuttgart-things/pvectl) from its container
image (`ghcr.io/stuttgart-things/pvectl`) against a Proxmox VE cluster, for the
LabUL template work in pipelines: list templates, check a config, follow
tasks, promote a fresh Packer template. Every function returns the JSON that
pvectl prints with `-o json`.

## Rules

- **VM descriptions (notes) are never returned.** In LabUL they hold
  plaintext passwords. pvectl drops them at the source.
- **TLS is always verified.** Pass the cluster CA (`--ca-cert`) or the pinned
  SHA-256 of the API certificate (`--fingerprint`). There is no insecure
  switch.
- **`template-promote` writes nothing without `--apply`.** By default it
  returns the plan (`--dry-run`).
- A non-zero pvectl exit fails the call with the exit code and the JSON
  pvectl printed (see [exit codes](https://github.com/stuttgart-things/pvectl#exit-codes)).

## Module Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `url` | string | required | `https://<host>:8006/api2/json` |
| `token-id` | Secret | | API token `user@realm!tokenname` (preferred) |
| `token-secret` | Secret | | its secret |
| `user` | Secret | | `user@realm`, instead of a token |
| `password` | Secret | | instead of a token |
| `ca-cert` | Secret | | cluster CA (`/etc/pve/pve-root-ca.pem`), PEM; survives certificate renewals |
| `fingerprint` | string | | SHA-256 of the API certificate (`AB:CD:…`, as in the web UI under node > System > Certificates) |
| `image` | string | `ghcr.io/stuttgart-things/pvectl:v0.6.1` | pvectl image; API tokens need v0.5.0 or later |

Credentials: a token (`token-id` + `token-secret`) or `user` + `password`.
`tls` needs neither.

## Functions

| Function | Parameters | Writes? |
|----------|------------|---------|
| `vms` | `--filter` (name substring or VMID), `--templates`, `--all` | no |
| `config` | `--vmid`; fails (exit 4) when a check fails | no |
| `tasks` | `--node`, `--vmid`, `--errors`, `--since 96h` | no |
| `tls` | the API certificate and its trust state: `ca`, `pinned` or `untrusted` | no |
| `template-promote` | `--vmid`, `--to` (default `ul-pve11`), `--fs-copy` (default true), `--allow-old`, `--repo` (stuttgart-things checkout, read only), `--apply` | only with `--apply` |

## Usage

```bash
E=~/.secrets.d/labul/proxmox.env
FP=33:91:DF:48:...:9D:65:57:13   # compare with the web UI

# certificate check (no login)
sops exec-env $E "dagger call -m github.com/stuttgart-things/dagger/pvectl \
  --url \$PVE_API_URL --fingerprint $FP tls"

# templates
sops exec-env $E "dagger call -m github.com/stuttgart-things/dagger/pvectl \
  --url \$PVE_API_URL --fingerprint $FP \
  --token-id env:PVE_API_TOKEN_ID --token-secret env:PVE_API_TOKEN_SECRET \
  vms --templates"

# promote plan, with the current repins from a checkout
sops exec-env $E "dagger call -m github.com/stuttgart-things/dagger/pvectl \
  --url \$PVE_API_URL --fingerprint $FP \
  --token-id env:PVE_API_TOKEN_ID --token-secret env:PVE_API_TOKEN_SECRET \
  template-promote --vmid 290 --repo ~/projects/stuttgart-things" | jq .plan

# carry it out (after review)
sops exec-env $E "dagger call ... template-promote --vmid 290 --apply" | jq .plan.repins
```

With the cluster CA instead of a fingerprint: `--ca-cert file:./pve-root-ca.pem`
(or `env:PVE_CA_CERT`).
