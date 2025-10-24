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
	})

	ctr = ctr.WithExec([]string{
		"gem",
		"install",
		"mdl",
		"--no-document",
	})

	return ctr
}
