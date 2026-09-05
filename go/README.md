# Go Dagger Module

This module provides Dagger functions for Go application development including building, testing, and container creation.

## Features

- ✅ Go project linting with golangci-lint
- ✅ Binary building with custom LDFLAGS and cross-compilation
- ✅ Ko-based container building and pushing
- ✅ Static security analysis with gosec
- ✅ Reachable-vulnerability scanning with govulncheck
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

```bash
# Build and push w/o auth
dagger call -m go ko-build \
  --src tests/go/calculator/ \
  --push=true \
  --repo=ttl.sh/test-calc:1h \
  --progress plain
```

### Release Binaries to GitHub

```bash
# CHECK ONLY + EXPORT RELEASER LOG
dagger call -m go release \
--src ~/projects/k2n \
--token=env:GITHUB_TOKEN \
--check-only=true \
--progress plain \
-vv export \
--path=/tmp/goreleaser.log
```

```bash
# RELEASE + EXPORT RELEASER LOG
dagger call -m go release \
--src ~/projects/k2n \
--token=env:GITHUB_TOKEN \
--progress plain \
-vv export \
--path=/tmp/goreleaser.log
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
  --go-version 1.25.5 \
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

`security-scan` runs gosec, a static analyser that looks for unsafe patterns in
code you wrote.

```bash
dagger call -m go security-scan \
  --src "." \
  --progress plain
```

### Vulnerability Scanning

`govulncheck` answers a different question: it reads the Go vulnerability
database and then does call-graph analysis, so it reports only vulnerabilities
on a path your binary can actually reach. A CVE in a dependency whose affected
function is never called stays quiet, which is what makes it usable as a build
gate instead of noise.

```bash
# Human-readable report
dagger call -m go govulncheck \
  --src "." \
  export --path=/tmp/govulncheck.txt
```

```bash
# SARIF, for GitHub code scanning
dagger call -m go govulncheck \
  --src "." \
  --format sarif \
  export --path=/tmp/govulncheck.sarif
```

```bash
# Fail the build on a reachable vulnerability
dagger call -m go govulncheck \
  --src "." \
  --fail-on-finding
```

| Flag | Default | Notes |
| --- | --- | --- |
| `--format` | `text` | `text`, `json`, `sarif` or `openvex` |
| `--govulncheck-version` | `v1.7.0` | Version installed via `go install` |
| `--go-version` | `1.25.5` | A floor, not a pin -- `GOTOOLCHAIN=auto`, so a `go.mod` requiring something newer wins |
| `--variant` | `alpine` | Base image variant |
| `--fail-on-finding` | `false` | Return an error when something reachable is found |
| `--no-cache` | `false` | Bypass Dagger's result cache; the vulnerability database moves independently of your source |

Two behaviours are worth knowing before wiring this into a workflow.

The report is written even when vulnerabilities are found, so it is never lost
at the moment it is most wanted. But with `--fail-on-finding` the function
returns an error, which means a trailing `export` never runs. To both keep the
report and gate the build, call twice -- the second call hits Dagger's cache
and costs nothing:

```bash
dagger call -m go govulncheck --src . --format sarif export --path=report.sarif
dagger call -m go govulncheck --src . --format sarif --fail-on-finding
```

Second, govulncheck only signals findings through its exit code in `text` mode;
in `json`, `sarif` and `openvex` it exits 0 and expects the caller to parse the
output. `--fail-on-finding` therefore runs an extra text-mode pass in the same
container to get a trustworthy verdict, so the gate behaves identically in
every format.

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
