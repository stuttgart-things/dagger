# Flux Dagger Module

A Dagger module for building and pushing OCI artifacts to container registries using the [Flux CLI](https://fluxcd.io/flux/cmd/flux_push_artifact/). Designed for Flux GitOps workflows where Kubernetes manifests, Kustomize overlays, or other configuration are stored as OCI artifacts.

## Features

- Build and push OCI artifacts from directories
- Batch-push multiple subdirectories as individual OCI artifacts
- Registry authentication via Docker config.json
- Optional source and revision metadata for Flux provenance tracking
- No local Docker daemon required

## Prerequisites

- Dagger CLI installed
- Registry credentials (for private registries)

## Module Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `base-image` | string | `cgr.dev/chainguard/wolfi-base:latest` | Base image for the Flux container |

## Functions

### PushArtifact

Builds and pushes a single OCI artifact from a directory.

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `src` | Directory | Yes | - | Source directory containing the artifact files |
| `artifact` | string | Yes | - | OCI artifact address (e.g., `oci://ghcr.io/org/repo:tag`) |
| `registry` | string | Yes | - | Registry URL for authentication (e.g., `ghcr.io`) |
| `username` | string | Yes | - | Registry username |
| `password` | Secret | Yes | - | Registry password |
| `source` | string | No | - | Source URL metadata (e.g., git remote URL) |
| `revision` | string | No | - | Revision metadata (e.g., `main@sha1:abc123`) |

### PushArtifacts

Builds and pushes OCI artifacts from multiple subdirectories. Each subdirectory name is appended to the base artifact address (e.g., `oci://ghcr.io/org/repo/subdir-name:tag`).

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `src` | Directory | Yes | - | Source directory containing subdirectories |
| `artifact` | string | Yes | - | Base OCI artifact address without tag (e.g., `oci://ghcr.io/org/repo`) |
| `tag` | string | Yes | - | Tag to use for all artifacts |
| `registry` | string | Yes | - | Registry URL for authentication (e.g., `ghcr.io`) |
| `username` | string | Yes | - | Registry username |
| `password` | Secret | Yes | - | Registry password |
| `source` | string | No | - | Source URL metadata (e.g., git remote URL) |
| `revision` | string | No | - | Revision metadata (e.g., `main@sha1:abc123`) |

## Usage Examples

### Push a Single Artifact

```bash
# Push Kubernetes manifests as an OCI artifact
dagger call -m flux push-artifact \
  --src=./manifests \
  --artifact="oci://ghcr.io/myorg/manifests:v1.0.0" \
  --registry="ghcr.io" \
  --username="myusername" \
  --password=env:GITHUB_TOKEN
```

### Push with Source and Revision Metadata

```bash
# Include provenance metadata for Flux
dagger call -m flux push-artifact \
  --src=./manifests \
  --artifact="oci://ghcr.io/myorg/manifests:v1.0.0" \
  --registry="ghcr.io" \
  --username="myusername" \
  --password=env:GITHUB_TOKEN \
  --source="https://github.com/myorg/my-repo" \
  --revision="main@sha1:abc123def456"
```

### Push Multiple Artifacts from Subdirectories

Given a directory structure like:

```
artifacts/
├── app-config/
│   ├── deployment.yaml
│   └── service.yaml
├── monitoring/
│   ├── prometheus.yaml
│   └── grafana.yaml
└── ingress/
    └── ingress.yaml
```

Each subdirectory becomes its own OCI artifact:

```bash
dagger call -m flux push-artifacts \
  --src=./artifacts \
  --artifact="oci://ghcr.io/myorg/cluster-config" \
  --tag="v1.0.0" \
  --registry="ghcr.io" \
  --username="myusername" \
  --password=env:GITHUB_TOKEN
```

This pushes three artifacts:
- `oci://ghcr.io/myorg/cluster-config/app-config:v1.0.0`
- `oci://ghcr.io/myorg/cluster-config/monitoring:v1.0.0`
- `oci://ghcr.io/myorg/cluster-config/ingress:v1.0.0`

## Registry-Specific Examples

### GitHub Container Registry (ghcr.io)

```bash
dagger call -m flux push-artifact \
  --src=./manifests \
  --artifact="oci://ghcr.io/myorg/manifests:latest" \
  --registry="ghcr.io" \
  --username="USERNAME" \
  --password=env:GITHUB_TOKEN
```

### Harbor

```bash
dagger call -m flux push-artifact \
  --src=./manifests \
  --artifact="oci://harbor.example.com/project/manifests:v1.0.0" \
  --registry="harbor.example.com" \
  --username="admin" \
  --password=env:HARBOR_PASSWORD
```

### Docker Hub

```bash
dagger call -m flux push-artifact \
  --src=./manifests \
  --artifact="oci://docker.io/myorg/manifests:v1.0.0" \
  --registry="docker.io" \
  --username="USERNAME" \
  --password=env:DOCKER_PASSWORD
```

## How It Works

1. Creates a Wolfi-based container with Flux CLI v2.4.0 installed
2. Generates a Docker `config.json` with registry credentials
3. Mounts the source directory into the container
4. Executes `flux push artifact` with the specified OCI address and options
5. Returns the command output

For `PushArtifacts`, the module iterates over each subdirectory in the source and pushes each as a separate OCI artifact, appending the subdirectory name to the base artifact address.

## Resources

- [Flux OCI Documentation](https://fluxcd.io/flux/cheatsheets/oci-artifacts/)
- [Flux Push Artifact Command](https://fluxcd.io/flux/cmd/flux_push_artifact/)
- [Dagger Documentation](https://docs.dagger.io/)
- [OCI Distribution Spec](https://github.com/opencontainers/distribution-spec)
