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
	// Configure the Trivy container
	trivyContainer := m.BaseTrivyContainer

	if withRegistryUsername != nil && withRegistryPassword != nil {
		trivyContainer = trivyContainer.
			WithSecretVariable("TRIVY_USERNAME", withRegistryUsername).
			WithSecretVariable("TRIVY_PASSWORD", withRegistryPassword)

		// Output a successful login message
		fmt.Printf("SUCCESSFULLY SET TRIVY CREDENTIALS FOR USER: %s\n", withRegistryUsername)
	}

	// Run Trivy scan on the image reference
	return trivyContainer.
		WithExec([]string{"trivy", "image", "--format", "table", imageRef}).
		Stdout(ctx)
}

// extractRegistryUrl extracts the registry URL from an image reference
func extractRegistryUrl(imageRef string) string {
	// Split the image reference into parts
	parts := strings.Split(imageRef, "/")
	if len(parts) == 1 {
		// No registry specified, default to Docker Hub
		return "docker.io"
	}

	// Check if the first part contains a port (e.g., "my-registry.com:5000")
	if strings.Contains(parts[0], ":") || strings.Contains(parts[0], ".") {
		// Assume the first part is the registry URL
		return parts[0]
	}

	// If the first part is not a valid registry URL, default to Docker Hub
	return "docker.io"
}
