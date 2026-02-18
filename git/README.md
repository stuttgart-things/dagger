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

### Delete GitHub Branch

```bash
dagger call -m git delete-github-branch \
--repository="stuttgart-things/dagger" \
--branch="feature-branch2" \
--token=env:GITHUB_TOKEN
```

### Create GitHub Branch

```bash
dagger call -m git create-github-branch \
--repository="stuttgart-things/dagger" \
--new-branch="feature-branch" \
--base-branch="main" \
--token=env:GITHUB_TOKEN
```

### Add file to GitHub Branch

```bash
# BRANCH MUST EXISTS
dagger call -m git add-file-to-github-branch \
--repository="stuttgart-things/dagger" \
--branch="feature-branch" \
--commit-message="Update README" \
--token=env:GITHUB_TOKEN \
--source-file=./README.md \
--destination-path="README-updated.md"
```

### Add folder to GitHub Branch

```bash
dagger call -m git add-folder-to-github-branch \
--repository="stuttgart-things/dagger" \
--branch="feature-branch2" \
--commit-message="Update docs" \
--token=env:GITHUB_TOKEN \
--source-dir=./git/internal \
--destination-path="whatever"
```

### Create github PullRequest

```bash
dagger call -m git create-github-pull-request \
--repository="stuttgart-things/dagger" \
--head-branch="feature-branch2" \
--title="Add new feature" \
--body="This PR adds a new feature" \
--labels="enhancement" \
--labels="documentation" \
--reviewers="patrick-hermann-sva" \
--token=env:GITHUB_TOKEN
```

### LIST GITHUB WORKFLOW RUNS

Returns a formatted table with columns: ID, WORKFLOW, STATUS, CONCLUSION, BRANCH, EVENT, TITLE, CREATED.

```bash
# LIST ALL WORKFLOW RUNS
dagger call -m git list-github-workflow-runs \
--repository="stuttgart-things/dagger" \
--token=env:GITHUB_TOKEN

# FILTER BY BRANCH AND STATUS
dagger call -m git list-github-workflow-runs \
--repository="stuttgart-things/dagger" \
--token=env:GITHUB_TOKEN \
--branch="main" \
--status="completed" \
--limit=5

# FILTER BY WORKFLOW NAME
dagger call -m git list-github-workflow-runs \
--repository="stuttgart-things/dagger" \
--token=env:GITHUB_TOKEN \
--workflow="ci.yaml"
```

Example output:

```
ID          WORKFLOW          STATUS     CONCLUSION  BRANCH  EVENT  TITLE                             CREATED
123456789   Release Pipeline  completed  success     main    push   chore(release): 0.80.0 [skip ci]  2026-02-18T12:34:56
123456788   Lint & Test       completed  failure     feat/x  push   feat: add new feature             2026-02-18T10:20:30
```

### Wait for GitHub Workflow Run

```bash
# WAIT FOR A SPECIFIC WORKFLOW RUN TO COMPLETE
dagger call -m git wait-for-github-workflow-run \
--repository="stuttgart-things/dagger" \
--run-id="1234567890" \
--token=env:GITHUB_TOKEN
```

## Examples

See the [main README](../README.md#git) for detailed usage examples.

## Resources

- [Git Documentation](https://git-scm.com/doc)
- [GitHub CLI](https://cli.github.com/)
- [GitLab CLI](https://gitlab.com/gitlab-org/cli)
- [Conventional Commits](https://conventionalcommits.org/)
- [Git Flow](https://nvie.com/posts/a-successful-git-branching-model/)
