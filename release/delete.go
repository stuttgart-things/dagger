package main

import (
	"context"
	"dagger/release/internal/dagger"
)

func (m *Release) DeleteTag(
	ctx context.Context,
	// +optional
	releaseTag string,
	// +optional
	// +default="1.0.18-light"
	semanticReleaseVersion string,
	// Source folder (e.g. ".")
	src *dagger.Directory,
	gitConfig *dagger.Secret,
) (*dagger.Directory, error) {

	gitConfigPath := "/root/.gitconfig"

	gitContainer := m.container(semanticReleaseVersion).
		WithMountedSecret(gitConfigPath, gitConfig).
		WithMountedDirectory("/repo", src).
		WithWorkdir("/repo")

	// GIT PULL TO REFRESH TAGS
	gitContainer, err := gitContainer.
		WithExec([]string{
			"git",
			"pull"}).
		Sync(ctx)

	if err != nil {
		return nil, err
	}

	gitContainer, err = gitContainer.
		WithExec([]string{"git", "fetch", "--tags"}).
		Sync(ctx)
	if err != nil {
		return nil, err
	}

	gitContainer, err = gitContainer.
		WithExec([]string{"git", "tag", "-d", releaseTag}).
		Sync(ctx)
	if err != nil {
		return nil, err
	}

	gitContainer = gitContainer.
		WithMountedSecret(gitConfigPath, gitConfig)

	_, err = gitContainer.
		WithExec([]string{"git", "push", "origin", "--delete", "tag", releaseTag}).
		Sync(ctx)
	if err != nil {
		return nil, err
	}

	gitContainer, err = gitContainer.
		WithExec([]string{
			"git",
			"pull"}).
		Sync(ctx)

	if err != nil {
		return nil, err
	}

	gitContainer, err = gitContainer.
		WithExec([]string{"git", "fetch", "--tags"}).
		Sync(ctx)
	if err != nil {
		return nil, err
	}

	return src, nil
}
