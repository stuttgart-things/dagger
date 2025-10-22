# GitLab Dagger Module

This module provides Dagger functions for GitLab API operations including project management, merge request handling, and repository cloning.

## Features

- ✅ Project ID resolution by name and group
- ✅ Merge request management and operations
- ✅ Repository cloning with authentication
- ✅ File change tracking and analysis
- ✅ Project listing by group
- ✅ Merge request state management

## Prerequisites

- Dagger CLI installed
- Docker runtime available
- GitLab API token

## Quick Start

### Get Project ID

```bash
# Find project ID by name and group
dagger call -m gitlab get-project-id \
  --token env:GITLAB_TOKEN \
  --server gitlab.com \
  --project-name "docs" \
  --group-path "Lab/stuttgart-things/idp"
```

### List Projects

```bash
# List all projects in a group
dagger call -m gitlab list-projects \
  --server gitlab.com \
  --token env:GITLAB_TOKEN \
  --group-path "Lab%2Fstuttgart-things" \
  --progress plain
```

### Clone Repository

```bash
# Clone GitLab repository
dagger call -m gitlab clone \
  --repo-url https://gitlab.com/Lab/stuttgart-things/idp/docs.git \
  --token env:GITLAB_TOKEN \
  --branch main \
  export --path /tmp/repo
```

### Test Module

```bash
# Run comprehensive tests
task test-gitlab
```

## API Reference

### Project Operations

```bash
# Get project ID
dagger call -m gitlab get-project-id \
  --token env:GITLAB_TOKEN \
  --server gitlab.com \
  --project-name "myproject" \
  --group-path "mygroup/subgroup"

# List group projects
dagger call -m gitlab list-projects \
  --server gitlab.com \
  --token env:GITLAB_TOKEN \
  --group-path "mygroup%2Fsubgroup"
```

### Merge Request Operations

```bash
# List merge requests
dagger call -m gitlab list-merge-requests \
  --token env:GITLAB_TOKEN \
  --server gitlab.com \
  --project-id 14160 \
  --progress plain

# Get specific merge request ID
dagger call -m gitlab get-merge-request-id \
  --token env:GITLAB_TOKEN \
  --server gitlab.com \
  --project-id 14466 \
  --merge-request-title "Feature: Add new functionality" \
  --progress plain

# List changes in merge request
dagger call -m gitlab list-merge-request-changes \
  --token env:GITLAB_TOKEN \
  --server gitlab.com \
  --project-id 14466 \
  --merge-request-id 1 \
  --progress plain

# Update merge request state
dagger call -m gitlab update-merge-request-state \
  --server gitlab.com \
  --token env:GITLAB_TOKEN \
  --merge-request-id 1 \
  --project-id 14466 \
  --action merge \
  --progress plain
```

### File Analysis

```bash
# Print files changed by merge request
dagger call -m gitlab print-merge-request-file-changes \
  --repo-url https://gitlab.com/group/project.git \
  --server gitlab.com \
  --token env:GITLAB_TOKEN \
  --merge-request-id 1 \
  --project-id 14466 \
  --branch "feature-branch" \
  --progress plain
```

### Repository Cloning

```bash
dagger call -m gitlab clone \
  --repo-url https://gitlab.com/group/project.git \
  --token env:GITLAB_TOKEN \
  --branch main \
  export --path /tmp/repo
```

## Examples

See the [main README](../README.md#gitlab) for detailed usage examples.

## Testing

```bash
task test-gitlab
```

## Resources

- [GitLab API Documentation](https://docs.gitlab.com/ee/api/)
- [GitLab Authentication](https://docs.gitlab.com/ee/api/#authentication)