# Go Dagger Module

This module provides Dagger functions for Go application development including building, testing, and container creation.

## Features

- ✅ Go project linting with golangci-lint
- ✅ Binary building with custom LDFLAGS and cross-compilation
- ✅ Ko-based container building and pushing
- ✅ Security scanning with vulnerability detection
- ✅ Flexible Go version and build configuration

## Prerequisites

- Dagger CLI installed
- Docker runtime available

## Quick Start

### Lint Project

```bash
# Lint Go project
dagger call -m go lint \
  --src "." \
  --timeout 300s \
  --progress plain -vv
```

### Build Binary

```bash
# Build Go binary with custom flags
dagger call -m go build-binary \
  --src "." \
  --os linux \
  --arch amd64 \
  --ldflags "cmd.version=1.278910; cmd.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --go-main-file main.go \
  --bin-name myapp \
  export --path=/tmp/go/build/ \
  --progress plain -vv
```

### Ko Build Container

```bash
# Build and push container with Ko
dagger call -m go ko-build \
  --src tests/go/calculator/ \
  --token env:GITHUB_TOKEN \
  --repo ghcr.io/stuttgart-things/myapp \
  --progress plain -vv
```

### Test Module

```bash
# Run comprehensive tests
task test-go
```

## API Reference

### Linting

```bash
dagger call -m go lint \
  --src "." \
  --timeout 300s \
  --progress plain
```

### Binary Building

```bash
dagger call -m go build-binary \
  --src "." \
  --os linux \
  --arch amd64 \
  --package-name github.com/stuttgart-things/myapp \
  --go-main-file main.go \
  --bin-name myapp \
  --go-version 1.24.4 \
  export --path=/tmp/go/build/
```

### Ko Container Building

```bash
# Local build only
dagger call -m go ko-build \
  --src tests/go/calculator/ \
  --push false \
  --progress plain

# Build and push
dagger call -m go ko-build \
  --src tests/go/calculator/ \
  --token env:GITHUB_TOKEN \
  --repo ghcr.io/stuttgart-things/myapp \
  --progress plain
```

### Security Scanning

```bash
dagger call -m go security-scan \
  --src "." \
  --progress plain
```

## Examples

See the [main README](../README.md#go) for detailed usage examples.

## Testing

```bash
task test-go
```

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Ko Documentation](https://ko.build/)
- [golangci-lint](https://golangci-lint.run/)
