# Crane Dagger Module

This module provides Dagger functions for container registry operations using Google's crane CLI tool for efficient image management and manipulation.

## Features

- ✅ Container image copying between registries
- ✅ Image manifest inspection and manipulation
- ✅ Multi-architecture image handling
- ✅ Registry authentication support
- ✅ Image layer analysis
- ✅ Efficient image operations (no local Docker required)

## Prerequisites

- Dagger CLI installed
- Registry credentials (for private registries)
- Network access to source and destination registries

## Quick Start

### Copy Image Between Registries

```bash
# Copy public image to private registry
dagger call -m crane copy-image \
  --source-image docker.io/nginx:latest \
  --destination-image myregistry.com/nginx:latest \
  --destination-auth env:REGISTRY_AUTH
```

### Inspect Image Manifest

```bash
# Get image manifest
dagger call -m crane inspect-image \
  --image-ref alpine:latest \
  export --path=./manifest.json

cat manifest.json | jq .
```

### List Image Tags

```bash
# List all tags for an image
dagger call -m crane list-tags \
  --repository docker.io/library/nginx \
  export --path=./tags.txt

cat tags.txt
```

## API Reference

### Image Operations

```bash
# Copy image with authentication
dagger call -m crane copy-image \
  --source-image source.registry.com/app:v1.0 \
  --destination-image dest.registry.com/app:v1.0 \
  --source-auth env:SOURCE_REGISTRY_AUTH \
  --destination-auth env:DEST_REGISTRY_AUTH

# Copy multi-arch image
dagger call -m crane copy-image \
  --source-image docker.io/library/golang:1.21 \
  --destination-image myregistry.com/golang:1.21 \
  --all-platforms true \
  --destination-auth env:REGISTRY_AUTH
```

### Manifest Operations

```bash
# Get raw manifest
dagger call -m crane get-manifest \
  --image-ref nginx:alpine \
  export --path=./manifest.json

# Get config blob
dagger call -m crane get-config \
  --image-ref nginx:alpine \
  export --path=./config.json

# Get image digest
dagger call -m crane get-digest \
  --image-ref nginx:alpine
```

### Registry Operations

```bash
# List repository tags
dagger call -m crane list-tags \
  --repository gcr.io/my-project/my-app \
  --auth env:REGISTRY_AUTH \
  export --path=./all-tags.txt

# Delete image by tag
dagger call -m crane delete-image \
  --image-ref myregistry.com/app:old-version \
  --auth env:REGISTRY_AUTH

# Delete image by digest
dagger call -m crane delete-image \
  --image-ref myregistry.com/app@sha256:abc123... \
  --auth env:REGISTRY_AUTH
```

### Image Analysis

```bash
# Analyze image layers
dagger call -m crane analyze-layers \
  --image-ref node:18-alpine \
  export --path=./layers.json

# Get image size
dagger call -m crane image-size \
  --image-ref python:3.11-slim

# Compare image manifests
dagger call -m crane compare-images \
  --image-a ubuntu:20.04 \
  --image-b ubuntu:22.04 \
  export --path=./comparison.json
```

## Authentication

### Registry Authentication Methods

**Docker Config JSON:**
```bash
# Use existing Docker config
export REGISTRY_AUTH=$(cat ~/.docker/config.json | base64 -w 0)
```

**Username/Password:**
```bash
# Basic auth (base64 encoded)
export REGISTRY_AUTH=$(echo -n "username:password" | base64)
```

**Token-based:**
```bash
# For registries like GitHub Container Registry
export REGISTRY_AUTH="Bearer ghp_token123"
```

### Registry-specific Examples

**Docker Hub:**
```bash
export DOCKER_AUTH=$(echo -n "$DOCKER_USERNAME:$DOCKER_PASSWORD" | base64)
dagger call -m crane copy-image \
  --source-image nginx:latest \
  --destination-image $DOCKER_USERNAME/nginx:latest \
  --destination-auth env:DOCKER_AUTH
```

**GitHub Container Registry:**
```bash
export GHCR_AUTH="Bearer $GITHUB_TOKEN"
dagger call -m crane copy-image \
  --source-image nginx:latest \
  --destination-image ghcr.io/$GITHUB_USERNAME/nginx:latest \
  --destination-auth env:GHCR_AUTH
```

**Google Container Registry:**
```bash
# Use service account key
export GCR_AUTH="_json_key:$(cat service-account.json | base64 -w 0)"
dagger call -m crane copy-image \
  --source-image nginx:latest \
  --destination-image gcr.io/my-project/nginx:latest \
  --destination-auth env:GCR_AUTH
```

## Multi-Architecture Support

```bash
# Copy all platform variants
dagger call -m crane copy-image \
  --source-image docker.io/library/redis:7-alpine \
  --destination-image myregistry.com/redis:7-alpine \
  --all-platforms true \
  --destination-auth env:REGISTRY_AUTH

# List available platforms
dagger call -m crane list-platforms \
  --image-ref golang:1.21 \
  export --path=./platforms.txt
```

## Efficient Image Operations

Crane operates directly on registry APIs without requiring local Docker daemon:

- **No Local Storage**: Operations don't download images locally
- **Fast Transfers**: Direct registry-to-registry copying
- **Bandwidth Efficient**: Only transfers necessary layers
- **Concurrent**: Parallel layer operations

## Image Migration Workflows

### Registry Migration
```bash
#!/bin/bash
# Migrate all images from old registry to new one

# List all repositories
dagger call -m crane list-repositories \
  --registry old-registry.com \
  --auth env:OLD_REGISTRY_AUTH \
  export --path=./repositories.txt

# Copy each repository
while read repo; do
  echo "Migrating $repo..."

  # List tags for repository
  dagger call -m crane list-tags \
    --repository old-registry.com/$repo \
    --auth env:OLD_REGISTRY_AUTH \
    export --path=./tags-$repo.txt

  # Copy each tag
  while read tag; do
    dagger call -m crane copy-image \
      --source-image old-registry.com/$repo:$tag \
      --destination-image new-registry.com/$repo:$tag \
      --source-auth env:OLD_REGISTRY_AUTH \
      --destination-auth env:NEW_REGISTRY_AUTH \
      --all-platforms true
  done < ./tags-$repo.txt
done < ./repositories.txt
```

## Examples

See the [main README](../README.md#crane) for detailed usage examples.

## Resources

- [Crane Documentation](https://github.com/google/go-containerregistry/tree/main/cmd/crane)
- [Container Registry API](https://docs.docker.com/registry/spec/api/)
- [OCI Distribution Spec](https://github.com/opencontainers/distribution-spec)
- [Multi-Platform Images](https://docs.docker.com/build/building/multi-platform/)