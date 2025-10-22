package main

import (
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
