# Crossplane Dagger Module

This module provides Dagger functions for Crossplane package management including initialization, building, and publishing to OCI registries.

## Features

- ✅ Crossplane package initialization and scaffolding
- ✅ Package building and validation
- ✅ Offline three-layer verification (XRD ↔ XR, provider CRD, embedded manifests)
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

### Verify a Configuration

Offline check of a single Configuration directory before push/PR-merge. Runs
`crossplane xpkg build`, validates each `examples/xr*.yaml` against the
Configuration's own XRD, renders the composition, and runs `kubeconform`
against both the rendered `Object` wrappers and the manifests embedded under
`spec.forProvider.manifest`. Exits non-zero on any failure.

```bash
dagger call -m crossplane verify \
  --src ./k8s/cloud-config

# Pin provider-kubernetes CRD schemas to match dependsOn in crossplane.yaml
dagger call -m crossplane verify \
  --src ./k8s/cloud-config \
  --provider-kubernetes-version v1.2.0
```

Sample output:

```
Configuration: cloud-config
  ✓ xpkg build
  ✓ xr-min.yaml: XRD-valid, render-ok, object-valid, embedded-valid
  ✓ xr.yaml:     XRD-valid, render-ok, object-valid, embedded-valid
  ✗ xr-max.yaml: XRD-valid, render-ok, object-valid, embedded-INVALID
      Secret/max-cloud-init.stringData.userdata: invalid YAML at line 42
```

A docker-in-docker service runs inside the call so `crossplane render` can
pull and execute composition functions without needing a Docker socket on the
host. No Kubernetes cluster is required.

### Test Module

```bash
# Run comprehensive tests
task test-crossplane
```

## API Reference

### Create Definition

```bash
# Initialize new defintion by kcl file
dagger call -m crossplane init-package \
  --name mypackage \
  --progress plain \
  export --path=/tmp/crossplane
```

### Modify Definition

```bash
# Modify xrd for api version2
dagger call -m crossplane modify-definition \
  --xrd /tmp/schema.yaml \
  --singular-name=cloudinit \
  --progress plain -vv \
  export --path=/tmp/xrd-v2.yaml
```

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
