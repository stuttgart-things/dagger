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
		WithFile("hugo.toml", config).
		WithExec([]string{"hugo", "mod", "init", name}).
		WithExec([]string{"hugo", "mod", "get", theme}).
		WithExec([]string{"hugo", "mod", "tidy"}).
		WithExec([]string{"hugo", "mod", "vendor"}).
		WithExec([]string{"tree"})

	// GET THE INITIALIZED SITE DIRECTORY
	siteDir := ctr.Directory(".").
		WithFile("hugo.toml", config).
		WithDirectory("content", content)

	return siteDir, nil
}
