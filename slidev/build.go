package main

import (
	"context"
	"dagger/slidev/internal/dagger"
)

// Build runs `slidev build` and returns the generated dist/ directory ready
// to be served by any static host (nginx, MinIO, S3, GitHub Pages, ...).
func (m *Slidev) Build(
	ctx context.Context,
	slides *dagger.File,
	// +optional
	style *dagger.File,
	// +optional
	// +default="@slidev/theme-default"
	theme string,
	// +optional
	addons []string,
	// +optional
	extras *dagger.Directory,
	// +optional
	// +default="/"
	base string,
) (*dagger.Directory, error) {
	deckDir, err := m.InitDeck(ctx, slides, style, theme, addons, extras)
	if err != nil {
		return nil, err
	}

	return m.container().
		WithMountedDirectory("/deck", deckDir).
		WithWorkdir("/deck").
		WithExec([]string{
			"pnpm", "exec", "slidev", "build", "slides.md",
			"--base", base,
		}).
		Directory("/deck/dist"), nil
}
