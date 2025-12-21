package main

import (
	"context"
	"dagger/crossplane/internal/dagger"
)

// GetXplaneContainer return the default image for crossplane
func (m *Crossplane) GetXplaneContainer(ctx context.Context) *dagger.Container {
	return dag.Container().
		From("cgr.dev/chainguard/wolfi-base:latest").
		WithExec([]string{"apk", "add", "curl"}).
		WithExec([]string{"curl", "https://releases.crossplane.io/stable/current/bin/linux_amd64/crank", "--output", "crossplane"}).
		WithExec([]string{"mv", "crossplane", "/usr/bin/crossplane"}).
		WithExec([]string{"chmod", "+x", "/usr/bin/crossplane"})
}
