# Git Dagger Module

This module provides comprehensive Git operations including repository management, branch operations, commit handling, and remote synchronization.

## Features

- ✅ Repository cloning and initialization
- ✅ Branch creation and management  
- ✅ Commit operations with signing
- ✅ Tag management
- ✅ Remote repository synchronization
- ✅ Merge and rebase operations
- ✅ Git hooks execution

## Prerequisites

- Dagger CLI installed
- Git credentials configured
- SSH keys or personal access tokens (for private repositories)

## Quick Start

### Clone Repository

```bash
# Clone public repository
dagger call -m git clone-repo \
  --url https://github.com/stuttgart-things/ansible.git \
  --destination ./cloned-repo

# Clone private repository with token
dagger call -m git clone-repo \
  --url https://github.com/private/repo.git \
  --token env:GITHUB_TOKEN \
  --destination ./private-repo
```

### Create and Push Branch

```bash
# Create new branch and commit changes
dagger call -m git create-branch \
  --src ./my-project \
  --branch-name feature/new-feature \
  --commit-message "Add new feature implementation"

# Push branch to remote
dagger call -m git push-branch \
  --src ./my-project \
  --branch-name feature/new-feature \
  --token env:GITHUB_TOKEN
```

### Tag Management

```bash
# Create and push tag
dagger call -m git create-tag \
  --src ./my-project \
  --tag-name v1.2.0 \
  --message "Release version 1.2.0" \
  --push true \
  --token env:GITHUB_TOKEN
```

## API Reference

### Repository Operations

```bash
# Initialize new repository
dagger call -m git init-repo \
  --destination ./new-project \
  --initial-commit "Initial commit"

# Clone with specific branch
dagger call -m git clone-repo \
  --url https://github.com/user/repo.git \
  --branch develop \
  --depth 1 \
  --destination ./shallow-clone

# Mirror repository
dagger call -m git mirror-repo \
  --source-url https://github.com/source/repo.git \
  --destination-url https://github.com/destination/repo.git \
  --token env:GITHUB_TOKEN
```

### Branch Management

```bash
# List branches
dagger call -m git list-branches \
  --src ./project \
  --include-remote true

# Switch branch
dagger call -m git checkout-branch \
  --src ./project \
  --branch-name develop

# Merge branches
dagger call -m git merge-branch \
  --src ./project \
  --source-branch feature/xyz \
  --target-branch main \
  --strategy merge  # or rebase, squash
```

### Commit Operations

```bash
# Create commit with signing
dagger call -m git commit-changes \
  --src ./project \
  --message "feat: implement user authentication" \
  --author "John Doe <john@example.com>" \
  --sign true \
  --gpg-key env:GPG_SIGNING_KEY

# Amend last commit
dagger call -m git amend-commit \
  --src ./project \
  --message "feat: implement user authentication (updated)"

# Cherry-pick commit
dagger call -m git cherry-pick \
  --src ./project \
  --commit-sha abc123def456 \
  --target-branch main
```

### Tag Operations

```bash
# List tags
dagger call -m git list-tags \
  --src ./project \
  --pattern "v*" \
  --sort-by date

# Create annotated tag
dagger call -m git create-tag \
  --src ./project \
  --tag-name v2.0.0 \
  --message "Major release with breaking changes" \
  --annotated true \
  --sign true

# Delete tag
dagger call -m git delete-tag \
  --src ./project \
  --tag-name v1.9.9 \
  --remote true
```

### Remote Operations

```bash
# Add remote
dagger call -m git add-remote \
  --src ./project \
  --name upstream \
  --url https://github.com/upstream/repo.git

# Fetch from remote
dagger call -m git fetch-remote \
  --src ./project \
  --remote origin \
  --prune true

# Sync with upstream
dagger call -m git sync-upstream \
  --src ./project \
  --upstream-remote upstream \
  --target-branch main
```

## Authentication Methods

### Personal Access Token
```bash
export GITHUB_TOKEN="ghp_xxx"
dagger call -m git clone-repo \
  --url https://github.com/private/repo.git \
  --token env:GITHUB_TOKEN
```

### SSH Key Authentication
```bash
export SSH_PRIVATE_KEY=$(cat ~/.ssh/id_rsa)
dagger call -m git clone-repo \
  --url git@github.com:private/repo.git \
  --ssh-key env:SSH_PRIVATE_KEY
```

### Username/Password
```bash
dagger call -m git clone-repo \
  --url https://github.com/private/repo.git \
  --username myuser \
  --password env:PASSWORD
```

## Git Hooks Integration

```bash
# Execute pre-commit hooks
dagger call -m git run-hook \
  --src ./project \
  --hook-type pre-commit \
  --args "--all-files"

# Validate commit message
dagger call -m git validate-commit \
  --src ./project \
  --commit-sha HEAD \
  --conventional-commits true
```

## Workflow Examples

### Feature Branch Workflow
```bash
# 1. Create feature branch
dagger call -m git create-branch \
  --src ./project \
  --branch-name feature/auth-system

# 2. Make changes and commit
dagger call -m git commit-changes \
  --src ./project \
  --message "feat: add user authentication"

# 3. Push and create PR
dagger call -m git push-branch \
  --src ./project \
  --branch-name feature/auth-system \
  --token env:GITHUB_TOKEN
```

### Release Preparation
```bash
# 1. Update version and changelog
dagger call -m git commit-changes \
  --src ./project \
  --message "chore: prepare release v2.1.0"

# 2. Create release tag
dagger call -m git create-tag \
  --src ./project \
  --tag-name v2.1.0 \
  --message "Release v2.1.0" \
  --push true

# 3. Merge to main
dagger call -m git merge-branch \
  --src ./project \
  --source-branch release/v2.1.0 \
  --target-branch main
```

## Examples

See the [main README](../README.md#git) for detailed usage examples.

## Resources

- [Git Documentation](https://git-scm.com/doc)
- [GitHub CLI](https://cli.github.com/)
- [GitLab CLI](https://gitlab.com/gitlab-org/cli)
- [Conventional Commits](https://conventionalcommits.org/)
- [Git Flow](https://nvie.com/posts/a-successful-git-branching-model/)