# Packer Dagger Module

This module provides Dagger functions for Packer operations including image building, vCenter operations, and Vault integration.

## Features

- ✅ Packer image building with HCL configurations
- ✅ vCenter VM template operations (move, rename)
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
- vCenter access (for VM operations)

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
| `--packer-version` | Packer version to install (default `1.12.0`) |
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
