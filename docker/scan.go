package main

import (
	"context"
)

// TrivyScan performs a security scan on a Docker image using its reference
func (m *Docker) TrivyScan(
	ctx context.Context,
	imageRef string, // Fully qualified image reference (e.g., "ttl.sh/my-repo:1.0.0")
) (string, error) {
	// Run Trivy scan on the image reference
	return m.BaseTrivyContainer.
		WithExec([]string{"trivy", "image", "--format", "table", imageRef}).
		Stdout(ctx)
}
