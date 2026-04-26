package main

import (
	"context"
	"dagger/ansible/internal/dagger"
	"fmt"
	"time"
)

// CREATES A GITHUB RELEASE FOR THE GIVEN TAG, UPLOADING THE PROVIDED FILES AS ASSETS.
// Uses a prebuilt Wolfi image with `apk add gh git` instead of the daggerverse `gh`
// module — that module builds its container via apko, which fails with
// `mkdir /work/cache/...: permission denied` when apko runs as non-root.
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

	repoFull := group + "/" + repo

	container := dag.Container().
		From("cgr.dev/chainguard/wolfi-base:latest").
		WithExec([]string{"apk", "add", "--no-cache", "gh", "git"}).
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithEnvVariable("GH_PROMPT_DISABLED", "true").
		WithEnvVariable("GH_NO_UPDATE_NOTIFIER", "true").
		WithEnvVariable("GH_REPO", repoFull).
		WithSecretVariable("GITHUB_TOKEN", token).
		WithWorkdir("/work")

	args := []string{"gh", "release", "create", tag, "--title", title, "--notes", notes}

	for i, f := range files {
		name, err := f.Name(ctx)
		if err != nil {
			return fmt.Errorf("failed to get file name for asset %d: %w", i, err)
		}
		assetPath := "/work/" + name
		container = container.WithMountedFile(assetPath, f)
		args = append(args, assetPath)
	}

	if _, err := container.WithExec(args).Sync(ctx); err != nil {
		return fmt.Errorf("failed to create GitHub release: %w", err)
	}

	return nil
}
