package main

import (
	"context"
	"dagger/kcl/internal/dagger"
)

func (m *Kcl) container() *dagger.Container {
	if m.BaseImage == "" {
		m.BaseImage = "kcllang/kcl:v0.12.0"
	}

	return dag.Container().From(m.BaseImage)
}

// ValidateKcl validates KCL configuration files by compiling them
func (m *Kcl) ValidateKcl(ctx context.Context, source *dagger.Directory) (string, error) {
	// Validation by compilation - if files compile without errors, they are valid
	return m.container().
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"sh", "-c", "kcl run main.k > /dev/null && echo 'Validation successful: KCL files are syntactically correct'"}).
		Stdout(ctx)
}

// KclVersion returns the installed KCL version
func (m *Kcl) KclVersion(ctx context.Context) (string, error) {
	return m.container().
		WithExec([]string{"kcl", "version"}).
		Stdout(ctx)
}
