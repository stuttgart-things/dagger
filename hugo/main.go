package main

import (
	"dagger/hugo/internal/dagger"
)

type Hugo struct {
	// +optional
	// +default="cgr.dev/chainguard/wolfi-base:latest"
	BaseImage string
}

func (m *Hugo) container() *dagger.Container {
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	ctr := dag.Container().From(m.BaseImage)
	ctr = ctr.WithExec([]string{"apk", "add", "--no-cache", "hugo", "go", "git"})
	ctr = ctr.WithEntrypoint([]string{"hugo"})

	return ctr
}
