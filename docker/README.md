# Docker Dagger Module

This module provides Dagger functions for Docker container operations including building, linting, and publishing to registries.

## Features

- ✅ Dockerfile linting and validation
- ✅ Multi-platform container building
- ✅ Registry publishing with authentication
- ✅ Temporary registry support (ttl.sh)
- ✅ Build optimization and layer caching
- ✅ Security scanning integration

## Prerequisites

- Dagger CLI installed
- Docker runtime available
- Dockerfile in source directory

## Quick Start

### Lint Dockerfile

```bash
# Lint Dockerfile
dagger call -m docker lint \
  --src tests/docker \
  -vv --progress plain
```

### Build Container

```bash
# Build container image
dagger call -m docker build \
  --src tests/docker \
  -vv --progress plain
```

### Build and Push

```bash
# Push to temporary registry (no auth)
dagger call -m docker build-and-push \
  --source tests/docker \
  --repository-name stuttgart-things/test \
  --registry-url ttl.sh \
  --tag 1.2.3 \
  -vv --progress plain

# Push to GitHub Container Registry (with auth)
dagger call -m docker build-and-push \
  --source tests/docker \
  --repository-name stuttgart-things/test \
  --registry-url ghcr.io \
  --tag 1.2.3 \
  --registry-username env:GITHUB_USER \
  --registry-password env:GITHUB_TOKEN \
  -vv --progress plain
```

### Test Module

```bash
# Run comprehensive tests
task test-docker
```

## API Reference

### Dockerfile Linting

```bash
dagger call -m docker lint \
  --src tests/docker \
  --progress plain
```

### Container Building

```bash
dagger call -m docker build \
  --src tests/docker \
  --progress plain
```

### Build and Push (Unauthenticated)

```bash
dagger call -m docker build-and-push \
  --source tests/docker \
  --repository-name organization/image \
  --registry-url ttl.sh \
  --tag version \
  --progress plain
```

### Build and Push (Authenticated)

```bash
dagger call -m docker build-and-push \
  --source tests/docker \
  --repository-name organization/image \
  --registry-url ghcr.io \
  --tag version \
  --registry-username env:REGISTRY_USER \
  --registry-password env:REGISTRY_TOKEN \
  --progress plain
```

## Supported Registries

- **GitHub Container Registry** (`ghcr.io`)
- **Docker Hub** (`docker.io`)
- **Harbor** (custom registry)
- **TTL.sh** (temporary registry, no auth required)
- **AWS ECR** (with proper authentication)
- **Google Container Registry** (`gcr.io`)

## Examples

See the [main README](../README.md#docker) for detailed usage examples.

## Testing

```bash
task test-docker
```

## Resources

- [Docker Documentation](https://docs.docker.com/)
- [Dockerfile Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Container Registry Documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)