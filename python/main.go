// Reusable Dagger module for Python CI/CD workflows.
//
// Provides lint, test, format check, security scan, and Docker image build
// functions for Python projects. Uses ruff for linting/formatting,
// pytest for testing, and bandit for security scanning.

package main

import (
	"context"
	"dagger/python/internal/dagger"
	"strings"
)

type Python struct{}

// Lint runs ruff linter on Python source code
func (m *Python) Lint(
	ctx context.Context,
	// Python source directory
	src *dagger.Directory,
	// +optional
	// +default="3.12-slim"
	pythonVersion string,
	// +optional
	// +default="0.8.6"
	ruffVersion string,
	// +optional
	// +default="src/,tests/"
	// Comma-separated list of paths to lint
	paths string,
) (string, error) {
	args := append([]string{"ruff", "check"}, strings.Split(paths, ",")...)
	return dag.Container().
		From("python:"+pythonVersion).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithMountedCache("/root/.cache/pip", dag.CacheVolume("python-pip-cache")).
		WithExec([]string{"pip", "install", "-q", "ruff==" + ruffVersion}).
		WithExec(args).
		Stdout(ctx)
}

// FormatCheck runs ruff format check (non-modifying) on Python source code
func (m *Python) FormatCheck(
	ctx context.Context,
	// Python source directory
	src *dagger.Directory,
	// +optional
	// +default="3.12-slim"
	pythonVersion string,
	// +optional
	// +default="0.8.6"
	ruffVersion string,
	// +optional
	// +default="src/,tests/"
	// Comma-separated list of paths to check
	paths string,
) (string, error) {
	args := append([]string{"ruff", "format", "--check"}, strings.Split(paths, ",")...)
	return dag.Container().
		From("python:"+pythonVersion).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithMountedCache("/root/.cache/pip", dag.CacheVolume("python-pip-cache")).
		WithExec([]string{"pip", "install", "-q", "ruff==" + ruffVersion}).
		WithExec(args).
		Stdout(ctx)
}

// Test runs pytest on the Python source code
func (m *Python) Test(
	ctx context.Context,
	// Python source directory
	src *dagger.Directory,
	// +optional
	// +default="3.12-slim"
	pythonVersion string,
	// +optional
	// +default="tests/"
	// Path to test directory
	testPath string,
	// +optional
	// +default=""
	// Extra pip install specifier (e.g., ".[dev]" or "-r requirements-dev.txt")
	installExtra string,
) (string, error) {
	install := ".[dev]"
	if installExtra != "" {
		install = installExtra
	}

	return dag.Container().
		From("python:"+pythonVersion).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithMountedCache("/root/.cache/pip", dag.CacheVolume("python-pip-cache")).
		WithExec([]string{"pip", "install", "-q", install}).
		WithExec([]string{"pytest", testPath, "-v", "--tb=short"}).
		Stdout(ctx)
}

// TestWithCoverage runs pytest with coverage reporting
func (m *Python) TestWithCoverage(
	ctx context.Context,
	// Python source directory
	src *dagger.Directory,
	// +optional
	// +default="3.12-slim"
	pythonVersion string,
	// +optional
	// +default="tests/"
	// Path to test directory
	testPath string,
	// +optional
	// +default=""
	// Extra pip install specifier
	installExtra string,
	// +optional
	// +default="src/"
	// Source path for coverage measurement
	coveragePath string,
) (string, error) {
	install := ".[dev]"
	if installExtra != "" {
		install = installExtra
	}

	return dag.Container().
		From("python:"+pythonVersion).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithMountedCache("/root/.cache/pip", dag.CacheVolume("python-pip-cache")).
		WithExec([]string{"pip", "install", "-q", install, "pytest-cov"}).
		WithExec([]string{"pytest", testPath, "-v", "--tb=short",
			"--cov=" + coveragePath, "--cov-report=term-missing"}).
		Stdout(ctx)
}

// SecurityScan runs bandit Python security scanner
func (m *Python) SecurityScan(
	ctx context.Context,
	// Python source directory
	src *dagger.Directory,
	// +optional
	// +default="3.12-slim"
	pythonVersion string,
	// +optional
	// +default="1.8.3"
	banditVersion string,
	// +optional
	// +default="src/"
	// Path to scan
	scanPath string,
) (string, error) {
	return dag.Container().
		From("python:"+pythonVersion).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithMountedCache("/root/.cache/pip", dag.CacheVolume("python-pip-cache")).
		WithExec([]string{"pip", "install", "-q", "bandit==" + banditVersion}).
		WithExec([]string{"bandit", "-r", scanPath, "-f", "txt", "--severity-level", "medium"}).
		Stdout(ctx)
}

// BuildImage builds a Docker image from a Dockerfile
func (m *Python) BuildImage(
	ctx context.Context,
	// Build context directory (must contain Dockerfile)
	src *dagger.Directory,
	// +optional
	// +default="Dockerfile"
	dockerfile string,
	// +optional
	// +default="dev"
	version string,
	// +optional
	// +default="unknown"
	commit string,
	// +optional
	// +default="unknown"
	date string,
) *dagger.Container {
	return src.DockerBuild(dagger.DirectoryDockerBuildOpts{
		Dockerfile: dockerfile,
		BuildArgs: []dagger.BuildArg{
			{Name: "VERSION", Value: version},
			{Name: "COMMIT", Value: commit},
			{Name: "DATE", Value: date},
		},
	})
}

// BuildAndPushImage builds and pushes a Docker image to a registry
func (m *Python) BuildAndPushImage(
	ctx context.Context,
	// Build context directory
	src *dagger.Directory,
	// Full image reference (e.g., ghcr.io/org/image:tag)
	imageRef string,
	// Registry authentication token
	token *dagger.Secret,
	// +optional
	// +default="Dockerfile"
	dockerfile string,
	// +optional
	// +default="dev"
	version string,
	// +optional
	// +default="unknown"
	commit string,
	// +optional
	// +default="unknown"
	date string,
	// +optional
	// +default="ghcr.io"
	registryUrl string,
	// +optional
	// +default=""
	username string,
) (string, error) {
	container := m.BuildImage(ctx, src, dockerfile, version, commit, date)

	if username != "" {
		container = container.WithRegistryAuth(registryUrl, username, token)
	}

	return container.Publish(ctx, imageRef)
}
