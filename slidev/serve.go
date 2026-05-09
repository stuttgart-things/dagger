package main

import (
	"context"
	"dagger/slidev/internal/dagger"
	"strconv"
)

// Serve runs `slidev` as a long-lived service bound to 0.0.0.0 on the given
// port. Inputs mirror InitDeck so the caller only ships markdown + css.
func (m *Slidev) Serve(
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
	// +default="3030"
	port string,
) (*dagger.Service, error) {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	deckDir, err := m.InitDeck(ctx, slides, style, theme, addons, extras)
	if err != nil {
		return nil, err
	}

	return m.container().
		WithMountedDirectory("/deck", deckDir).
		WithWorkdir("/deck").
		WithExposedPort(portNum).
		AsService(dagger.ContainerAsServiceOpts{
			Args: []string{
				"pnpm", "exec", "slidev", "slides.md",
				"--port", port,
				"--remote",
			},
		}), nil
}
