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

## Complete Tutorial

This tutorial walks through a complete encryption/decryption workflow with explicit file exports at each step.

### Step 1: Generate a new AGE key pair

```bash
# Generate and view the key
dagger call -m ./sops generate-age-key contents

# Export to file
dagger call -m ./sops generate-age-key export --path="/tmp/age-key.txt"
```

**Output (`/tmp/age-key.txt`):**
```
# created: 2026-01-30T15:02:27Z
# public key: age1j9rj7x597dsxwt7y66nvlamr0py8eqnp6sp0gnjkmaw2d3ckdvfsm4gh0t # pragma: allowlist secret
AGE-SECRET-KEY-1... # pragma: allowlist secret
```

**Extract keys for later use:**
```bash
# Extract public key (line 2, after "# public key: ")
export AGE_PUBLIC_KEY=$(grep "public key:" /tmp/age-key.txt | cut -d' ' -f4)

# Extract private key (line 3)
export AGE_PRIVATE_KEY=$(grep "AGE-SECRET-KEY" /tmp/age-key.txt)

# Verify
echo "Public:  $AGE_PUBLIC_KEY"
echo "Private: $AGE_PRIVATE_KEY"
```

### Step 2: Generate a SOPS config file

```bash
# Generate and view the config
dagger call -m ./sops generate-sops-config \
  --age-public-key="$AGE_PUBLIC_KEY" \
  --file-extensions="yaml,json,env" \
  contents

# Export to file
dagger call -m ./sops generate-sops-config \
  --age-public-key="$AGE_PUBLIC_KEY" \
  --file-extensions="yaml,json,env" \
  export --path="/tmp/.sops.yaml"
```

**Output (`/tmp/.sops.yaml`):**
```yaml
---
creation_rules:
  - path_regex: .*\.yaml
    age: "age1..." # pragma: allowlist secret
  - path_regex: .*\.json
    age: "age1..." # pragma: allowlist secret
  - path_regex: .*\.env
    age: "age1..." # pragma: allowlist secret
```

### Step 3: Create a plaintext secrets file

```bash
cat <<EOF > /tmp/secrets.yaml
database:
  host: localhost
  username: admin
  password: super-secret-password
api_key: sk-12345-abcdef
EOF

cat /tmp/secrets.yaml
```

### Step 4: Encrypt the secrets file

```bash
# Encrypt and view the result
AGE_PUBLIC_KEY="$AGE_PUBLIC_KEY" \
dagger call -m ./sops encrypt \
  --age-key="env:AGE_PUBLIC_KEY" \
  --plaintext-file="/tmp/secrets.yaml" \
  --file-extension="yaml" \
  contents

# Export encrypted file
AGE_PUBLIC_KEY="$AGE_PUBLIC_KEY" \
dagger call -m ./sops encrypt \
  --age-key="env:AGE_PUBLIC_KEY" \
  --plaintext-file="/tmp/secrets.yaml" \
  --file-extension="yaml" \
  export --path="/tmp/secrets.enc.yaml"
```

**Output (`/tmp/secrets.enc.yaml`):**
```yaml
database:
    host: ENC[AES256_GCM,data:...,type:str]
    username: ENC[AES256_GCM,data:...,type:str]
    password: ENC[AES256_GCM,data:...,type:str]
api_key: ENC[AES256_GCM,data:...,type:str]
sops:
    age:
        - recipient: age1... # pragma: allowlist secret
          enc: |
            -----BEGIN AGE ENCRYPTED FILE-----
            ...
            -----END AGE ENCRYPTED FILE-----
    ...
```

### Step 5: Decrypt the encrypted file

```bash
# Decrypt and view the result
AGE_PRIVATE_KEY="$AGE_PRIVATE_KEY" \
dagger call -m ./sops decrypt \
  --age-key="env:AGE_PRIVATE_KEY" \
  --encrypted-file="/tmp/secrets.enc.yaml" \
  contents

# Export decrypted file
AGE_PRIVATE_KEY="$AGE_PRIVATE_KEY" \
dagger call -m ./sops decrypt \
  --age-key="env:AGE_PRIVATE_KEY" \
  --encrypted-file="/tmp/secrets.enc.yaml" \
  export --path="/tmp/secrets.decrypted.yaml"
```

**Output (`/tmp/secrets.decrypted.yaml`):**
```yaml
database:
  host: localhost
  username: admin
  password: super-secret-password
api_key: sk-12345-abcdef
```

### Step 6: Verify round-trip

```bash
# Compare original and decrypted
diff /tmp/secrets.yaml /tmp/secrets.decrypted.yaml && echo "✓ Files match!"
```

### Output Methods Reference

| Method | Description |
|--------|-------------|
| `contents` | Returns file content as a string (printed to stdout) |
| `export --path="/tmp/out.yaml"` | Saves the file to a local path |
| `name` | Returns the filename |
| `size` | Returns the file size in bytes |

### Files created in this tutorial

```
/tmp/
├── age-key.txt            # AGE key pair
├── .sops.yaml             # SOPS configuration
├── secrets.yaml           # Original plaintext
├── secrets.enc.yaml       # Encrypted file
└── secrets.decrypted.yaml # Decrypted (should match original)
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
