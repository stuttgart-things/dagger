package main

import (
	"context"
	"dagger/hugo/internal/dagger"
	"fmt"
	"strconv"
)

func (m *Hugo) Serve(
	ctx context.Context,
	config *dagger.File,
	content *dagger.Directory,
	// The project name
	// +optional
	// +default="hugo"
	name string,
	// The base url to use
	// +optional
	// +default="0.0.0.0"
	baseURL string,
	// The Port to use
	// +optional
	// +default="1313"
	port string,
	// The Theme to use
	// +optional
	// +default="github.com/joshed-io/reveal-hugo"
	theme string,
) (*dagger.Service, error) {

	// STEP 0: GET PORT AS INT
	portNumber, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	// STEP 1: INIT HUGO SITE
	siteDir, err := m.InitSite(ctx, name, config, content, theme)
	if err != nil {
		return nil, err
	}

	// Patch vendored theme to fix Hugo v0.157+ compatibility (.Site.Author removed)
	headPath := fmt.Sprintf("_vendor/%s/layouts/partials/layout/head.html", theme)
	patchedDir := m.container().
		WithMountedDirectory("/src", siteDir).
		WithWorkdir("/src").
		WithExec([]string{
			"sh", "-c",
			fmt.Sprintf(`[ -f "%s" ] && sed -i 's/\.Site\.Author\.name/site.Title/g' "%s" || true`, headPath, headPath),
		}).
		Directory("/src")

	// STEP 2: START HUGO SERVER FROM GENERATED SITE
	svc := m.container().
		WithMountedDirectory("/src", patchedDir).
		WithWorkdir("/src").
		WithExposedPort(portNumber).
		AsService(dagger.ContainerAsServiceOpts{
			Args: []string{
				"hugo", "server",
				"--bind", baseURL,
				"--baseURL", baseURL,
				"--port", port,
			},
		})

	return svc, nil
}
