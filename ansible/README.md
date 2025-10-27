# Ansible Dagger Module

This module provides Dagger functions for Ansible automation including playbook execution and collection building.

## Features

- ✅ Execute Ansible playbooks with inventory and requirements
- ✅ Build and package Ansible collections
- ✅ Create GitHub releases with collection artifacts
- ✅ Support for SSH authentication and Vault integration
- ✅ Flexible parameter and variable management

## Prerequisites

- Dagger CLI installed
- Docker runtime available

## Quick Start

### Execute Playbooks

```bash
# Basic playbook execution
dagger call -m ansible execute \
  --src . \
  --playbooks tests/ansible/hello.yaml,tests/ansible/hello2.yaml \
  -vv --progress plain
```

### Build Collection

```bash
# Build Ansible collection package
dagger call -m ansible run-collection-build-pipeline \
  --src ansible/collections/baseos \
  --progress plain \
  export --path=/tmp/ansible/output/
```

### Test Module

```bash
# Run comprehensive tests
task test-ansible
```

## API Reference

### Execute Ansible Playbooks

```bash
dagger call -m ansible execute \
  --requirements tests/ansible/requirements.yaml \
  --src . \
  --playbooks tests/ansible/hello.yaml,tests/ansible/hello2.yaml \
  --inventory /path/to/inventory \
  --ssh-user env:SSH_USER \
  --ssh-password env:SSH_PASSWORD \
  --parameters "send_to_homerun=false" \
  -vv --progress plain
```

### Collection Building

```bash
dagger call -m ansible run-collection-build-pipeline \
  --src tests/ansible/collection \
  --progress plain \
  export --path=/tmp/ansible
```

### GitHub Release Creation

```bash
dagger call -m ansible github-release \
  --token env:GITHUB_TOKEN \
  --group stuttgart-things \
  --repo dagger \
  --files "tests/test-values.yaml,tests/registry/README.md" \
  --notes "test" \
  --tag 09.1.6 \
  --title hello
```

## Examples

See the [main README](../README.md#ansible) for detailed usage examples.

## Testing

```bash
task test-ansible
```

## Resources

- [Ansible Documentation](https://docs.ansible.com/)
- [Ansible Galaxy](https://galaxy.ansible.com/)
- [Stuttgart-Things Ansible Collections](https://github.com/stuttgart-things/ansible)
