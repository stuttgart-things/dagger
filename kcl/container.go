package main

import (
	"context"
	"dagger/kcl/internal/dagger"
)

func (m *Kcl) container() *dagger.Container {
	if m.BaseImage == "" {
		m.BaseImage = "kcllang/kcl:v0.12.3"
	}

	return dag.Container().
		From(m.BaseImage).
		WithExec([]string{"sh", "-c", "apt-get update && apt-get install -y --no-install-recommends jq curl && rm -rf /var/lib/apt/lists/*"}).
		WithExec([]string{"sh", "-c", "curl -sL https://github.com/mikefarah/yq/releases/download/v4.44.1/yq_linux_amd64 -o /usr/local/bin/yq && chmod +x /usr/local/bin/yq"})
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
