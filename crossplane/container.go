package main

import (
	"context"
	"dagger/crossplane/internal/dagger"
)

// GetXplaneContainer returns the default image for Crossplane with crossplane and kcl2xrd installed
func (m *Crossplane) GetXplaneContainer(ctx context.Context) *dagger.Container {
	return dag.Container().
		From("cgr.dev/chainguard/wolfi-base:latest").
		// Install dependencies
		WithExec([]string{"apk", "add", "curl", "yq"}).
		// Install crossplane
		WithExec([]string{"curl", "https://releases.crossplane.io/stable/current/bin/linux_amd64/crank", "--output", "crossplane"}).
		WithExec([]string{"mv", "crossplane", "/usr/bin/crossplane"}).
		WithExec([]string{"chmod", "+x", "/usr/bin/crossplane"}).
		// Install kcl2xrd
		WithExec([]string{"curl", "-L", "https://github.com/ggkhrmv/kcl2xrd/releases/download/v0.8.0/kcl2xrd-linux-amd64", "--output", "kcl2xrd"}).
		WithExec([]string{"mv", "kcl2xrd", "/usr/bin/kcl2xrd"}).
		WithExec([]string{"chmod", "+x", "/usr/bin/kcl2xrd"})
}
