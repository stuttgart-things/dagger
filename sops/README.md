# SOPS Dagger Module

This module provides Dagger functions for SOPS (Secrets OPerationS) encryption and decryption operations with age, GPG, and cloud KMS support.

## Features

- ✅ Secret encryption/decryption
- ✅ Age key support
- ✅ GPG key integration
- ✅ Cloud KMS support (AWS, GCP, Azure)
- ✅ Kubernetes secrets management
- ✅ Multiple file format support (YAML, JSON, ENV)

## Prerequisites

- Dagger CLI installed
- SOPS CLI available
- Age keys or GPG keys configured
- Cloud credentials (for KMS operations)

## Quick Start

### Encrypt File with Age

```bash
# Generate age key first
age-keygen -o age-key.txt

# Encrypt file with age
export AGE_KEY=$(cat age-key.txt)
dagger call -m sops encrypt-file \
  --src ./secrets.yaml \
  --age-key env:AGE_KEY \
  export --path=./secrets.enc.yaml
```

### Decrypt File

```bash
# Decrypt with age key
export AGE_KEY=$(cat age-key.txt)
dagger call -m sops decrypt-file \
  --src ./secrets.enc.yaml \
  --age-key env:AGE_KEY \
  export --path=./secrets.dec.yaml
```

### Kubernetes Secrets

```bash
# Encrypt Kubernetes secret
dagger call -m sops encrypt-k8s-secret \
  --secret-name my-secret \
  --namespace default \
  --data-file ./secret-data.yaml \
  --age-key env:AGE_KEY \
  export --path=./secret.enc.yaml
```

## API Reference

### File Operations

```bash
# Encrypt file with multiple keys
dagger call -m sops encrypt-file \
  --src ./config.yaml \
  --age-key env:AGE_KEY \
  --gpg-fingerprint "1234567890ABCDEF" \
  export --path=./config.enc.yaml

# Decrypt file
dagger call -m sops decrypt-file \
  --src ./config.enc.yaml \
  --age-key env:AGE_KEY \
  export --path=./config.dec.yaml
```

### Key Management

```bash
# Generate age key pair
dagger call -m sops generate-age-key \
  export --path=./age-key.txt

# Import GPG key
dagger call -m sops import-gpg-key \
  --gpg-key-file ./my-key.asc
```

### Cloud KMS Integration

```bash
# Encrypt with AWS KMS
dagger call -m sops encrypt-file \
  --src ./secrets.yaml \
  --aws-kms-key "arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012" \
  export --path=./secrets.enc.yaml

# Encrypt with GCP KMS
dagger call -m sops encrypt-file \
  --src ./secrets.yaml \
  --gcp-kms-key "projects/my-project/locations/global/keyRings/my-ring/cryptoKeys/my-key" \
  export --path=./secrets.enc.yaml
```

## Configuration

SOPS uses `.sops.yaml` configuration files for encryption rules:

```yaml
creation_rules:
  - path_regex: \.dev\.yaml$
    age: age1234567890abcdef...
  - path_regex: \.prod\.yaml$
    kms: arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012
  - path_regex: \.json$
    pgp: "1234567890ABCDEF1234567890ABCDEF12345678"
```

## Key Types

### Age Keys
- **Pros**: Modern, simple, fast
- **Use Case**: Local development, CI/CD
- **Format**: `age1234567890abcdef...`

### GPG Keys
- **Pros**: Established, web of trust
- **Use Case**: Team environments, existing GPG infrastructure
- **Format**: Fingerprint `1234567890ABCDEF`

### Cloud KMS
- **Pros**: Centralized, auditable, HSM-backed
- **Use Case**: Production, compliance requirements
- **Providers**: AWS KMS, GCP KMS, Azure Key Vault

## Security Best Practices

1. **Key Rotation**: Regularly rotate encryption keys
2. **Access Control**: Limit key access to necessary personnel
3. **Backup**: Securely backup decryption keys
4. **Audit**: Monitor key usage and file access
5. **Environment Separation**: Use different keys per environment

## Examples

See the [main README](../README.md#sops) for detailed usage examples.

## Resources

- [SOPS Documentation](https://github.com/mozilla/sops)
- [Age Encryption](https://age-encryption.org/)
- [GPG Guide](https://gnupg.org/documentation/)
- [AWS KMS](https://aws.amazon.com/kms/)
- [GCP KMS](https://cloud.google.com/kms)