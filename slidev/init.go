package main

import (
	"context"
	"dagger/slidev/internal/dagger"
)

// InitDeck scaffolds a Slidev project from a single slides.md (+ optional
// style.css), installs @slidev/cli plus the chosen theme and any addons,
// and returns the populated /deck directory.
//
// `extras` is an optional directory overlaid on /deck/ between dependency
// install and the final slides.md/style.css writes. Use it to ship
// additional files the deck needs at runtime — e.g. partials referenced via
// `src: ./slides/...` includes, components/, public/ assets, or setup/.
// The explicit `slides` and `style` files always win over what the overlay
// brings.
func (m *Slidev) InitDeck(
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
) (*dagger.Directory, error) {
	pkgs := append([]string{"pnpm", "add", "@slidev/cli", "vue", theme}, addons...)

	ctr := m.container().
		WithWorkdir("/deck").
		WithExec([]string{"pnpm", "init"}).
		WithExec(pkgs)

	if extras != nil {
		ctr = ctr.WithDirectory(".", extras)
	}

	ctr = ctr.WithFile("slides.md", slides)

	if style != nil {
		ctr = ctr.WithFile("style.css", style)
	}

	return ctr.Directory("/deck"), nil
}
