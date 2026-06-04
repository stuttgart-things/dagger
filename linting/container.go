package main

import (
	"dagger/linting/internal/dagger"
)

func (m *Linting) container() *dagger.Container {
	if m.BaseImage == "" {
		// Pin the base image (tag, not :latest) so pre-commit runs are
		// reproducible and don't depend on docker.io resolving a mutable
		// alpine:latest on every run. Overridable via --base-image. See #288.
		m.BaseImage = "alpine:3.21"
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
		"jq",
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
