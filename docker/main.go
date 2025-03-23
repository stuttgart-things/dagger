// A generated module for Docker functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/docker/internal/dagger"
	"fmt"
)

type Docker struct {
	// +private
	BaseHadolintContainer *dagger.Container
	BaseTrivyContainer    *dagger.Container

	// +private
	BuildContainer *dagger.Container
}

func New(
	// base hadolint container
	// It need contain hadolint
	// +optional
	baseHadolintContainer *dagger.Container,

	// base hadolint container
	// It need contain hadolint
	// +optional
	baseTrivyContainer *dagger.Container,

	// The external build of container
	// Usefull when need build args
	// +optional
	buildContainer *dagger.Container,
) *Docker {
	image := &Docker{
		BuildContainer: buildContainer,
	}

	if baseHadolintContainer != nil {
		image.BaseHadolintContainer = baseHadolintContainer
	} else {
		image.BaseHadolintContainer = image.GetBaseHadolintContainer()
	}

	if baseTrivyContainer != nil {
		image.BaseTrivyContainer = baseTrivyContainer
	} else {
		image.BaseTrivyContainer = image.GetTrivyContainer()
	}

	return image
}

// GetBaseHadolintContainer return the default image for hadolint
func (m *Docker) GetBaseHadolintContainer() *dagger.Container {
	return dag.Container().
		From("ghcr.io/hadolint/hadolint:2.12.0")
}

// GetBaseHadolintContainer return the default image for hadolint
func (m *Docker) GetTrivyContainer() *dagger.Container {
	return dag.Container().
		From("aquasec/trivy:0.60.0")
}

func (m *Docker) BuildAndPush(
	ctx context.Context,
	// The source directory
	source *dagger.Directory,
	// The repository name
	repositoryName string,
	// The version
	version string,
	// The registry username
	// +optional
	withRegistryUsername *dagger.Secret,
	// The registry password
	// +optional
	withRegistryPassword *dagger.Secret,
	// The registry URL
	registryUrl string,
	// The Dockerfile path
	// +optional
	// +default="Dockerfile"
	dockerfile string,
	// Set extra directories
	// +optional
	withDirectories []*dagger.Directory,
) (string, error) {
	// Step 1: Run linting
	lintOutput, err := m.Lint(ctx, source, dockerfile, "error") // Use default threshold
	if err != nil {
		return "", fmt.Errorf("linting failed: %w", err)
	}
	fmt.Println("Linting Results:")
	fmt.Println(lintOutput)

	// Step 2: Build the image
	builtImage := m.Build(source, dockerfile, withDirectories)
	if builtImage == nil {
		return "", fmt.Errorf("build failed: builtImage is nil")
	}

	// Step 3: Push the image to the registry
	imageRef := fmt.Sprintf("%s/%s:%s", registryUrl, repositoryName, version)
	_, err = builtImage.Push(ctx, repositoryName, version, withRegistryUsername, withRegistryPassword, registryUrl)
	if err != nil {
		return "", fmt.Errorf("push failed: %w", err)
	}

	// Step 4: Run Trivy scan on the built image
	trivyOutput, err := m.TrivyScan(ctx, imageRef, withRegistryUsername, withRegistryPassword)
	if err != nil {
		return "", fmt.Errorf("Trivy scan failed: %w", err)
	}
	fmt.Println("Trivy Scan Results:")
	fmt.Println(trivyOutput)

	// Return success along with the linting and Trivy results
	return fmt.Sprintf("Successfully built and pushed %s\nLinting Results:\n%s\nTrivy Scan Results:\n%s", imageRef, lintOutput, trivyOutput), nil
}
