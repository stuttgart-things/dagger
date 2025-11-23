# Dagger Git Module - Usage Examples

## Setup

First, set your GitHub token as an environment variable:
```bash
export GITHUB_TOKEN="ghp_your_token_here"
```

## 1. Create a GitHub Issue

Create an issue in a repository:

```bash
dagger call create-github-issue \
  --repository="owner/repo" \
  --ref="main" \
  --title="Bug: Application crashes on startup" \
  --body="The application crashes when started with empty config" \
  --label="bug" \
  --assignees="username1,username2" \
  --token=env:GITHUB_TOKEN
```

## 2. Clone a GitHub Repository

Clone a repo and checkout a specific branch:

```bash
dagger call clone-github \
  --repository="owner/repo" \
  --ref="feature-branch" \
  --token=env:GITHUB_TOKEN \
  export --path=./cloned-repo
```

## 3. Add Files to Branch (Root Level)

Add files from a local directory to the root of a branch:

```bash
dagger call add-files-to-branch \
  --repository="owner/repo" \
  --branch="feature-branch" \
  --files=./local-directory \
  --commit-message="Add new files" \
  --token=env:GITHUB_TOKEN \
  --author-name="John Doe" \
  --author-email="john@example.com"
```

## 4. Add Files to Specific Path

Add files to a specific directory within the repository:

```bash
dagger call add-files-to-branch-at-paths \
  --repository="owner/repo" \
  --branch="main" \
  --commit-message="Update documentation" \
  --token=env:GITHUB_TOKEN \
  --source-dir=./docs \
  --destination-path="docs/" \
  --author-name="Bot" \
  --author-email="bot@example.com"
```

Example: Add configuration files to a nested path:

```bash
dagger call add-files-to-branch-at-paths \
  --repository="owner/repo" \
  --branch="config-update" \
  --commit-message="Add Kubernetes configs" \
  --token=env:GITHUB_TOKEN \
  --source-dir=./k8s-configs \
  --destination-path="deploy/kubernetes/" \
  --author-name="DevOps Bot"
```

## 5. Add Single File to Specific Path

Add or update a single file at a precise location:

```bash
dagger call add-file-to-path \
  --repository="owner/repo" \
  --branch="docs-update" \
  --commit-message="Update README" \
  --token=env:GITHUB_TOKEN \
  --source-file=./README.md \
  --destination-path="README.md"
```

Example: Add a config file to a nested directory:

```bash
dagger call add-file-to-path \
  --repository="owner/repo" \
  --branch="config" \
  --commit-message="Add application config" \
  --token=env:GITHUB_TOKEN \
  --source-file=./app.yaml \
  --destination-path="config/app.yaml"
```

Example: Update a file in a subdirectory:

```bash
dagger call add-file-to-path \
  --repository="owner/repo" \
  --branch="feature-auth" \
  --commit-message="Update auth handler" \
  --token=env:GITHUB_TOKEN \
  --source-file=./auth.go \
  --destination-path="src/handlers/auth.go" \
  --author-name="Developer" \
  --author-email="dev@example.com"
```

## 6. Create a New Branch

Create a new branch from an existing branch:

```bash
dagger call create-branch \
  --repository="owner/repo" \
  --new-branch="feature-branch" \
  --base-branch="main" \
  --token=env:GITHUB_TOKEN
```

Create a branch from develop:

```bash
dagger call create-branch \
  --repository="owner/repo" \
  --new-branch="bugfix-auth" \
  --base-branch="develop" \
  --token=env:GITHUB_TOKEN
```

## Complete Workflow Example

Here's a complete workflow that creates a branch, adds files, and creates a PR:

```bash
# 1. Create a new branch
dagger call create-branch \
  --repository="owner/repo" \
  --new-branch="new-feature" \
  --base-branch="main" \
  --token=env:GITHUB_TOKEN

# 2. Add multiple files to different locations
dagger call add-file-to-path \
  --repository="owner/repo" \
  --branch="new-feature" \
  --commit-message="Add configuration files" \
  --token=env:GITHUB_TOKEN \
  --source-file=./config.yaml \
  --destination-path="config/config.yaml"

# 3. Add documentation
dagger call add-files-to-branch-at-paths \
  --repository="owner/repo" \
  --branch="new-feature" \
  --commit-message="Add documentation" \
  --token=env:GITHUB_TOKEN \
  --source-dir=./docs \
  --destination-path="docs/"

# 4. Create a PR for the changes
gh pr create \
  --repo owner/repo \
  --head new-feature \
  --base main \
  --title "Add new feature" \
  --body "This PR adds the new feature with configs and docs"
```

## Tips

1. **Using with Dagger Directory**: You can chain Dagger directory operations:
```bash
dagger call add-files-to-branch-at-paths \
  --source-dir=$(dagger call some-other-function) \
  --destination-path="target/path/"
```

2. **Default Values**: If you don't specify `author-name` and `author-email`, they default to "Dagger Bot" and "bot@dagger.io"

3. **Branch Creation**: These functions expect the branch to already exist. Create it first using `gh` CLI or GitHub API.

4. **Multiple Files in One Commit**: Currently, each function call creates a separate commit. To add multiple files/directories in one commit, you'd need to chain the operations or create a custom function.
