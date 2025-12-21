package main

import (
	"context"
	"dagger/crossplane/internal/dagger"
	reg "dagger/crossplane/registry"
)

// Push Crossplane Package
func (m *Crossplane) Push(
	ctx context.Context,
	src *dagger.Directory,
	registry string,
	username string,
	password *dagger.Secret,
	destination string,
) string {

	// ✅ Ensure container is initialized
	if m.XplaneContainer == nil {
		m.XplaneContainer = m.GetXplaneContainer(ctx)
	}

	dirWithPackage := m.Package(ctx, src)

	passwordPlaintext, err := password.Plaintext(ctx)
	if err != nil {
		panic(err)
	}

	configJSON, err := reg.CreateDockerConfigJSON(username, passwordPlaintext, registry)
	if err != nil {
		panic(err)
	}

	status, err := m.XplaneContainer.
		WithNewFile("/root/.docker/config.json", configJSON).
		WithDirectory("/src", dirWithPackage).
		WithWorkdir("/src").
		WithExec([]string{"crossplane", "xpkg", "push", destination}).
		Stdout(ctx)

	if err != nil {
		panic(err)
	}

	return status
}
