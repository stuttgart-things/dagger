package main

import (
	"context"
	"dagger/hugo/internal/dagger"
)

func (m *Hugo) InitSite(
	ctx context.Context,
	name string,
	config *dagger.File,
	content *dagger.Directory,
	// The Theme to use
	// +optional
	// +default="github.com/joshed-io/reveal-hugo"
	theme string,

) (*dagger.Directory, error) {
	// CREATE BASE CONTAINER WITH HUGO INSTALLED
	baseCtr := m.container()

	// CREATE NEW SITE AND INITIALIZE MODULES
	ctr := baseCtr.
		WithExec([]string{"hugo", "new", "site", name}).
		WithWorkdir(name).
		WithExec([]string{"hugo", "mod", "init", name}).
		WithExec([]string{"hugo", "mod", "get", theme}).
		WithExec([]string{"hugo", "mod", "vendor"})

	// GET THE INITIALIZED SITE DIRECTORY
	siteDir := ctr.Directory(".")
	siteDir = siteDir.WithFile("hugo.toml", config)
	siteDir = siteDir.WithDirectory("/content", content)

	return siteDir, nil
}
