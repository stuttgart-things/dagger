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
	return m.installDeck(m.container(), slides, style, theme, addons, extras, nil).
		Directory("/deck"), nil
}

// installDeck is the shared scaffolding pipeline used by InitDeck, Build,
// Serve and Export. It is lowercase so the Dagger SDK does not expose it as
// a module function — callers pass in the base container (which can be
// further configured, e.g. with Chromium for Export) and any extra dev
// dependencies to install alongside @slidev/cli.
func (m *Slidev) installDeck(
	ctr *dagger.Container,
	slides *dagger.File,
	style *dagger.File,
	theme string,
	addons []string,
	extras *dagger.Directory,
	extraDev []string,
) *dagger.Container {
	pkgs := append([]string{"pnpm", "add", "@slidev/cli", "vue", theme}, addons...)

	c := ctr.WithWorkdir("/deck").
		WithExec([]string{"pnpm", "init"}).
		WithExec(pkgs)

	if len(extraDev) > 0 {
		devPkgs := append([]string{"pnpm", "add", "-D", "--ignore-scripts"}, extraDev...)
		c = c.WithExec(devPkgs)
	}

	if extras != nil {
		c = c.WithDirectory(".", extras)
	}

	c = c.WithFile("slides.md", slides)

	if style != nil {
		c = c.WithFile("style.css", style)
	}

	return c
}
