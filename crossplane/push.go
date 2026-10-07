package main

import (
	"context"
	"crypto/sha256"
	"dagger/crossplane/internal/dagger"
	reg "dagger/crossplane/registry"
	"encoding/hex"
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
// The name is derived from the content, so different credentials never share
// one secret name within a session.
func dockerConfigSecret(configJSON string) *dagger.Secret {
	sum := sha256.Sum256([]byte(configJSON))
	return dag.SetSecret("docker-config-"+hex.EncodeToString(sum[:8]), configJSON)
}
