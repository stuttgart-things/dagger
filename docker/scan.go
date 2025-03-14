package main

import (
	"context"
	"dagger/docker/internal/dagger"
	"fmt"
	"strings"
)

// TrivyScan performs a security scan on a Docker image using its reference
func (m *Docker) TrivyScan(
	ctx context.Context,
	imageRef string, // Fully qualified image reference (e.g., "ttl.sh/my-repo:1.0.0")
	// The registry username
	// +optional
	withRegistryUsername *dagger.Secret,

	// The registry password
	// +optional
	withRegistryPassword *dagger.Secret,
) (string, error) {
	// Configure the Trivy container with registry authentication (if credentials are provided)
	trivyContainer := m.BaseTrivyContainer

	if withRegistryUsername != nil && withRegistryPassword != nil {
		// Get the plaintext username from the secret
		username, err := withRegistryUsername.Plaintext(ctx)
		if err != nil {
			return "", fmt.Errorf("FAILED TO GET REGISTRY USERNAME: %w", err)
		}

		// Authenticate with the registry
		trivyContainer = trivyContainer.WithRegistryAuth(
			strings.Split(imageRef, "/")[0], // Extract registry URL from imageRef
			username,
			withRegistryPassword,
		)
	}

	// Run Trivy scan on the image reference
	return trivyContainer.
		WithExec([]string{"trivy", "image", "--format", "table", imageRef}).
		Stdout(ctx)
}
