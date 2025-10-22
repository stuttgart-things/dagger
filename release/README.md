# Release Dagger Module

This module provides Dagger functions for release management including semantic versioning, changelog generation, and release automation.

## Features

- ✅ Semantic version calculation
- ✅ Changelog generation from commits
- ✅ Git tag management
- ✅ GitHub/GitLab release creation
- ✅ Release asset uploading
- ✅ Multi-format output (JSON, Markdown, YAML)

## Prerequisites

- Dagger CLI installed
- Git repository with commit history
- GitHub/GitLab API tokens (for release creation)
- Conventional commit format recommended

## Quick Start

### Calculate Next Version

```bash
# Calculate semantic version from git history
dagger call -m release next-version \
  --src ./my-project \
  --current-version v1.2.3

# Output: v1.3.0 (for minor changes)
```

### Generate Changelog

```bash
# Generate changelog from git commits
dagger call -m release generate-changelog \
  --src ./my-project \
  --from-version v1.2.3 \
  --to-version v1.3.0 \
  export --path=./CHANGELOG.md
```

### Create Release

```bash
# Create GitHub release
export GITHUB_TOKEN="ghp_xxx"
dagger call -m release create-github-release \
  --src ./my-project \
  --version v1.3.0 \
  --github-token env:GITHUB_TOKEN \
  --repository stuttgart-things/my-project
```

## API Reference

### Version Management

```bash
# Calculate next version with type
dagger call -m release next-version \
  --src ./project \
  --current-version v1.2.3 \
  --release-type patch  # major, minor, patch

# Get current version from git tags
dagger call -m release current-version \
  --src ./project
```

### Changelog Operations

```bash
# Full changelog generation
dagger call -m release generate-changelog \
  --src ./project \
  --from-version v1.0.0 \
  --to-version v2.0.0 \
  --format markdown \
  export --path=./CHANGELOG.md

# Incremental changelog
dagger call -m release changelog-since \
  --src ./project \
  --since-version v1.2.3 \
  export --path=./CHANGELOG-latest.md
```

### Release Creation

```bash
# GitHub release with assets
dagger call -m release create-github-release \
  --src ./project \
  --version v1.3.0 \
  --github-token env:GITHUB_TOKEN \
  --repository owner/repo \
  --release-notes ./CHANGELOG-latest.md \
  --asset ./dist/binary-linux-amd64 \
  --asset ./dist/binary-darwin-amd64

# GitLab release
dagger call -m release create-gitlab-release \
  --src ./project \
  --version v1.3.0 \
  --gitlab-token env:GITLAB_TOKEN \
  --project-id 12345
```

## Conventional Commits

The module works best with [Conventional Commits](https://conventionalcommits.org/):

```bash
# Examples that trigger version bumps:
feat: add new API endpoint     # → minor version bump
fix: resolve memory leak       # → patch version bump
feat!: breaking API change     # → major version bump
docs: update README           # → no version bump
```

## Release Types

### Patch Release (1.2.3 → 1.2.4)
- Bug fixes
- Documentation updates
- Internal refactoring

### Minor Release (1.2.3 → 1.3.0)
- New features (backward compatible)
- Deprecations
- Performance improvements

### Major Release (1.2.3 → 2.0.0)
- Breaking changes
- API modifications
- Architecture changes

## Changelog Format

Generated changelogs follow Keep a Changelog format:

```markdown
# Changelog

## [1.3.0] - 2024-01-15

### Added
- New authentication system
- User profile management

### Changed
- Improved error handling
- Updated dependencies

### Fixed
- Memory leak in worker process
- Race condition in cache

### Deprecated
- Old API endpoints (will be removed in v2.0.0)

### Security
- Fixed XSS vulnerability
```

## Automation Workflows

### CI/CD Integration

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    branches: [main]

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Calculate Version
        run: |
          VERSION=$(dagger call -m release next-version --src .)
          echo "VERSION=$VERSION" >> $GITHUB_ENV

      - name: Generate Changelog
        run: |
          dagger call -m release generate-changelog \
            --src . --to-version $VERSION \
            export --path=./CHANGELOG.md

      - name: Create Release
        run: |
          dagger call -m release create-github-release \
            --src . --version $VERSION \
            --github-token ${{ secrets.GITHUB_TOKEN }} \
            --repository ${{ github.repository }}
```

## Examples

See the [main README](../README.md#release) for detailed usage examples.

## Resources

- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://conventionalcommits.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [GitHub Releases API](https://docs.github.com/en/rest/releases)
- [GitLab Releases API](https://docs.gitlab.com/ee/api/releases/)