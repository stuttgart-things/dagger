package main

import (
	"context"
	"fmt"

	"dagger/gitlab/internal/dagger"
)

// ClonePrivateRepo clones a private repository and returns a Dagger Directory

// Clone clones a git repo using a container and returns the Directory
func (g *Gitlab) Clone(
	ctx context.Context,
	repoURL string,
	token dagger.Secret,
	branch string,
) (*dagger.Directory, error) {
	authToken, err := token.Plaintext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read token: %w", err)
	}

	// Prepare the URL with token for private repo access
	// Example: https://oauth2:<token>@gitlab.com/yourgroup/yourrepo.git
	authenticatedRepoURL := fmt.Sprintf("https://oauth2:%s@%s", authToken, repoURL[len("https://"):])

	// Start a container with Git installed
	container := dag.Container().From("alpine/git")
	container = container.WithExec([]string{"git", "clone", "--branch", branch, authenticatedRepoURL, "/repo"})

	// Return the directory where the repo is cloned
	return container.Directory("/repo"), nil
}

func (g *Gitlab) Print(
	ctx context.Context,
	repoURL string,
	server string,
	token dagger.Secret,
	projectID string,
	mergeRequestID string,
	branch string,
) error {

	// 1. Clone the repo
	repoDir, err := g.Clone(ctx, repoURL, token, branch)
	if err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	// 2. Get list of changed files from MR
	changedFiles, err := g.ListMergeRequestChanges(ctx, server, token, projectID, mergeRequestID)
	if err != nil {
		return fmt.Errorf("failed to list changed files: %w", err)
	}

	// 3. For each changed file, read and print its content
	for _, filePath := range changedFiles {
		file := repoDir.File(filePath)

		content, err := file.Contents(ctx)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		fmt.Printf("=== File: %s ===\n%s\n\n", filePath, content)
	}

	return nil
}
