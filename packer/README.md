# Packer Dagger Module

This module provides Dagger functions for Packer operations including image building, vCenter operations, and Vault integration.

## Features

- ✅ Packer image building with HCL configurations
- ✅ vCenter VM template operations (move, rename)
- ✅ Vault integration for secure credential management
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
```

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
