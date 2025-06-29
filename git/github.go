package main

import (
	"context"
	"dagger/git/internal/dagger"
	"fmt"
)

var (
	workDir = "/src"
)

func (m *Git) CloneGitHub(
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
