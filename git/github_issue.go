package main

import (
	"context"
	"dagger/git/internal/dagger"
	"fmt"
	"strings"
)

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
