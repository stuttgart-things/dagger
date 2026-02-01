# Crane Dagger Module

A Dagger module for copying container images between registries using Google's [crane](https://github.com/google/go-containerregistry/tree/main/cmd/crane) CLI.

## Features

- Copy container images between registries
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
| `version` | string | `latest` | Crane version to install |

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
