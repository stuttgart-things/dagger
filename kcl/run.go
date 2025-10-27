package main

import (
	"context"
	"dagger/kcl/internal/dagger"
)

// RunKcl executes KCL code from a provided directory
func (m *Kcl) RunKcl(ctx context.Context, source *dagger.Directory, entrypoint string) (string, error) {
	if entrypoint == "" {
		entrypoint = "main.k"
	}

	return m.container().
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"kcl", "run", entrypoint}).
		Stdout(ctx)
}
