package main

import (
	"context"
	"dagger/crossplane/internal/dagger"
	reg "dagger/crossplane/registry"
	"fmt"
)

// Push Crossplane Package
func (m *Crossplane) Push(
	ctx context.Context,
	src *dagger.Directory,
	// +default="ghcr.io"
	registry string,
	username string,
	password *dagger.Secret,
	destination string) string {

	dirWithPackage := m.Package(ctx, src)

	passwordPlaintext, err := password.Plaintext(ctx)

	configJSON, err := reg.CreateDockerConfigJSON(username, passwordPlaintext, registry)
	if err != nil {
		fmt.Printf("ERROR CREATING DOCKER config.json: %v\n", err)
	}

	status, err := m.XplaneContainer.
		WithNewFile("/root/.docker/config.json", configJSON).
		WithDirectory("/src", dirWithPackage).
		WithWorkdir("/src").
		WithExec([]string{"crossplane", "xpkg", "push", destination}).
		Stdout(ctx)

	if err != nil {
		fmt.Println("ERROR PUSHING PACKAGE: ", err)
	}

	fmt.Println("PACKAGE STATUS: ", status)

	return status
}
