package main

import (
	"context"
	"dagger/ansible/internal/dagger"
	"fmt"
)

// BUILDS A GIVEN COLLECTION DIR TO A ARCHIVE FILE (.TGZ)
func (m *Ansible) GithubRelease(
	ctx context.Context,
	tag string,
	title string,
	group string,
	repo string,
	files []*dagger.File,
	notes string,
	token *dagger.Secret,
) error {

	releaseOptions := dagger.GhReleaseCreateOpts{
		Repo:      group + "/" + repo,
		VerifyTag: false,
		Files:     files,
		Token:     token,
	}

	// CREATE GITHUB RELEASE
	err := dag.
		Gh().
		Release().
		Create(
			ctx,
			tag,
			title,
			releaseOptions,
		)
	if err != nil {
		return fmt.Errorf("failed to create GitHub release: %w", err)
	}

	return nil
}
