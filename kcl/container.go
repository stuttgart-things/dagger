package main

import (
	"context"
	"dagger/kcl/internal/dagger"
)

func (m *Kcl) container() *dagger.Container {
	if m.BaseImage == "" {
		m.BaseImage = "ubuntu:24.04"
	}

	ctr := dag.Container().From(m.BaseImage)

	// Combine apt-get update and install into a single command for efficiency
	// This reduces container layers and speeds up container creation
	ctr = ctr.WithExec([]string{
		"sh", "-c",
		"apt-get update && apt-get install -y --no-install-recommends curl wget git ca-certificates && rm -rf /var/lib/apt/lists/*",
	})

	// Install KCL CLI using official installation script (same as container-use)
	ctr = ctr.WithExec([]string{"bash", "-c", "curl -fsSL https://kcl-lang.io/script/install-cli.sh | bash"})

	// Verify installation
	ctr = ctr.WithExec([]string{"kcl", "version"})

	// Set KCL as the default entrypoint
	ctr = ctr.WithEntrypoint([]string{"kcl"})

	return ctr
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
