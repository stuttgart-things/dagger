// Slidev module: scaffold, serve, and build Slidev decks from a single
// slides.md (+ optional style.css) without keeping generated artifacts in git.
package main

import "dagger/slidev/internal/dagger"

type Slidev struct {
	// +optional
	// +default="node:22-alpine"
	BaseImage string
}

func (m *Slidev) container() *dagger.Container {
	if m.BaseImage == "" {
		m.BaseImage = "node:22-alpine"
	}

	return dag.Container().
		From(m.BaseImage).
		WithExec([]string{"apk", "add", "--no-cache", "git"}).
		WithExec([]string{"npm", "install", "-g", "pnpm"}).
		WithEnvVariable("npm_config_node_linker", "hoisted")
}
