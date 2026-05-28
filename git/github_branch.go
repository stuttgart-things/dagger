package main

import (
	"context"
	"dagger/git/internal/dagger"
	"fmt"
	"strings"
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

	// Get the base container with git and gh
	ctr, err := m.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	// Set up authentication and repository context
	ctr = ctr.
		WithSecretVariable("GH_TOKEN", token).
		WithEnvVariable("GH_REPO", repository)

	// Get base branch SHA
	getShaCmd := fmt.Sprintf("gh api repos/%s/git/refs/heads/%s --jq .object.sha",
		repository, baseBranch)

	baseSha, err := ctr.
		WithExec([]string{"sh", "-c", getShaCmd}).
		Stdout(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to get base branch SHA: %w", err)
	}

	baseSha = strings.TrimSpace(baseSha)

	// Try to create the branch; if it already exists, update it to the base SHA
	createCmd := fmt.Sprintf(
		"gh api repos/%s/git/refs -f ref=refs/heads/%s -f sha=%s 2>/dev/null || "+
			"gh api repos/%s/git/refs/heads/%s -X PATCH -f sha=%s -F force=true",
		repository, newBranch, baseSha,
		repository, newBranch, baseSha)

	_, err = ctr.
		WithExec([]string{"sh", "-c", createCmd}).
		Stdout(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to create or update branch: %w", err)
	}

	return newBranch, nil
}

// DeleteGithubBranch deletes a branch from a GitHub repository.
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
