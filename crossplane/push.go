package main

import (
	"context"
	"crypto/rand"
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
		WithMountedSecret("/root/.docker/config.json", dockerConfigSecret(configJSON)).
		WithDirectory("/src", dirWithPackage).
		WithWorkdir("/src").
		WithExec([]string{"crossplane", "xpkg", "push", destination}).
		Stdout(ctx)

	if err != nil {
		panic(err)
	}

	return status
}

// dockerConfigSecret wraps a docker config.json as a Secret, so it is mounted
// instead of passed as an op argument, which --progress plain prints (#318).
// The name is random per call: secret names can appear in traces, and a name
// derived from the content (an unsalted hash of the base64 credential) lets a
// guessed password be confirmed offline.
func dockerConfigSecret(configJSON string) *dagger.Secret {
	return dag.SetSecret("docker-config-"+rand.Text(), configJSON)
}
