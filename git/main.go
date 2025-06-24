package main

import (
	"dagger/git/internal/dagger"
)

type Git struct{}

func (m *Git) CloneGitHub(repository string, token *dagger.Secret) *dagger.Directory {
	return dag.
		Gh().
		Repo().
		Clone(repository, dagger.GhRepoCloneOpts{Token: token})
}
