package main

import (
	"context"

	"dagger/git/internal/dagger"
)

func (m *Git) container(
	ctx context.Context) (*dagger.Container, error) {
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	ctr := dag.Container().
		From(m.BaseImage).
		WithExec([]string{"apk", "add", "--no-cache", "git"}).
		WithEntrypoint([]string{"git"})

	return ctr, nil
}
