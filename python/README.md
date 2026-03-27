# Python Dagger Module

This module provides Dagger functions for Python CI/CD workflows including linting, testing, security scanning, and Docker image building.

## Features

- ✅ Python linting with ruff
- ✅ Format checking with ruff
- ✅ Testing with pytest
- ✅ Test coverage reporting with pytest-cov
- ✅ Security scanning with bandit
- ✅ Docker image building and pushing

## Prerequisites

- Dagger CLI installed
- Docker runtime available

## Quick Start

### Lint

```bash
# Lint Python source code
dagger call -m python lint \
  --src tests/python/ \
  --progress plain
```

### Format Check

```bash
# Check Python code formatting
dagger call -m python format-check \
  --src tests/python/ \
  --progress plain
```

### Test

```bash
# Run pytest
dagger call -m python test \
  --src tests/python/ \
  --progress plain
```

### Test with Coverage

```bash
# Run pytest with coverage reporting
dagger call -m python test-with-coverage \
  --src tests/python/ \
  --coverage-path src/ \
  --progress plain
```

### Security Scan

```bash
# Run bandit security scanner
dagger call -m python security-scan \
  --src tests/python/ \
  --scan-path src/ \
  --progress plain
```

### Build Image

```bash
# Build Docker image from Dockerfile
dagger call -m python build-image \
  --src tests/python/ \
  --progress plain
```

### Build and Push Image

```bash
# Build and push to temporary registry (no auth)
dagger call -m python build-and-push-image \
  --src tests/python/ \
  --image-ref ttl.sh/stuttgart-things/python-test:1h \
  --token env:GITHUB_TOKEN \
  --progress plain

# Build and push to GitHub Container Registry (with auth)
dagger call -m python build-and-push-image \
  --src tests/python/ \
  --image-ref ghcr.io/stuttgart-things/python-test:1.0.0 \
  --token env:GITHUB_TOKEN \
  --username patrick-hermann-sva \
  --registry-url ghcr.io \
  --progress plain
```

### Test Module

```bash
# Run comprehensive tests
task test-python
```

## API Reference

### Linting

```bash
dagger call -m python lint \
  --src "." \
  --ruff-version 0.8.6 \
  --paths "src/,tests/" \
  --progress plain
```

### Format Checking

```bash
dagger call -m python format-check \
  --src "." \
  --ruff-version 0.8.6 \
  --paths "src/,tests/" \
  --progress plain
```

### Testing

```bash
dagger call -m python test \
  --src "." \
  --python-version 3.12-slim \
  --test-path tests/ \
  --install-extra ".[dev]" \
  --progress plain
```

### Test with Coverage

```bash
dagger call -m python test-with-coverage \
  --src "." \
  --python-version 3.12-slim \
  --test-path tests/ \
  --coverage-path src/ \
  --progress plain
```

### Security Scanning

```bash
dagger call -m python security-scan \
  --src "." \
  --bandit-version 1.8.3 \
  --scan-path src/ \
  --progress plain
```

### Image Building

```bash
dagger call -m python build-image \
  --src "." \
  --dockerfile Dockerfile \
  --version 1.0.0 \
  --progress plain
```

### Build and Push

```bash
dagger call -m python build-and-push-image \
  --src "." \
  --image-ref ghcr.io/org/image:tag \
  --token env:REGISTRY_TOKEN \
  --username env:REGISTRY_USER \
  --registry-url ghcr.io \
  --progress plain
```

## Examples

See the [main README](../README.md#python) for detailed usage examples.

## Testing

```bash
task test-python
```

## Resources

- [Ruff Documentation](https://docs.astral.sh/ruff/)
- [Pytest Documentation](https://docs.pytest.org/)
- [Bandit Documentation](https://bandit.readthedocs.io/)
- [Python Packaging Guide](https://packaging.python.org/)
