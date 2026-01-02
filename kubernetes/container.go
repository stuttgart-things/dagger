package main

import (
	"dagger/kubernetes/internal/dagger"
)

func (m *Kubernetes) container() *dagger.Container {

	// --- BASE IMAGE ---
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	ctr := dag.Container().
		From(m.BaseImage).
		WithExec([]string{
			"apk",
			"add",
			"--no-cache",
			"wget",
			"yq",
			"jq",
			"curl",
			"git",
			"kubectl",
			"helm"})

	return ctr
}
