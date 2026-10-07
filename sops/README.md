# SOPS Dagger Module

This module provides Dagger functions for SOPS (Secrets OPerationS) encryption and decryption operations with AGE key support.

## Features

- ✅ AGE key generation, public key derivation
- ✅ SOPS config generation (incl. Flux-style `encrypted_regex` rules)
- ✅ Encryption/decryption, partial encryption (`--encrypted-regex`, `--sops-config`)
- ✅ Single-value read (`--extract`) and update (`set`) without a plaintext file
- ✅ Recipient changes (`update-keys`)
- ✅ Pinned sops (3.13.3) and age (1.3.2), kept current by Renovate
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
export AGE_PUB=$(grep "public key" age-key.txt | awk '{print $NF}')
dagger call -m sops encrypt \
  --age-key env:AGE_PUB \
  --plaintext-file ./secrets.yaml \
  --file-extension yaml \
  export --path=./secrets.enc.yaml
```

### Decrypt File

```bash
# Decrypt with age key (file: keeps the private key out of your environment)
dagger call -m sops decrypt \
  --age-key file:./age-key.txt \
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

# Verify (never echo the private key)
echo "Public:  $AGE_PUBLIC_KEY"
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
  - path_regex: '.*\.yaml'
    age: "age1..." # pragma: allowlist secret
  - path_regex: '.*\.json'
    age: "age1..." # pragma: allowlist secret
  - path_regex: '.*\.env'
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

The examples use a throwaway key file `age-key.txt` (from `generate-age-key`).
Private keys and secret values are always passed as Dagger secrets
(`file:` or `env:`), so they never show up in `--progress plain` output.

### generate-age-key

Generates a new AGE key pair. Never cached: every call returns a new key.

```bash
dagger call -m sops generate-age-key export --path=./age-key.txt
```

### age-public-key

Derives the public key (`age1...`) from a private key (`age-keygen -y`).

```bash
dagger call -m sops age-public-key --age-key file:./age-key.txt
```

### generate-sops-config

Writes a `.sops.yaml` with one creation rule per file extension (default `yaml,json`), or one rule for `--path-regex`. `--encrypted-regex` is added to every rule; `--age-public-key` takes several comma-separated recipients.

```bash
dagger call -m sops generate-sops-config \
  --age-public-key "$AGE_PUB" \
  --path-regex '.*\.enc\.yaml$' \
  --encrypted-regex '^(data|stringData)$' \
  export --path=./.sops.yaml
```

```yaml
---
creation_rules:
  - path_regex: '.*\.enc\.yaml$'
    encrypted_regex: '^(data|stringData)$'
    age: "age1..." # pragma: allowlist secret
```

### encrypt

Encrypts a file. `--age-key` takes the public key(s). With `--sops-config` its creation rule applies (recipients, `encrypted_regex`, ...); `--encrypted-regex` works without a config file.

```bash
dagger call -m sops encrypt \
  --age-key env:AGE_PUB \
  --plaintext-file ./config.yaml \
  --file-extension yaml \
  export --path=./config.enc.yaml
```

| Parameter | Description |
|---|---|
| `--age-key` | AGE public key(s), comma-separated (required) |
| `--plaintext-file` | File to encrypt (required) |
| `--file-extension` | `yaml`, `json`, `env` (default `yaml`) |
| `--sops-config` | `.sops.yaml` whose creation rules apply. Rules are matched against `encrypted.<extension>` |
| `--encrypted-regex` | Encrypt only values whose keys match, e.g. `'^(data|stringData)$'` |

### decrypt

Decrypts a file, or with `--extract` returns a single value.

```bash
dagger call -m sops decrypt \
  --age-key file:./age-key.txt \
  --encrypted-file ./secret.enc.yaml \
  --extract '["stringData"]["password"]' \
  contents
```

| Parameter | Description |
|---|---|
| `--age-key` | AGE private key (required) |
| `--encrypted-file` | File to decrypt (required) |
| `--sops-config` | Optional `.sops.yaml` |
| `--extract` | sops path of one value, e.g. `'["stringData"]["KEY"]'` |

### set

Changes one value in an encrypted file without writing the rest out in plaintext (`sops set`). The value is a Secret and is passed to sops through a file, not the command line. Without `--json-value` it is set as a string.

```bash
export NEW_PASSWORD=...   # or file:./password.txt
dagger call -m sops set \
  --encrypted-file ./secret.enc.yaml \
  --path '["stringData"]["password"]' \
  --value env:NEW_PASSWORD \
  --age-key file:./age-key.txt \
  export --path=./secret.enc.yaml
```

### update-keys

Re-encrypts the file's data key for the recipients the `.sops.yaml` now lists (`sops updatekeys -y`), e.g. after adding a second recipient. The creation rule has to match the file's name.

```bash
dagger call -m sops generate-sops-config \
  --age-public-key "$AGE_PUB,$AGE_PUB_COLLEAGUE" \
  --encrypted-regex '^(data|stringData)$' \
  export --path=./.sops.yaml

dagger call -m sops update-keys \
  --encrypted-file ./secret.enc.yaml \
  --sops-config ./.sops.yaml \
  --age-key file:./age-key.txt \
  export --path=./secret.enc.yaml
```

## Flux: encrypt a Kubernetes Secret

Flux's kustomize-controller can only apply a SOPS Secret whose `apiVersion`, `kind` and `metadata` stay readable. Encrypt only `data`/`stringData`:

```bash
cat > secret.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: app-credentials
  namespace: apps
type: Opaque
stringData:
  password: change-me
EOF

dagger call -m sops encrypt \
  --age-key env:AGE_PUB \
  --plaintext-file ./secret.yaml \
  --encrypted-regex '^(data|stringData)$' \
  export --path=./secret.enc.yaml
```

```yaml
apiVersion: v1
kind: Secret
metadata:
    name: app-credentials
    namespace: apps
type: Opaque
stringData:
    password: ENC[AES256_GCM,data:...,type:str]
sops:
    age:
        - recipient: age1... # pragma: allowlist secret
    encrypted_regex: ^(data|stringData)$
    ...
```

The same works with `--sops-config` and a rule from `generate-sops-config --encrypted-regex '^(data|stringData)$'`.

## Configuration

SOPS uses `.sops.yaml` configuration files for encryption rules:

```yaml
creation_rules:
  - path_regex: \.yaml$
    encrypted_regex: ^(data|stringData)$
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
