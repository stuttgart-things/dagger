# Trivy Dagger Module

This module provides Dagger functions for Trivy security scanning including filesystem scans, container image scans, and vulnerability reporting.

## Features

- ✅ Filesystem vulnerability scanning (local and remote)
- ✅ Container image security scanning
- ✅ Registry authentication support
- ✅ Git repository scanning
- ✅ JSON report generation
- ✅ Multi-format output support

## Prerequisites

- Dagger CLI installed
- Docker runtime available
- Registry credentials (for private images)

## Quick Start

### Filesystem Scan (Local)

```bash
# Scan local filesystem
dagger call -m trivy scan-filesystem \
  --src /home/user/projects/myproject \
  --progress plain \
  export --path=/tmp/trivy-fs.json

cat /tmp/trivy-fs.json
```

### Filesystem Scan (Remote Git)

```bash
# Scan remote Git repository
dagger call -m trivy scan-filesystem \
  --src git://github.com/stuttgart-things/ansible.git \
  --progress plain \
  export --path=/tmp/trivy-fs.json

cat /tmp/trivy-fs.json
```

### Image Scan

```bash
# Scan container image with registry auth
export REG_USER="myuser"
export REG_PW="mypassword"

dagger call -m trivy scan-image \
  --image-ref nginx:latest \
  --registry-user env:REG_USER \
  --registry-password env:REG_PW \
  --progress plain \
  export --path=/tmp/image-nginx.json

cat /tmp/image-nginx.json
```

## API Reference

### Filesystem Scanning

```bash
# Local filesystem scan
dagger call -m trivy scan-filesystem \
  --src ./project-directory \
  export --path=/tmp/scan-results.json

# Remote Git repository scan
dagger call -m trivy scan-filesystem \
  --src git://github.com/user/repo.git \
  export --path=/tmp/scan-results.json
```

### Container Image Scanning

```bash
# Public image scan
dagger call -m trivy scan-image \
  --image-ref alpine:latest \
  export --path=/tmp/image-scan.json

# Private image scan with authentication
dagger call -m trivy scan-image \
  --image-ref myregistry.com/private/image:tag \
  --registry-user env:REGISTRY_USER \
  --registry-password env:REGISTRY_PASSWORD \
  export --path=/tmp/private-scan.json
```

## Report Analysis

Trivy generates comprehensive JSON reports including:

- **Vulnerabilities**: CVE details with severity levels
- **Package Information**: Affected packages and versions
- **Fix Information**: Available patches and updates
- **Metadata**: Scan timestamp and configuration

**Example Report Structure:**
```json
{
  "SchemaVersion": 2,
  "ArtifactName": "nginx:latest",
  "ArtifactType": "container_image",
  "Results": [
    {
      "Target": "nginx:latest (debian 11.6)",
      "Class": "os-pkgs",
      "Type": "debian",
      "Vulnerabilities": [
        {
          "VulnerabilityID": "CVE-2023-1234",
          "PkgName": "libssl1.1",
          "Severity": "HIGH",
          "FixedVersion": "1.1.1n-0+deb11u4"
        }
      ]
    }
  ]
}
```

## Supported Registries

- **Docker Hub** (`docker.io`)
- **GitHub Container Registry** (`ghcr.io`)
- **AWS ECR**
- **Google Container Registry** (`gcr.io`)
- **Harbor** (custom registries)
- **Azure Container Registry** (`*.azurecr.io`)

## Examples

See the [main README](../README.md#trivy) for detailed usage examples.

## Resources

- [Trivy Documentation](https://trivy.dev/)
- [Vulnerability Database](https://github.com/aquasecurity/trivy-db)
- [Security Best Practices](https://trivy.dev/latest/docs/)