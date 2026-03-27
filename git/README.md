# Git Dagger Module

Git and GitHub operations: clone, branch, commit, PR, issues, and workflow management.

## Setup

```bash
export GITHUB_TOKEN="ghp_your_token_here"
```

## Functions

### Clone GitHub Repository

```bash
dagger call -m git clone-github \
--repository stuttgart-things/stuttgart-things \
--token env:GITHUB_TOKEN \
export --path=/tmp/cloned-repo

# Clone specific branch
dagger call -m git clone-github \
--repository stuttgart-things/stuttgart-things \
--ref feature-branch \
--token env:GITHUB_TOKEN \
export --path=/tmp/cloned-repo
```

### Create GitHub Branch

```bash
dagger call -m git create-github-branch \
--repository stuttgart-things/dagger \
--new-branch feature-branch \
--base-branch main \
--token env:GITHUB_TOKEN
```

### Delete GitHub Branch

```bash
dagger call -m git delete-github-branch \
--repository stuttgart-things/dagger \
--branch feature-branch \
--token env:GITHUB_TOKEN
```

### Add File to GitHub Branch

```bash
# Branch must exist
dagger call -m git add-file-to-github-branch \
--repository stuttgart-things/dagger \
--branch feature-branch \
--commit-message "Update README" \
--token env:GITHUB_TOKEN \
--source-file ./README.md \
--destination-path "README-updated.md"
```

### Add Folder to GitHub Branch

```bash
dagger call -m git add-folder-to-github-branch \
--repository stuttgart-things/dagger \
--branch feature-branch \
--commit-message "Update docs" \
--token env:GITHUB_TOKEN \
--source-dir ./docs \
--destination-path "docs/"
```

### Add Multiple Files to GitHub Branch

```bash
dagger call -m git add-files-to-github-branch \
--repository stuttgart-things/dagger \
--branch feature-branch \
--commit-message "Add config files" \
--token env:GITHUB_TOKEN \
--source-dir ./k8s-configs \
--destination-path "deploy/kubernetes/"
```

### Create GitHub Pull Request

```bash
dagger call -m git create-github-pull-request \
--repository stuttgart-things/dagger \
--head-branch feature-branch \
--title "Add new feature" \
--body "This PR adds a new feature" \
--labels enhancement \
--labels documentation \
--reviewers patrick-hermann-sva \
--token env:GITHUB_TOKEN
```

### Create GitHub Issue

```bash
dagger call -m git create-github-issue \
--repository stuttgart-things/stuttgart-things \
--token env:GITHUB_TOKEN \
--title "Bug: Application crashes on startup" \
--body "The application crashes when started with empty config" \
--label bug \
--label automation
```

### List GitHub Workflow Runs

```bash
# List all runs
dagger call -m git list-github-workflow-runs \
--repository stuttgart-things/dagger \
--token env:GITHUB_TOKEN

# Filter by branch, status, and limit
dagger call -m git list-github-workflow-runs \
--repository stuttgart-things/dagger \
--token env:GITHUB_TOKEN \
--branch main \
--status completed \
--limit 5

# Filter by workflow name
dagger call -m git list-github-workflow-runs \
--repository stuttgart-things/dagger \
--token env:GITHUB_TOKEN \
--workflow ci.yaml
```

Example output:

```
ID          WORKFLOW          STATUS     CONCLUSION  BRANCH  EVENT  TITLE                             CREATED
123456789   Release Pipeline  completed  success     main    push   chore(release): 0.80.0 [skip ci]  2026-02-18T12:34:56
123456788   Lint & Test       completed  failure     feat/x  push   feat: add new feature             2026-02-18T10:20:30
```

### Wait for GitHub Workflow Run

```bash
dagger call -m git wait-for-github-workflow-run \
--repository stuttgart-things/dagger \
--run-id 1234567890 \
--token env:GITHUB_TOKEN
```

## Complete Workflow Example

Create a branch, add files, and open a PR:

```bash
# 1. Create a new branch
dagger call -m git create-github-branch \
--repository stuttgart-things/stuttgart-things \
--new-branch feat/new-config \
--base-branch main \
--token env:GITHUB_TOKEN

# 2. Add rendered config files
dagger call -m git add-folder-to-github-branch \
--repository stuttgart-things/stuttgart-things \
--branch feat/new-config \
--commit-message "feat: add rendered config" \
--token env:GITHUB_TOKEN \
--source-dir ./rendered-output \
--destination-path "config/"

# 3. Create a PR
dagger call -m git create-github-pull-request \
--repository stuttgart-things/stuttgart-things \
--head-branch feat/new-config \
--title "feat: add rendered config" \
--body "This PR adds rendered configuration files" \
--token env:GITHUB_TOKEN
```

## Notes

- Default author: "Dagger Bot" / "bot@dagger.io" (override with `--author-name` / `--author-email`)
- Branch must exist before adding files (create it first with `create-github-branch`)
- Each `add-*` call creates a separate commit
