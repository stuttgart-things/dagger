package main

import (
	"dagger/linting/internal/dagger"
)

func (m *Linting) container() *dagger.Container {
	if m.BaseImage == "" {
		m.BaseImage = "alpine"
	}

	ctr := dag.Container().From(m.BaseImage)

	ctr = ctr.WithExec([]string{
		"apk",
		"add",
		"--no-cache",
		"wget",
		"yamllint",
		"ruby",
		"python3",
		"py3-pip",
		"git",
		"bash",
		"shellcheck",
		"docker-cli",
	})

	// Combine gem and pip installs into a single shell command for efficiency
	// This reduces container layer overhead
	ctr = ctr.WithExec([]string{
		"sh", "-c",
		"gem install mdl --no-document && pip3 install --break-system-packages --no-cache-dir pre-commit detect-secrets",
	})

	return ctr
}
