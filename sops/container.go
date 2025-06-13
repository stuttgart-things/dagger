package main

import (
	"context"
	"dagger/sops/internal/dagger"
)

func (m *Sops) container(ctx context.Context) (*dagger.Container, error) {
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}
	ctr := dag.Container().
		From(m.BaseImage).
		WithExec([]string{"apk", "add", "--no-cache", "sops"}).
		WithEntrypoint([]string{"sops"})

	return ctr, nil
}
