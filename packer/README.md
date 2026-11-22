# Packer Dagger Module

This module provides Dagger functions for Packer operations including image building, vCenter and Proxmox operations, and Vault integration.

## Features

- ✅ Packer image building with HCL configurations
- ✅ vCenter VM template operations (move, rename, delete) + datastore/network inspection
- ✅ Proxmox VE operations (move/migrate, rename, delete, lookup VMID by name) + storage/network/resource inspection via the [pvectl](https://github.com/stuttgart-things/pvectl) library (API token auth, verified TLS)
- ✅ Vault integration for secure credential management
- ✅ Build variables via comma-separated CLI string (`--vars`)
- ✅ Build variables via plain YAML file (`--vars-file`)
- ✅ Secret build variables via SOPS-encrypted YAML file (`--sops-file`)
- ✅ Multi-platform build support
- ✅ Local and remote build directory support
- ✅ Automated template management workflows

## Prerequisites

- Dagger CLI installed
- Docker runtime available
- Packer HCL configuration files
- vCenter access (for vSphere operations)
- Proxmox VE API token (for Proxmox operations) — create under *Datacenter → Permissions → API Tokens*

## Quick Start

### Build Image

```bash
# Basic Packer build
dagger call -m packer bake \
  --local-dir "." \
  --build-path tests/packer/hello/hello.pkr.hcl \
  --progress plain
```

### With Vault Authentication

```bash
# Build with Vault credentials
export VAULT_ROLE_ID=<role-id>
export VAULT_TOKEN=<token>
export VAULT_SECRET_ID=<secret-id>

dagger call -m packer bake \
--local-dir "/path/to/packer/builds/" \
--build-path ubuntu24-base-os.pkr.hcl \
--vault-addr https://vault.example.com:8200 \
--vault-role-id env:VAULT_ROLE_ID \
--vault-token env:VAULT_TOKEN \
--vault-secret-id env:VAULT_SECRET_ID \
--progress plain
```

### Build Variables

`Bake` accepts variables for `packer build` from three sources. They are applied
in order — **`--vars-file` → `--sops-file` → `--vars`** — so CLI values override
file values.

- `--vars` — comma-separated `key=value` pairs passed as `-var key=value`
- `--vars-file` — path (relative to the build dir) to a **plain YAML** file;
  converted to JSON inside the container with `yq` and consumed via
  `packer build -var-file=...`
- `--sops-file` — path (relative to the build dir) to a **SOPS-encrypted YAML**
  file; decrypted with `sops -d`, converted to JSON, then consumed via
  `-var-file=...`
- `--sops-age-key` — age private key (as a Dagger secret) exported into the
  container as `SOPS_AGE_KEY` for decryption

```bash
# Comma-separated CLI vars
dagger call -m packer bake \
  --local-dir "." \
  --build-path tests/packer/vars/template.pkr.hcl \
  --vars "name=tpl,environment=ci,owner=dagger" \
  --progress plain

# Plain YAML vars file
dagger call -m packer bake \
  --local-dir "." \
  --build-path tests/packer/vars/template.pkr.hcl \
  --vars-file vars.yaml \
  --progress plain

# Vars file + SOPS-encrypted secrets file + CLI override
export SOPS_AGE_KEY=$(cat age.key)

dagger call -m packer bake \
  --local-dir "." \
  --build-path tests/packer/vars/template.pkr.hcl \
  --vars-file vars.yaml \
  --sops-file secrets.enc.yaml \
  --sops-age-key env:SOPS_AGE_KEY \
  --vars "environment=prod" \
  --progress plain
```

To create a SOPS fixture:

```bash
age-keygen -o age.key
export SOPS_AGE_RECIPIENTS=$(grep 'public key:' age.key | awk '{print $4}')
sops -e --age "$SOPS_AGE_RECIPIENTS" secrets.plain.yaml > secrets.enc.yaml
```

### vCenter Operations

```bash
# Move VM template
export VCENTER_FQDN=https://vcenter.example.com/sdk
export VCENTER_USER=<username>
export VCENTER_PASSWORD=<password>

dagger call -m packer vcenteroperation \
  --operation move \
  --vcenter env:VCENTER_FQDN \
  --username env:VCENTER_USER \
  --password env:VCENTER_PASSWORD \
  --source /Datacenter/vm/folder/vm-name \
  --target /Datacenter/vm/templates/ \
  --progress plain

# Rename VM template
dagger call -m packer vcenteroperation \
  --operation rename \
  --vcenter env:VCENTER_FQDN \
  --username env:VCENTER_USER \
  --password env:VCENTER_PASSWORD \
  --source /Datacenter/vm/templates/old-name \
  --target new-name \
  --progress plain
```

```bash
# CHECK DATASTORES AND EXPORT TO FILE
dagger call -m packer \
check-datastores \
--vcenter env:VCENTER_FQDN \
--username env:VCENTER_USER \
--password env:VCENTER_PASSWORD \
--datacenter="LabUL" \
--progress plain -vv \
export --path=./datastore-info.txt
```

```bash
# CHECK NETORKS AND EXPORT TO FILE
dagger call -m packer \
check-netorks \
--vcenter env:VCENTER_FQDN \
--username env:VCENTER_USER \
--password env:VCENTER_PASSWORD \
--datacenter="LabUL" \
--progress plain -vv \
export --path=./network-info.txt
```

### Test Module

```bash
# Run comprehensive tests
task test-packer
```

## API Reference

### Image Building

```bash
# Basic build
dagger call -m packer bake \
  --local-dir ./packer-configs \
  --build-path image.pkr.hcl

# Build with Vault integration
dagger call -m packer bake \
  --local-dir ./packer-configs \
  --build-path image.pkr.hcl \
  --vault-addr https://vault.example.com:8200 \
  --vault-role-id env:VAULT_ROLE_ID \
  --vault-token env:VAULT_TOKEN \
  --vault-secret-id env:VAULT_SECRET_ID

# Build with mixed variable sources
dagger call -m packer bake \
  --local-dir ./packer-configs \
  --build-path image.pkr.hcl \
  --vars-file vars.yaml \
  --sops-file secrets.enc.yaml \
  --sops-age-key env:SOPS_AGE_KEY \
  --vars "name=my-template,environment=prod"
```

**`Bake` parameters:**

| Flag | Description |
| --- | --- |
| `--local-dir` | Local directory mounted as the packer source |
| `--build-path` | Path to the `.pkr.hcl` file (the parent is used as the build dir) |
| `--packer-version` | Packer version to install (default `1.15.1`) |
| `--arch` | Packer arch (default `linux_amd64`) |
| `--init-only` | Only run `packer init`, skip `packer build` |
| `--force` | Pass `-force` to `packer build` |
| `--vault-addr` / `--vault-token` / `--vault-role-id` / `--vault-secret-id` | Vault env/secret plumbing |
| `--vars` | Comma-separated `key=value` pairs appended as `-var` args |
| `--vars-file` | Plain YAML file (relative to build dir), normalized to JSON and passed via `-var-file` |
| `--sops-file` | SOPS-encrypted YAML (relative to build dir), decrypted + normalized to JSON and passed via `-var-file` |
| `--sops-age-key` | Age private key secret used to decrypt `--sops-file` (exported as `SOPS_AGE_KEY`) |

### vCenter Operations

```bash
# Move VM template
dagger call -m packer vcenteroperation \
  --operation move \
  --vcenter https://vcenter.example.com/sdk \
  --username env:VCENTER_USER \
  --password env:VCENTER_PASSWORD \
  --source /Datacenter/vm/source/path \
  --target /Datacenter/vm/target/path

# Rename VM template
dagger call -m packer vcenteroperation \
  --operation rename \
  --vcenter https://vcenter.example.com/sdk \
  --username env:VCENTER_USER \
  --password env:VCENTER_PASSWORD \
  --source /Datacenter/vm/templates/old-name \
  --target new-template-name
```

### Proxmox Operations

All Proxmox functions use the [pvectl](https://github.com/stuttgart-things/pvectl)
library (`pkg/pve`). Authentication uses an API token:

- `--proxmox-url` — API URL, e.g. `https://pve.example.com:8006` (`/api2/json` is optional)
- `--token-id` — token identifier, e.g. `automation@pam!dagger`
- `--token-secret` — token UUID

**TLS is verified**, one of these is required:

- `--proxmox-ca-cert` — the cluster CA (`/etc/pve/pve-root-ca.pem` on any node), PEM;
  survives certificate renewals
- `--proxmox-fingerprint` — the SHA-256 of the API certificate (`AB:CD:…`, as the web UI shows
  it under *node → System → Certificates*, or `pvectl tls`)

Every function returns `{"data": …}` JSON (the operations: the task UPID);
`get-proxmox-template-id-by-name` returns the plain VMID. VM descriptions are never returned.

For a LabUL template round (checks, migrate to ul-pve11, FS9100 copy,
repins) use the [pvectl module](../pvectl/README.md) (`template-promote`).

#### Create a Proxmox API token

1. **Create the token** — in the Proxmox web UI: *Datacenter → Permissions → API Tokens → Add*.
   - **User**: an existing user, e.g. `phermann@LabUL` (or `automation@pam`).
   - **Token ID**: a short name, e.g. `pvectl`. The full token ID becomes `phermann@LabUL!pvectl`.
   - **Privilege Separation**: leave **unchecked** if you want the token to inherit the user's permissions. If you keep it checked, you must grant permissions to the token itself (see step 2).
   - Copy the generated **secret UUID** immediately — it is only shown once.

2. **Grant permissions** (only with privilege separation) — *Datacenter → Permissions → Add → API Token Permission*:
   - `PVEAuditor` — read-only, sufficient for `check-proxmox-storage`, `check-proxmox-networks`, `list-proxmox-resources`, `get-proxmox-template-id-by-name`.
   - `PVEVMAdmin` on `/vms` — required for `proxmoxoperation` (move/rename/delete).

   > **Note:** If you see `{ "data": [] }` from `check-proxmox-storage` despite a successful call, the token authenticated but lacks `Datastore.Audit` on `/storage`.

#### LabUL start commands

The token and URL live in the sops file `~/.secrets.d/labul/proxmox.env`
(`PVE_API_URL`, `PVE_API_TOKEN_ID`, `PVE_API_TOKEN_SECRET`):

```bash
E=~/.secrets.d/labul/proxmox.env
M=github.com/stuttgart-things/dagger/packer
FP=33:91:DF:48:E5:34:B2:BC:76:AD:5C:92:AE:DB:8E:7D:3B:17:6D:A1:45:41:2C:44:12:1F:C1:EB:9D:65:57:13
PVE="--proxmox-url env:PVE_API_URL --token-id env:PVE_API_TOKEN_ID --token-secret env:PVE_API_TOKEN_SECRET --proxmox-fingerprint $FP"

# templates (VMID, node, name)
sops exec-env $E "dagger call -m $M list-proxmox-resources $PVE --resource-type vm" \
  | jq -r '.data[] | select(.template) | "\(.vmid) \(.node) \(.name)"'

# VMID of a template by name
sops exec-env $E "dagger call -m $M get-proxmox-template-id-by-name $PVE --name ubuntu26-base-20261116-0842 --pool stuttgart-things"

# VM-disk storages and bridges of ul-pve11
sops exec-env $E "dagger call -m $M check-proxmox-storage $PVE --node ul-pve11" | jq -r '.data[].storage'
sops exec-env $E "dagger call -m $M check-proxmox-networks $PVE --node ul-pve11" | jq -r '.data[].iface'
```

The fingerprint is the one of `ul-pve01.labul.sva.de` until 2028-01-28.

#### Generic examples

```bash
export PVE_URL=https://pve.example.com:8006
export PVE_TOKEN_ID='automation@pam!dagger'   # single quotes: ! triggers bash history expansion
export PVE_TOKEN_SECRET=<uuid>
export PVE_FP=AB:CD:...                       # or --proxmox-ca-cert file:./pve-root-ca.pem
PVE="--proxmox-url env:PVE_URL --token-id env:PVE_TOKEN_ID --token-secret env:PVE_TOKEN_SECRET --proxmox-fingerprint $PVE_FP"

# Migrate VM 9001 from pve1 → pve2 (offline)
dagger call -m packer proxmoxoperation $PVE --operation move --node pve1 --vmid 9001 --target pve2

# Rename VM 9001
dagger call -m packer proxmoxoperation $PVE --operation rename --node pve1 --vmid 9001 --target new-template-name

# Delete VM 9001 (purges disks + unreferenced)
dagger call -m packer proxmoxoperation $PVE --operation delete --node pve1 --vmid 9001

# Look up VMID by template name (optional --pool / --node narrow the search)
dagger call -m packer get-proxmox-template-id-by-name $PVE --name ubuntu26-labul-proxmox-base-os --pool stuttgart-things

# Golden image swap: lookup → delete old → rename new
OLD_ID=$(dagger call -m packer get-proxmox-template-id-by-name $PVE --name ubuntu26-labul-proxmox-base-os)
dagger call -m packer proxmoxoperation $PVE --operation delete --node ul-pve01 --vmid "$OLD_ID"
NEW_ID=$(dagger call -m packer get-proxmox-template-id-by-name $PVE --name 2026-04-25-ubuntu26-labul-proxmox)
dagger call -m packer proxmoxoperation $PVE --operation rename --node ul-pve01 --vmid "$NEW_ID" --target ubuntu26-labul-proxmox-base-os

# Datacenter-wide resources (optional filter: vm | storage | node)
dagger call -m packer list-proxmox-resources $PVE --resource-type vm

# Storages: every storage on every node, or the VM-disk storages of one node
dagger call -m packer check-proxmox-storage $PVE --node pve1

# Network interfaces/bridges on a node
dagger call -m packer check-proxmox-networks $PVE --node pve1
```

**Proxmox function reference:**

| Function | Description |
| --- | --- |
| `proxmoxoperation` | `move` (migrate to target node), `rename` (set new VM name), `delete` (destroy + purge disks) |
| `get-proxmox-template-id-by-name` | Returns VMID for a VM/template by `--name`; optional `--pool` / `--node` narrow the search; errors on zero or multiple matches |
| `check-proxmox-storage` | Lists storages with usage: every storage on every node, or with `--node` the VM-disk storages of that node |
| `check-proxmox-networks` | Lists interfaces/bridges on a node |
| `list-proxmox-resources` | Datacenter-wide resources (VMs/templates, storages, nodes), optional `--resource-type vm\|storage\|node` |

## Configuration Example

**hello.pkr.hcl:**
```hcl
packer {
  required_plugins {
    vsphere = {
      version = ">= 1.0.0"
      source  = "github.com/hashicorp/vsphere"
    }
  }
}

build {
  sources = ["source.vsphere-iso.example"]

  provisioner "shell" {
    inline = [
      "echo 'Hello from Packer!'",
      "apt-get update && apt-get upgrade -y"
    ]
  }
}
```

## Examples

See the [main README](../README.md#packer) for detailed usage examples.

## Testing

```bash
task test-packer
```

## Resources

- [Packer Documentation](https://packer.io/docs/)
- [vSphere Plugin](https://github.com/hashicorp/packer-plugin-vsphere)
- [HashiCorp Vault](https://vaultproject.io/)
