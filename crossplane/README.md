# Crossplane Dagger Module

This module provides Dagger functions for Crossplane package management including initialization, building, and publishing to OCI registries.

## Features

- ✅ Crossplane package initialization and scaffolding
- ✅ Package building and validation
- ✅ OCI registry publishing with authentication
- ✅ Custom package type creation
- ✅ Multi-registry support (GitHub, Harbor, etc.)
- ✅ Package dependency management

## Prerequisites

- Dagger CLI installed
- Docker runtime available
- Registry credentials for publishing

## Quick Start

### Initialize Package

```bash
# Initialize new Crossplane package
dagger call -m crossplane init-package \
  --name mypackage \
  --progress plain \
  export --path=/tmp/crossplane
```

### Build Package

```bash
# Build existing package
dagger call -m crossplane package \
  --src tests/registry \
  --progress plain \
  export --path=/tmp/crossplane/built
```

### Push to Registry

```bash
# Push package to GitHub Container Registry
dagger call -m crossplane push \
  --src tests/registry \
  --username stuttgart-things \
  --password env:GITHUB_TOKEN \
  --registry ghcr.io \
  --destination ghcr.io/stuttgart-things/xplane-registry \
  --progress plain
```

### Test Module

```bash
# Run comprehensive tests
task test-crossplane
```

## API Reference

### Package Initialization

```bash
dagger call -m crossplane init-package \
  --name mypackage \
  export --path=/tmp/crossplane/
```

### Package Building

```bash
dagger call -m crossplane package \
  --src ./crossplane-package \
  export --path=/tmp/built-package/
```

### Registry Publishing

```bash
dagger call -m crossplane push \
  --src ./crossplane-package \
  --username myuser \
  --password env:REGISTRY_TOKEN \
  --registry ghcr.io \
  --destination ghcr.io/myorg/mypackage:v1.0.0 \
  --progress plain
```

### Custom Package Creation

```bash
dagger call -m crossplane init-custom-package \
  --name mycustompackage \
  export --path=/tmp/crossplane/custom/
```

## Package Structure

A typical Crossplane package structure:

```
crossplane-package/
├── crossplane.yaml     # Package metadata
├── compositions/       # Composition resources
│   └── *.yaml
├── functions/          # Composition functions
│   └── *.yaml
└── examples/          # Usage examples
    └── *.yaml
```

## Registry Support

- **GitHub Container Registry** (`ghcr.io`)
- **Harbor** (custom registry)
- **Docker Hub** (`docker.io`)
- **AWS ECR**
- **Google Container Registry** (`gcr.io`)

## Examples

See the [main README](../README.md#crossplane) for detailed usage examples.

## Testing

```bash
task test-crossplane
```

## Resources

- [Crossplane Documentation](https://docs.crossplane.io/)
- [Crossplane Package Format](https://docs.crossplane.io/latest/concepts/packages/)
- [Crossplane Marketplace](https://marketplace.upbound.io/)