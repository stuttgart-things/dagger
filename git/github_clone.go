package main

import (
	"context"
	"dagger/git/internal/dagger"
	"fmt"
	"time"
)

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
		panic(fmt.Errorf("CONTAINER INIT FAILED: %w", err))
	}

	// Use gh to clone with authentication and checkout the specific branch.
	// CACHE_BUSTER forces a fresh clone on every run so callers (AddFile/AddFiles)
	// don't push on top of a stale origin HEAD (blueprints#158).
	ctr = ctr.
		WithEnvVariable("CACHE_BUSTER", fmt.Sprintf("%d", time.Now().UnixNano())).
		WithSecretVariable("GH_TOKEN", token).
		WithEnvVariable("GH_REPO", repository).
		WithExec([]string{"gh", "repo", "clone", repository, workDir, "--", "--branch", ref})

	return ctr.Directory(workDir)
}
