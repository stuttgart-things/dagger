package main

import (
	"context"
	"dagger/hugo/internal/dagger"
	"strconv"
)

type Hugo struct {
	// +optional
	// +default="cgr.dev/chainguard/wolfi-base:latest"
	BaseImage string
}

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
		WithExec([]string{"hugo", "mod", "get", theme})

	// GET THE INITIALIZED SITE DIRECTORY
	siteDir := ctr.Directory(".")
	siteDir = siteDir.WithFile("hugo.toml", config)
	siteDir = siteDir.WithDirectory("/content", content)

	return siteDir, nil
}

func (m *Hugo) container() *dagger.Container {
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	ctr := dag.Container().From(m.BaseImage)
	ctr = ctr.WithExec([]string{"apk", "add", "--no-cache", "hugo", "go", "git"})
	ctr = ctr.WithEntrypoint([]string{"hugo"})

	return ctr
}

func (m *Hugo) Serve(
	ctx context.Context,
	config *dagger.File,
	content *dagger.Directory,
	// The project name
	// +optional
	// +default="hugo"
	name string,
	// The bindAddr to use
	// +optional
	// +default="1313"
	bindAddr string,
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

	// STEP 2: START HUGO SERVER FROM GENERATED SITE
	svc := m.container().
		WithMountedDirectory("/src", siteDir).
		WithWorkdir("/src").
		WithExposedPort(portNumber).
		AsService(dagger.ContainerAsServiceOpts{
			Args: []string{
				"hugo", "server",
				"--bind", bindAddr,
				"--baseURL", baseURL,
				"--port", port,
			},
		})

	return svc, nil
}
