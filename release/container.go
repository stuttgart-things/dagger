package main

import (
	"dagger/release/internal/dagger"
)

func (m *Release) container(
	// The Semantic release version to use
	// +optional
	// +default="1.0.18-light"
	semanticReleaseVersion string,
) *dagger.Container {

	ctr := dag.
		Container().
		From("hoppr/semantic-release:" + semanticReleaseVersion)

	// INSTALL DEPS
	ctr = ctr.WithExec([]string{
		"npm", "install", "-g",
		"semantic-release@24.2.7",
		"@semantic-release/changelog",
		"@semantic-release/git",
		"@semantic-release/exec",
	})

	ctr = ctr.WithExec([]string{
		"apk",
		"add",
		"--no-cache",
		"make"})

	return ctr
}
