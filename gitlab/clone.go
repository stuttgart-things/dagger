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
