# Crane Dagger Module

A Dagger module for copying container images between registries using Google's [crane](https://github.com/google/go-containerregistry/tree/main/cmd/crane) CLI.

## Features

- Copy container images between registries
- Resolve a tag to the digest it currently points at, and check that two tags agree
- Support for source and target registry authentication
- Platform-specific image copying
- Insecure registry support (for self-hosted/air-gapped setups)
- Docker config.json authentication support
- No local Docker daemon required

## Prerequisites

- Dagger CLI installed
- Registry credentials (for private registries)

## Module Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `base-image` | string | `cgr.dev/chainguard/wolfi-base:latest` | Base image for crane container |
| `version` | string | `0.22.1` | Crane release (or `latest`); the binary is copied from `gcr.io/go-containerregistry/crane:v<version>` |

## Copy Function

The `copy` function copies an image from a source registry to a target registry.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `source` | string | Yes | - | Source image reference (e.g., `nginx:latest`) |
| `target` | string | Yes | - | Target image reference (e.g., `ghcr.io/user/nginx:latest`) |
| `source-registry` | string | No | Auto-detected | Source registry URL |
| `source-username` | string | No | - | Source registry username |
| `source-password` | Secret | No | - | Source registry password |
| `target-registry` | string | No | Auto-detected | Target registry URL |
| `target-username` | string | No | - | Target registry username |
| `target-password` | Secret | No | - | Target registry password |
| `insecure` | bool | No | `false` | Allow insecure registries |
| `platform` | string | No | `linux/amd64` | Target platform |
| `docker-config-secret` | Secret | No | - | Docker config.json for authentication |

## Digest Function

The `digest` function resolves a reference to the digest it points at right
now and prints it bare (`sha256:…`), so it can be used as `repo@$(…)`.

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `ref` | string | Yes | - | Reference to resolve |
| `platform` | string | No | *(empty)* | Resolve one platform's manifest instead of the index |
| `registry` | string | No | Auto-detected | Registry to log in to |
| `username` | string | No | - | Registry username |
| `password` | Secret | No | - | Registry password |
| `insecure` | bool | No | `false` | Allow insecure registries |

`platform` defaults to empty, **unlike `copy`**. A multi-arch release is signed
on its index digest; resolving one platform's manifest returns a digest nothing
signed. For a single-manifest image crane ignores `platform` altogether.

A reference that does not exist is an error with the registry's message
(`MANIFEST_UNKNOWN`), never an empty string. The result is never served from
Dagger's cache, since a tag can move between runs.

```bash
DIGEST=$(dagger call -m crane digest --ref ghcr.io/myorg/app:1.2.3)
cosign sign "ghcr.io/myorg/app@${DIGEST}"
```

## SameDigest Function

The `same-digest` function checks that two references point at the same digest
and returns it; otherwise it fails and names both digests. It takes the same
`platform`, credential and `insecure` flags as `digest`.

```bash
# fails when latest and the release tag name different images
dagger call -m crane same-digest \
  --ref-a ghcr.io/myorg/app:latest \
  --ref-b ghcr.io/myorg/app:1.2.3
```

## Usage Examples

### Copy Public Image to Private Registry

```bash
# Copy nginx from Docker Hub to GitHub Container Registry
dagger call -m crane copy \
  --source nginx:latest \
  --target ghcr.io/myorg/nginx:latest \
  --target-username myusername \
  --target-password env:GITHUB_TOKEN
```

### Copy Between Private Registries

```bash
# Copy with authentication on both ends
dagger call -m crane copy \
  --source harbor.internal.com/project/app:v1.0 \
  --target ghcr.io/myorg/app:v1.0 \
  --source-username admin \
  --source-password env:HARBOR_PASSWORD \
  --target-username myusername \
  --target-password env:GITHUB_TOKEN
```

### Copy with Platform Selection

```bash
# Copy ARM64 variant
dagger call -m crane copy \
  --source nginx:latest \
  --target ghcr.io/myorg/nginx:latest-arm64 \
  --target-username myusername \
  --target-password env:GITHUB_TOKEN \
  --platform linux/arm64
```

### Copy with Insecure Registry

```bash
# For self-signed certificates or HTTP registries
dagger call -m crane copy \
  --source my-internal-registry:5000/app:latest \
  --target ghcr.io/myorg/app:latest \
  --target-username myusername \
  --target-password env:GITHUB_TOKEN \
  --insecure
```

### Copy Using Docker Config

```bash
# Use existing Docker config.json for authentication
dagger call -m crane copy \
  --source private-registry.com/app:latest \
  --target ghcr.io/myorg/app:latest \
  --docker-config-secret file:~/.docker/config.json
```

## Registry-Specific Examples

### GitHub Container Registry (ghcr.io)

```bash
dagger call -m crane copy \
  --source nginx:latest \
  --target ghcr.io/USERNAME/nginx:latest \
  --target-username USERNAME \
  --target-password env:GITHUB_TOKEN
```

### Docker Hub

```bash
dagger call -m crane copy \
  --source alpine:latest \
  --target docker.io/USERNAME/alpine:latest \
  --target-username USERNAME \
  --target-password env:DOCKER_PASSWORD
```

### Harbor

```bash
dagger call -m crane copy \
  --source nginx:latest \
  --target harbor.example.com/library/nginx:latest \
  --target-username admin \
  --target-password env:HARBOR_PASSWORD
```

### ttl.sh (Ephemeral Registry)

```bash
# ttl.sh is public, no source auth needed
dagger call -m crane copy \
  --source ttl.sh/my-temp-image:1h \
  --target ghcr.io/myorg/my-image:latest \
  --target-username myusername \
  --target-password env:GITHUB_TOKEN
```

## How It Works

1. Creates a Wolfi-based container with crane CLI installed
2. Authenticates to source/target registries (if credentials provided)
3. Executes `crane copy` with the specified parameters
4. Returns the command output

Crane operates directly on registry APIs without requiring a local Docker daemon, making it efficient for CI/CD pipelines.

## Resources

- [Crane Documentation](https://github.com/google/go-containerregistry/tree/main/cmd/crane)
- [Dagger Documentation](https://docs.dagger.io/)
- [OCI Distribution Spec](https://github.com/opencontainers/distribution-spec)
