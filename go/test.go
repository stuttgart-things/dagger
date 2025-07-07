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

	testOutput, err := container.
		WithExec([]string{
			"go",
			"test",
			"-v",
			testArg}).
		Stdout(ctx)

	// RUN GO TESTS WITH COVERAGE
	coverageOutput, err := container.
		WithExec([]string{
			"go",
			"test",
			"-cover",
			testArg}).
		Stdout(ctx)

	output := testOutput + "\n" + coverageOutput

	if err != nil {
		return "", fmt.Errorf("error running tests: %w", err)
	}

	return output, nil
}
