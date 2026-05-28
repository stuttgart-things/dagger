package main

import (
	"context"
	"dagger/git/internal/dagger"
	"fmt"
	"strings"
)

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
