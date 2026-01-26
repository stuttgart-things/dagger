# SOPS Dagger Module

This module provides Dagger functions for SOPS (Secrets OPerationS) encryption and decryption operations with AGE key support.

## Features

- ✅ AGE key generation
- ✅ SOPS config generation
- ✅ Secret encryption/decryption
- ✅ AGE key support
- ✅ Multiple file format support (YAML, JSON, ENV)

## Prerequisites

- Dagger CLI installed
- AGE keys configured (or generate with this module)

## Quick Start

### Generate AGE Key

```bash
# Generate age key pair
dagger call -m sops generate-age-key \
  export --path=./age-key.txt
```

### Generate SOPS Config

```bash
# Generate .sops.yaml config
export AGE_KEY=$(grep "public key" age-key.txt | awk '{print $NF}')
dagger call -m sops generate-sops-config \
  --age-public-key "$AGE_KEY" \
  --file-extensions "yaml,json,log" \
  export --path=./.sops.yaml
```

### Encrypt File with AGE

```bash
# Encrypt file with age
export AGE_KEY=$(grep "public key" age-key.txt | awk '{print $NF}')
dagger call -m sops encrypt \
  --age-key env:AGE_KEY \
  --plaintext-file ./secrets.yaml \
  --file-extension yaml \
  export --path=./secrets.enc.yaml
```

### Decrypt File

```bash
# Decrypt with age key
export SOPS_KEY=$(cat age-key.txt)
dagger call -m sops decrypt \
  --sops-key env:SOPS_KEY \
  --encrypted-file ./secrets.enc.yaml \
  export --path=./secrets.dec.yaml
```

## API Reference

### GenerateAgeKey

Generates a new AGE key pair using age-keygen.

```bash
dagger call -m sops generate-age-key \
  export --path=./age-key.txt
```

### GenerateSopsConfig

Generates a .sops.yaml configuration file with creation rules for the given AGE key.

```bash
# Generate config for yaml and json files (default)
dagger call -m sops generate-sops-config \
  --age-public-key "age19vgzvmpt9tdlcsu8rzaacj397yz8gguz38nsmuy6eeelt5vjsyms542xtm" \ # pragma: allowlist secret
  export --path=./.sops.yaml

# Generate config with custom file extensions
dagger call -m sops generate-sops-config \
  --age-public-key "age19vgzvmpt9tdlcsu8rzaacj397yz8gguz38nsmuy6eeelt5vjsyms542xtm" \ # pragma: allowlist secret
  --file-extensions "yaml,json,env,log" \
  export --path=./.sops.yaml
```

**Parameters:**
- `--age-public-key`: AGE public key (required)
- `--file-extensions`: Comma-separated list of extensions (default: "yaml,json")

### Encrypt

Encrypts a plaintext file using SOPS with an AGE key.

```bash
dagger call -m sops encrypt \
  --age-key env:AGE_KEY \
  --plaintext-file ./config.yaml \
  --file-extension yaml \
  --sops-config ./.sops.yaml \
  export --path=./config.enc.yaml
```

**Parameters:**
- `--age-key`: AGE public key (required)
- `--plaintext-file`: File to encrypt (required)
- `--file-extension`: Format - yaml, json, env (default: yaml)
- `--sops-config`: Optional .sops.yaml config file

### Decrypt

Decrypts a SOPS-encrypted file using an AGE key.

```bash
dagger call -m sops decrypt \
  --sops-key env:SOPS_KEY \
  --encrypted-file ./config.enc.yaml \
  export --path=./config.dec.yaml
```

**Parameters:**
- `--sops-key`: AGE private key (required)
- `--encrypted-file`: Encrypted file to decrypt (required)

## Configuration

SOPS uses `.sops.yaml` configuration files for encryption rules:

```yaml
creation_rules:
  - path_regex: \.yaml$
    age: age1234567890abcdef...
```

## AGE Keys

- **Pros**: Modern, simple, fast
- **Use Case**: Local development, CI/CD
- **Format**: `age1234567890abcdef...`

## Security Best Practices

1. **Key Rotation**: Regularly rotate encryption keys
2. **Access Control**: Limit key access to necessary personnel
3. **Backup**: Securely backup decryption keys
4. **Environment Separation**: Use different keys per environment

## Resources

- [SOPS Documentation](https://github.com/getsops/sops)
- [Age Encryption](https://age-encryption.org/)
