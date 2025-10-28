# Git Dagger Module

This module provides comprehensive Git operations including repository management, branch operations, commit handling, and remote synchronization.

## Features

- ✅ Repository cloning and initialization

## Prerequisites

- Dagger CLI installed
- Git credentials configured
- SSH keys or personal access tokens (for private repositories)

## Quick Start

### Clone GitHUB Repository

```bash
# CLONE PRIVATE REPOSITORY WITH TOKEN
dagger call -m git clone-github \
--repository stuttgart-things/stuttgart-things \
--token env:GITHUB_TOKEN \
export --path=/tmp/private-repo
```

### Create GitHub Issue

```bash
dagger call -m git create-github-issue \
--repository stuttgart-things/stuttgart-things \
--token env:GITHUB_TOKEN \
--title "🧪 Test Issue from Dagger" \
--body "This issue was automatically created using Dagger!" \
--label automation \
--label test
```

## Examples

See the [main README](../README.md#git) for detailed usage examples.

## Resources

- [Git Documentation](https://git-scm.com/doc)
- [GitHub CLI](https://cli.github.com/)
- [GitLab CLI](https://gitlab.com/gitlab-org/cli)
- [Conventional Commits](https://conventionalcommits.org/)
- [Git Flow](https://nvie.com/posts/a-successful-git-branching-model/)
