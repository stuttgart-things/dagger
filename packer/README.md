# Packer Dagger Module

This module provides Dagger functions for Packer operations including image building, vCenter and Proxmox operations, and Vault integration.

## Features

- ✅ Packer image building with HCL configurations
- ✅ vCenter VM template operations (move, rename, delete) + datastore/network inspection
- ✅ Proxmox VE operations (move/migrate, rename, delete) + storage/network/resource inspection via REST API (pure Go, API token auth)
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

All Proxmox functions talk to the Proxmox VE REST API directly in Go
(`net/http`, no extra deps, no `curl`/`jq`). Authentication uses an API token:

- `--proxmox-url` — base URL, e.g. `https://pve.example.com:8006`
- `--token-id` — token identifier, e.g. `automation@pam!dagger`
- `--token-secret` — token UUID

TLS verification is currently skipped (homelab-friendly).

#### Create a Proxmox API token

1. **Create the token** — in the Proxmox web UI: *Datacenter → Permissions → API Tokens → Add*.
   - **User**: an existing user, e.g. `phermann@LabUL` (or `automation@pam`).
   - **Token ID**: a short name, e.g. `dagger`. The full token ID becomes `phermann@LabUL!dagger`.
   - **Privilege Separation**: leave **unchecked** if you want the token to inherit the user's permissions. If you keep it checked, you must grant permissions to the token itself (see step 2).
   - Copy the generated **secret UUID** immediately — it is only shown once.

2. **Grant permissions** — *Datacenter → Permissions → Add → API Token Permission* (or *User Permission* if privilege separation is disabled):
   - **Path**: `/` for full visibility, or scope it down (e.g. `/storage`, `/nodes/<node>`, `/vms`).
   - **API Token**: the token from step 1 (e.g. `phermann@LabUL!dagger`).
   - **Role**:
     - `PVEAuditor` — read-only, sufficient for `check-proxmox-storage`, `check-proxmox-networks`, `list-proxmox-resources`.
     - `PVEVMAdmin` on `/vms` — required for `proxmoxoperation` (move/rename/delete).
     - `PVEDatastoreAdmin` on `/storage` — if you need to manage storage.

3. **Export and use**:

   ```bash
   export PVE_URL="https://pve.example.com:8006"
   export PVE_TOKEN_ID='phermann@LabUL!dagger'   # single quotes — ! triggers bash history expansion
   export PVE_TOKEN_SECRET="<uuid-from-step-1>"
   ```

   > **Note:** If you see `{ "data": [] }` from `check-proxmox-storage` despite a successful call, the token authenticated but lacks `Datastore.Audit` on `/storage`. Re-check step 2.

```bash
export PVE_URL=https://pve.example.com:8006
export PVE_TOKEN_ID='automation@pam!dagger'
export PVE_TOKEN_SECRET=<uuid>

# Migrate VM 9001 from pve1 → pve2
dagger call -m packer proxmoxoperation \
  --operation move \
  --node pve1 --vmid 9001 --target pve2 \
  --proxmox-url env:PVE_URL \
  --token-id env:PVE_TOKEN_ID \
  --token-secret env:PVE_TOKEN_SECRET

# Rename VM 9001
dagger call -m packer proxmoxoperation \
  --operation rename \
  --node pve1 --vmid 9001 --target new-template-name \
  --proxmox-url env:PVE_URL \
  --token-id env:PVE_TOKEN_ID \
  --token-secret env:PVE_TOKEN_SECRET

# Delete VM 9001 (purges disks + unreferenced)
dagger call -m packer proxmoxoperation \
  --operation delete \
  --node pve1 --vmid 9001 \
  --proxmox-url env:PVE_URL \
  --token-id env:PVE_TOKEN_ID \
  --token-secret env:PVE_TOKEN_SECRET
```

```bash
# Datacenter-wide resources (optional filter: vm | storage | node | sdn)
dagger call -m packer list-proxmox-resources \
  --proxmox-url env:PVE_URL \
  --token-id env:PVE_TOKEN_ID \
  --token-secret env:PVE_TOKEN_SECRET \
  --resource-type vm

# Storage pools (cluster-wide or per-node)
dagger call -m packer check-proxmox-storage \
  --proxmox-url env:PVE_URL \
  --token-id env:PVE_TOKEN_ID \
  --token-secret env:PVE_TOKEN_SECRET \
  --node pve1

# Network interfaces/bridges on a node
dagger call -m packer check-proxmox-networks \
  --proxmox-url env:PVE_URL \
  --token-id env:PVE_TOKEN_ID \
  --token-secret env:PVE_TOKEN_SECRET \
  --node pve1
```

**Proxmox function reference:**

| Function | Description |
| --- | --- |
| `proxmoxoperation` | `move` (migrate to target node), `rename` (set new VM name), `delete` (destroy + purge disks) |
| `check-proxmox-storage` | Lists storage pools with usage; cluster-wide if `--node` omitted |
| `check-proxmox-networks` | Lists interfaces/bridges on a node |
| `list-proxmox-resources` | Datacenter-wide resources, optional `--resource-type vm\|storage\|node\|sdn` |

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
