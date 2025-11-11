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

	ctr = ctr.WithExec([]string{
		"gem",
		"install",
		"mdl",
		"--no-document",
	})

	ctr = ctr.WithExec([]string{
		"pip3",
		"install",
		"--break-system-packages",
		"pre-commit",
		"detect-secrets",
	})

	return ctr
}
