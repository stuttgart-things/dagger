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

	// Install required packages for KCL (same as container-use environment)
	ctr = ctr.WithExec([]string{"apt-get", "update"})
	ctr = ctr.WithExec([]string{"apt-get", "install", "-y", "curl", "wget", "git", "ca-certificates"})

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
