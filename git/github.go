package main

import (
	"context"
	"dagger/git/internal/dagger"
	"fmt"
	"strings"
)

var (
	workDir = "/src"
)

// CreateGithubBranch creates a new branch in a GitHub repository based on an existing ref.
// If the branch already exists, it returns an error.
// Returns the name of the created branch.
func (m *Git) CreateGithubBranch(
	ctx context.Context,
	// Repository in format "owner/repo"
	repository string,
	// Name of the new branch to create
	newBranch string,
	// GitHub token for authentication
	token *dagger.Secret,
	// Base ref/branch to create from (e.g., "main", "develop")
	// +optional
	// +default="main"
	baseBranch string) (string, error) {

	if baseBranch == "" {
		baseBranch = "main"
	}

	// Get the base container with git and gh
	ctr, err := m.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	// Set up authentication and repository context
	ctr = ctr.
		WithSecretVariable("GITHUB_TOKEN", token).
		WithEnvVariable("GH_REPO", repository)

	// Create new branch using gh CLI directly (no cloning needed)
	createCmd := fmt.Sprintf("gh api repos/%s/git/refs/heads/%s --jq .object.sha | xargs -I {} gh api repos/%s/git/refs -f ref=refs/heads/%s -f sha={}",
		repository, baseBranch, repository, newBranch)

	output, err := ctr.
		WithExec([]string{"sh", "-c", createCmd}).
		Stdout(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to create branch: %w", err)
	}

	// Verify branch was created
	if strings.Contains(output, newBranch) {
		return newBranch, nil
	}

	return newBranch, nil
} // DeleteGithubBranch deletes a branch from a GitHub repository.
// This will delete the branch from the remote repository.
// Returns a success message with the deleted branch name.
func (m *Git) DeleteGithubBranch(
	ctx context.Context,
	// Repository in format "owner/repo"
	repository string,
	// Name of the branch to delete
	branch string,
	// GitHub token for authentication
	token *dagger.Secret) (string, error) {

	// Clone the repository
	gitDir := m.CloneGithub(ctx, repository, "main", token)

	// Get the base container with git and gh
	ctr, err := m.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	// Mount the repository and configure
	ctr = ctr.
		WithDirectory(workDir, gitDir).
		WithWorkdir(workDir).
		WithSecretVariable("GH_TOKEN", token)

	// Configure git to use gh for authentication
	ctr = ctr.
		WithExec([]string{"git", "config", "--global", "credential.helper", "!gh auth git-credential"})

	// Delete the remote branch
	ctr = ctr.WithExec([]string{"git", "push", "origin", "--delete", branch})

	// Verify branch was deleted
	output, err := ctr.WithExec([]string{"git", "branch", "-r"}).Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to verify branch deletion: %w", err)
	}

	// Check if the branch still exists in remote branches
	if strings.Contains(output, "origin/"+branch) {
		return "", fmt.Errorf("branch %s was not deleted successfully", branch)
	}

	return fmt.Sprintf("Successfully deleted branch: %s", branch), nil
}

// CreateGithubPullRequest creates a new pull request in a GitHub repository.
// It uses the GitHub CLI to create a PR from the head branch to the base branch.
// Returns the URL of the created pull request.
func (m *Git) CreateGithubPullRequest(
	ctx context.Context,
	// Repository in format "owner/repo"
	repository string,
	// Head branch (the branch with your changes)
	headBranch string,
	// Base branch to merge into (e.g., "main", "develop")
	// +optional
	// +default="main"
	baseBranch string,
	// Pull request title
	title string,
	// Pull request body/description
	body string,
	// Optional labels to add to the PR
	// +optional
	labels []string,
	// Optional reviewers to request
	// +optional
	reviewers []string,
	// GitHub token for authentication
	token *dagger.Secret) (string, error) {

	// Get the base container with git and gh
	ctr, err := m.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	// Set authentication
	ctr = ctr.WithSecretVariable("GH_TOKEN", token)

	// Build the gh pr create command with --repo flag (no need to clone)
	args := []string{"gh", "pr", "create", "--repo", repository, "--head", headBranch, "--base", baseBranch, "--title", title, "--body", body}

	// Add labels if provided
	for _, label := range labels {
		args = append(args, "--label", label)
	}

	// Add reviewers if provided
	for _, reviewer := range reviewers {
		args = append(args, "--reviewer", reviewer)
	}

	// Execute the command and capture output
	output, err := ctr.
		WithEntrypoint([]string{}).
		WithExec(args).
		Stdout(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to create pull request: %w", err)
	}

	// The output from gh pr create is the URL
	return strings.TrimSpace(output), nil
}

// CreateGithubIssue creates a new GitHub issue in the specified repository.
// It clones the repository, authenticates using the provided token, and uses
// the GitHub CLI to create an issue with the given title, body, optional labels,
// and optional assignees. Returns the URL of the created issue.
func (m *Git) CreateGithubIssue(
	ctx context.Context,
	// Repository in format "owner/repo"
	repository string,
	// Ref/Branch to checkout - If not specified, defaults to "main"
	// +optional
	// +default="main"
	ref string,
	title,
	body,
	// +optional
	label string,
	// +optional
	assignees []string,
	// GitHub token for authentication
	token *dagger.Secret) (string, error) {

	// Clone the repository to establish context for gh CLI
	gitDir := m.CloneGithub(ctx, repository, ref, token)

	// Get the base container with git and gh
	ctr, err := m.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	// Mount the repository and set as working directory
	ctr = ctr.
		WithDirectory("/work", gitDir).
		WithWorkdir("/work").
		WithSecretVariable("GH_TOKEN", token)

	// Build the gh issue create command
	args := []string{"gh", "issue", "create", "--title", title, "--body", body}

	// Add label if provided
	if label != "" {
		args = append(args, "--label", label)
	}

	// Add assignees if provided
	for _, assignee := range assignees {
		args = append(args, "--assignee", assignee)
	}

	// Execute the command and capture output
	output, err := ctr.
		WithEntrypoint([]string{}).
		WithExec(args).
		Stdout(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to create issue: %w", err)
	}

	// The output from gh issue create is the URL
	return strings.TrimSpace(output), nil
}

// CloneGithub clones a GitHub repository and checks out the specified branch or ref.
// It uses the GitHub CLI to perform the clone operation with authentication,
// then checks out the requested ref/branch. Returns a Dagger Directory containing
// the cloned repository at the specified ref.
func (m *Git) CloneGithub(
	ctx context.Context,
	repository string,
	// Ref/Branch to checkout - If not specified, defaults to "main"
	// +optional
	// +default="main"
	ref string,
	token *dagger.Secret) *dagger.Directory {

	if ref == "" {
		ref = "main"
	}

	// Get the base container with git and gh
	ctr, err := m.container(ctx)
	if err != nil {
		fmt.Errorf("CONTAINER INIT FAILED: %w", err)
	}

	// Use gh to clone with authentication and checkout the specific branch
	ctr = ctr.
		WithSecretVariable("GITHUB_TOKEN", token).
		WithEnvVariable("GH_REPO", repository).
		WithExec([]string{"gh", "repo", "clone", repository, workDir, "--", "--branch", ref})

	return ctr.Directory(workDir)
}

// AddFileToPath adds a single file to a specific path in a remote branch.
// This is useful when you need to add or update individual files at precise locations
// within the repository structure.
// Returns the commit SHA of the created commit.
func (m *Git) AddFileToGithubBranch(
	ctx context.Context,
	// Repository in format "owner/repo"
	repository string,
	// Branch name to add files to
	branch string,
	// Commit message
	commitMessage string,
	// GitHub token for authentication
	token *dagger.Secret,
	// Source file to copy
	sourceFile *dagger.File,
	// Destination path within the repository (e.g., "docs/README.md" or "config.yaml")
	destinationPath string,
	// Git author name
	// +optional
	// +default="Dagger Bot"
	authorName string,
	// Git author email
	// +optional
	// +default="bot@dagger.io"
	authorEmail string) (string, error) {

	// Clone the repository
	gitDir := m.CloneGithub(ctx, repository, branch, token)

	// Get the base container with git and gh
	ctr, err := m.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	// Configure git user
	ctr = ctr.
		WithDirectory(workDir, gitDir).
		WithWorkdir(workDir).
		WithSecretVariable("GH_TOKEN", token).
		WithExec([]string{"git", "config", "user.name", authorName}).
		WithExec([]string{"git", "config", "user.email", authorEmail})

	// Configure git to use gh for authentication
	ctr = ctr.WithExec([]string{"git", "config", "--global", "credential.helper", "!gh auth git-credential"})

	// Copy the source file to the destination path
	targetPath := workDir + "/" + destinationPath
	ctr = ctr.WithFile(targetPath, sourceFile)

	// Stage all changes
	ctr = ctr.WithExec([]string{"git", "add", "."})

	// Check if there are changes to commit
	statusOutput, err := ctr.WithExec([]string{"git", "status", "--porcelain"}).Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to check git status: %w", err)
	}

	if strings.TrimSpace(statusOutput) == "" {
		return "", fmt.Errorf("no changes to commit")
	}

	// Commit the changes
	ctr = ctr.WithExec([]string{"git", "commit", "-m", commitMessage})

	// Push to remote
	ctr = ctr.WithExec([]string{"git", "push", "origin", branch})

	// Get the commit SHA
	commitSha, err := ctr.WithExec([]string{"git", "rev-parse", "HEAD"}).Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get commit SHA: %w", err)
	}

	return strings.TrimSpace(commitSha), nil
}

// AddFolderToGithubBranch adds a folder to a specific path in a remote branch.
// This allows you to place a directory and its contents at a specific location within
// the repository structure. Returns the commit SHA of the created commit.
func (m *Git) AddFolderToGithubBranch(
	ctx context.Context,
	// Repository in format "owner/repo"
	repository string,
	// Branch name to add files to
	branch string,
	// Commit message
	commitMessage string,
	// GitHub token for authentication
	token *dagger.Secret,
	// Source directory containing files to copy
	sourceDir *dagger.Directory,
	// Destination path within the repository (e.g., "docs/" or "src/config/")
	destinationPath string,
	// Git author name
	// +optional
	// +default="Dagger Bot"
	authorName string,
	// Git author email
	// +optional
	// +default="bot@dagger.io"
	authorEmail string) (string, error) {

	// Clone the repository using gh CLI
	gitDir := dag.Gh().Repo().Clone(repository, dagger.GhRepoCloneOpts{Token: token})

	// Get the base container with git and gh
	ctr, err := m.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	// Configure git user and authentication
	ctr = ctr.
		WithDirectory(workDir, gitDir).
		WithWorkdir(workDir).
		WithSecretVariable("GH_TOKEN", token).
		WithExec([]string{"git", "config", "user.name", authorName}).
		WithExec([]string{"git", "config", "user.email", authorEmail}).
		WithExec([]string{"git", "config", "--global", "credential.helper", "!gh auth git-credential"})

	// Fetch all branches to get the newly created branch
	ctr = ctr.WithExec([]string{"git", "fetch", "origin", "--force"})

	// Checkout the target branch (tracking remote)
	checkoutCmd := fmt.Sprintf(
		"git checkout %s 2>/dev/null || git checkout -b %s origin/%s",
		branch, branch, branch,
	)
	ctr = ctr.WithExec([]string{"sh", "-c", checkoutCmd}) // Copy the source directory to the destination path
	targetPath := workDir + "/" + destinationPath
	ctr = ctr.WithDirectory(targetPath, sourceDir, dagger.ContainerWithDirectoryOpts{
		Exclude: []string{".git"},
	})

	// Stage all changes
	ctr = ctr.WithExec([]string{"git", "add", "."})

	// Check if there are changes to commit
	statusOutput, err := ctr.WithExec([]string{"git", "status", "--porcelain"}).Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to check git status: %w", err)
	}

	if strings.TrimSpace(statusOutput) == "" {
		return "", fmt.Errorf("no changes to commit")
	}

	// Commit the changes
	ctr = ctr.WithExec([]string{"git", "commit", "-m", commitMessage})

	// Push to remote
	ctr = ctr.WithExec([]string{"git", "push", "origin", branch})

	// Get the commit SHA
	commitSha, err := ctr.WithExec([]string{"git", "rev-parse", "HEAD"}).Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get commit SHA: %w", err)
	}

	return strings.TrimSpace(commitSha), nil
}

// AddFilesToGithubBranch adds multiple files or directories to specific paths in a remote branch.
// This allows you to place files at different locations within the repository structure.
// Each file or directory is copied to its specified destination path before committing and pushing.
// Returns the commit SHA of the created commit.
func (m *Git) AddFilesToGithubBranch(
	ctx context.Context,
	// Repository in format "owner/repo"
	repository string,
	// Branch name to add files to
	branch string,
	// Commit message
	commitMessage string,
	// GitHub token for authentication
	token *dagger.Secret,
	// Source directory containing file to copy
	sourceDir *dagger.Directory,
	// Destination path within the repository (e.g., "docs/README.md" or "src/config/")
	destinationPath string,
	// Git author name
	// +optional
	// +default="Dagger Bot"
	authorName string,
	// Git author email
	// +optional
	// +default="bot@dagger.io"
	authorEmail string) (string, error) {

	// Clone the repository
	gitDir := m.CloneGithub(ctx, repository, branch, token)

	// Get the base container with git and gh
	ctr, err := m.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	// Configure git user
	ctr = ctr.
		WithDirectory(workDir, gitDir).
		WithWorkdir(workDir).
		WithSecretVariable("GH_TOKEN", token).
		WithExec([]string{"git", "config", "user.name", authorName}).
		WithExec([]string{"git", "config", "user.email", authorEmail})

	// Configure git to use gh for authentication
	ctr = ctr.WithExec([]string{"git", "config", "--global", "credential.helper", "!gh auth git-credential"})

	// Copy the source directory to the destination path
	targetPath := workDir + "/" + destinationPath
	ctr = ctr.WithDirectory(targetPath, sourceDir, dagger.ContainerWithDirectoryOpts{
		Exclude: []string{".git"},
	})

	// Stage all changes
	ctr = ctr.WithExec([]string{"git", "add", "."})

	// Check if there are changes to commit
	statusOutput, err := ctr.WithExec([]string{"git", "status", "--porcelain"}).Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to check git status: %w", err)
	}

	if strings.TrimSpace(statusOutput) == "" {
		return "", fmt.Errorf("no changes to commit")
	}

	// Commit the changes
	ctr = ctr.WithExec([]string{"git", "commit", "-m", commitMessage})

	// Push to remote
	ctr = ctr.WithExec([]string{"git", "push", "origin", branch})

	// Get the commit SHA
	commitSha, err := ctr.WithExec([]string{"git", "rev-parse", "HEAD"}).Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get commit SHA: %w", err)
	}

	return strings.TrimSpace(commitSha), nil
}
