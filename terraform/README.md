# Terraform Dagger Module

This module provides Dagger functions for Terraform infrastructure automation including plan, apply, destroy operations with Vault integration.

## Features

- ✅ Terraform plan, apply, and destroy operations
- ✅ Variable management with JSON and environment variables
- ✅ Vault integration for secure secret management
- ✅ State file handling and export capabilities
- ✅ Output extraction and processing
- ✅ Multi-environment support

## Prerequisites

- Dagger CLI installed
- Docker runtime available
- Terraform configuration files

## Quick Start

### Apply Configuration

```bash
# Basic Terraform apply
dagger call -m terraform execute \
  --terraform-dir tests/terraform \
  --operation apply \
  -vv --progress plain \
  export --path=/tmp/terraform/
```

### With Variables

```bash
# Apply with variables
dagger call -m terraform execute \
  --terraform-dir tests/terraform \
  --variables "name=patrick,food=kaiserschmarrn" \
  --operation apply \
  -vv --progress plain \
  export --path=/tmp/terraform/
```

### Get Outputs

```bash
# Extract Terraform outputs
dagger call -m terraform output \
  --terraform-dir ~/projects/terraform/vms/dagger/ \
  -vv --progress plain
```

### Test Module

```bash
# Run comprehensive tests
task test-terraform
```

## API Reference

### Version Check

```bash
dagger call -m terraform version \
  --progress plain
```

### Execute Operations

```bash
# Apply with secret variables
dagger call -m terraform execute \
  --terraform-dir tests/terraform \
  --variables "name=patrick" \
  --secret-json-variables file://tests/terraform/terraform.tfvars.json \
  --operation apply \
  export --path=/tmp/terraform/

# Destroy infrastructure
dagger call -m terraform execute \
  --terraform-dir tests/terraform \
  --operation destroy \
  --progress plain
```

### Vault Integration

```bash
# Apply with Vault secrets
dagger call -m terraform execute \
  --terraform-dir /path/to/terraform \
  --vault-secret-id env:VAULT_SECRET_ID \
  --vault-role-id env:VAULT_ROLE_ID \
  --variables "vault_addr=https://vault.example.com:8200" \
  --operation apply \
  export --path=/tmp/terraform/
```

### Apply Part Of A Shared State, Safely

For a configuration that manages more than the caller owns. For example, one
list of secrets for many clusters: an untargeted apply with only the caller's
entries would plan to delete everyone else's.

```bash
dagger call -m terraform execute \
  --terraform-dir /path/to/terraform \
  --operation apply \
  --targets 'vault_kv_secret_v2.this["kv-a/app"],vault_kv_secret_v2.this["kv-a/db"]' \
  --refuse-destroy \
  export --path=/tmp/terraform/
```

- `--targets`: comma-separated resource addresses, one `-target` each. Also honoured by `destroy`.
- `--refuse-destroy` (apply only): plans to a file and aborts with
  `REFUSING TO APPLY` if any resource change contains `delete` (a replace
  does too). Otherwise it applies **that** plan, not a fresh one. The plan
  files hold values and are removed before the directory is returned.

### Pin A Hostname (`--bind-service`)

`/etc/hosts` is read-only inside a Dagger exec. When the engine's resolver
cannot answer a name reliably (a lab zone without public NS records), let
the caller's host forward the port and bind it under the real name. TLS and
SNI still see that name:

```bash
dagger call -m terraform execute \
  --terraform-dir /path/to/terraform \
  --bind-service tcp://10.31.103.9:443 \
  --bind-service-alias openbao.example.lab \
  --vault-addr https://openbao.example.lab \
  --operation apply
```

The two flags go together. One name per call; the service forwards the one port it was given.

### Output Extraction

```bash
dagger call -m terraform output \
  --terraform-dir /path/to/terraform \
  --progress plain
```

## Variable Priority

Variables are processed in the following priority order (highest to lowest):

1. `--variables` parameter (command line)
2. `--secret-json-variables` parameter (JSON file)
3. Environment variables
4. Terraform default values

## Examples

See the [main README](../README.md#terraform) for detailed usage examples.

## Testing

```bash
task test-terraform
```

## Resources

- [Terraform Documentation](https://terraform.io/docs/)
- [HashiCorp Vault](https://vaultproject.io/)
- [Terraform Providers](https://registry.terraform.io/)
