package main

import (
	"context"
	"dagger/go/internal/dagger"
	"fmt"
)

func (m *Go) Test(
	ctx context.Context,
	src *dagger.Directory,
	// Go version to use for testing
	// +optional
	// +default="1.24.4"
	goVersion string,
	// Test arguments to pass to `go test`
	// +optional
	// +default="./..."
	testArg string, // Arguments for `go test`
) (string, error) {
	// Create a container with the specified Go version
	container := dag.Container().
		From(fmt.Sprintf("golang:%s", goVersion)). // Use the specified Go version
		WithDirectory("/src", src).
		WithWorkdir("/src")

	// Run tests with both verbose output AND coverage in a single pass
	// This is more efficient than running tests twice
	testOutput, err := container.
		WithExec([]string{
			"go",
			"test",
			"-v",
			"-cover",
			testArg}).
		Stdout(ctx)

	if err != nil {
		return "", fmt.Errorf("error running tests: %w", err)
	}

	return testOutput, nil
}
