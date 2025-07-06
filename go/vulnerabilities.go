package main

import (
	"context"
	"dagger/go/internal/dagger"
)

func (m *Go) SecurityScan(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	// +default="2.22.5"
	secureGoVersion string,
) (*dagger.File, error) {
	// Create a container with gosec installed
	container := dag.Container().
		From("securego/gosec:"+secureGoVersion). // Use the official gosec image
		WithDirectory("/src", src).
		WithWorkdir("/src")

	// Run gosec to scan the source code and write the output to a file
	reportPath := "/src/security-report.txt"
	container = container.
		WithExec([]string{"sh", "-c", "gosec ./... > " + reportPath + " || true"}) // Ignore exit code

	// Get the security report file
	reportFile := container.File(reportPath)

	// Return the security report file
	return reportFile, nil
}
