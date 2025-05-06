package main

import (
	"context"
	"dagger/hugo/internal/dagger"
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
) (*dagger.Directory, error) {
	// Create base container with Hugo installed
	baseCtr := m.container()

	// Create new site and initialize modules
	ctr := baseCtr.
		WithExec([]string{"hugo", "new", "site", name}).
		WithWorkdir(name).
		WithExec([]string{"hugo", "mod", "init", "github.com/stuttgart-things/docs/k8s-backup"}).
		WithExec([]string{"hugo", "mod", "get", "github.com/joshed-io/reveal-hugo"})

	// Get the initialized site directory
	siteDir := ctr.Directory(".")

	// Add default content
	siteDir = siteDir.WithFile("hugo.toml", config)

	siteDir = siteDir.WithDirectory("/content", content)

	return siteDir, nil
}

func (m *Hugo) container() *dagger.Container {
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	ctr := dag.Container().From(m.BaseImage)

	pkg := "hugo"
	ctr = ctr.WithExec([]string{"apk", "add", "--no-cache", pkg, "go", "git"})
	ctr = ctr.WithEntrypoint([]string{"hugo"})

	return ctr
}

func (m *Hugo) Serve(
	ctx context.Context,
	config *dagger.File,
	content *dagger.Directory,
	// The Port to use
	// +optional
	// +default="1313"
	port string,

) (*dagger.Service, error) {
	// Step 1: Init Hugo site
	siteDir, err := m.InitSite(ctx, "mysite", config, content)
	if err != nil {
		return nil, err
	}

	// Step 2: Start hugo server from generated site
	svc := m.container().
		WithMountedDirectory("/src", siteDir).
		WithWorkdir("/src").
		WithExposedPort(1313).
		AsService(dagger.ContainerAsServiceOpts{
			Args: []string{
				"hugo", "server",
				"--bind", "0.0.0.0",
				"--baseURL", "http://localhost",
				"--port", port,
			},
		})

	return svc, nil
}

// Start and return an HTTP service
func (m *Hugo) HttpService() *dagger.Service {
	return dag.Container().
		From("python").
		WithWorkdir("/srv").
		WithNewFile("index.html", "Hello, world!").
		WithExposedPort(8080).
		AsService(dagger.ContainerAsServiceOpts{Args: []string{"python", "-m", "http.server", "8080"}})
}

// Send a request to an HTTP service and return the response
func (m *Hugo) Get(ctx context.Context) (string, error) {
	return dag.Container().
		From("alpine").
		WithServiceBinding("www", m.HttpService()).
		WithExec([]string{"wget", "-O-", "http://www:8080"}).
		Stdout(ctx)
}
