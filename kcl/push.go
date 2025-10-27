package main

import (
	"context"
	"dagger/kcl/internal/dagger"
	"fmt"
)

func (m *Kcl) PushModule(
	ctx context.Context,
	src *dagger.Directory,
	address string,
	moduleName string,
	// +optional
	// +default="GITHUB_USER"
	userName string,
	// +optional
	user *dagger.Secret,
	// +optional
	// +default="GITHUB_TOKEN"
	passwordName string,
	// +optional
	password *dagger.Secret,
) (string, error) {

	// Extract registry from address (e.g., "oci://ghcr.io/stuttgart-things" -> "ghcr.io")
	registry := "ghcr.io"
	if len(address) > 6 && address[:6] == "oci://" {
		registry = address[6:]
		// Get just the host part (before first /)
		for i, c := range registry {
			if c == '/' {
				registry = registry[:i]
				break
			}
		}
	}

	return m.container().
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithSecretVariable(userName, user).
		WithSecretVariable(passwordName, password).
		// Login to registry first
		WithExec([]string{"sh", "-c", fmt.Sprintf("kcl registry login %s -u $%s -p $%s", registry, userName, passwordName)}).
		// Then push the module
		WithExec([]string{"kcl", "mod", "push", address + "/" + moduleName}).
		Stdout(ctx)
}
