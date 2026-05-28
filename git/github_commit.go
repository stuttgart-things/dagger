package main

import (
	"context"
	"dagger/git/internal/dagger"
	"fmt"
	"strings"
	"time"
)

// AddFileToGithubBranch adds a single file to a specific path in a remote branch.
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

	// Get the base container with git and gh
	ctr, err := m.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	// Clone via gh CLI inside our own container (avoids the upstream gh module,
	// whose apko-built base image fails on shared cache volumes with "permission denied").
	// CACHE_BUSTER forces Dagger to re-run clone+fetch+checkout every invocation;
	// without it the cache key is static and /src ends up at a stale origin/main,
	// causing "fetch first" push rejections (blueprints#158).
	ctr = ctr.
		WithEnvVariable("CACHE_BUSTER", fmt.Sprintf("%d", time.Now().UnixNano())).
		WithSecretVariable("GH_TOKEN", token).
		WithEnvVariable("GH_REPO", repository).
		WithExec([]string{"gh", "repo", "clone", repository, workDir})

	// Configure git user and authentication
	ctr = ctr.
		WithWorkdir(workDir).
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
	ctr = ctr.WithExec([]string{"sh", "-c", checkoutCmd})

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
