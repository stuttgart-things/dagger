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

func (m *Git) CloneGithub(
	ctx context.Context,
	repository string,
	// Ref/Branch to checkout - If not specified, defaults to "main"
	// +optional
	// +default="main"
	ref string,
	token *dagger.Secret) *dagger.Directory {

	gitDir := dag.
		Gh().
		Repo().
		Clone(repository, dagger.GhRepoCloneOpts{Token: token})

	// GET THE BASE CONTAINER WITH GIT
	ctr, err := m.container(ctx)
	if err != nil {
		fmt.Errorf("CONTAINER INIT FAILED: %w", err)
	}

	// SWITCH TO REF/BRANCH
	ctr = ctr.WithDirectory(workDir, gitDir).WithWorkdir(workDir)
	ctr = ctr.WithExec([]string{"git", "checkout", ref})

	return ctr.Directory(workDir)
}
