// Docker module for container image handling
//
// This module provides Docker-related functionality using Dagger. It supports
// building Docker images (optionally with extra directories), performing lint
// checks with Hadolint, and pushing built images to a container registry.
//
// The module allows you to inject a custom base container for Hadolint, or
// a custom container for advanced build scenarios that require specific
// build arguments or tooling.
//
// Typical usage includes:
//   - Building a Docker image from a provided source and Dockerfile
//   - Optionally attaching extra directories to the build context
//   - Authenticating and pushing the image to a Docker registry
//
// This module is designed to be used as part of a CI/CD pipeline, either via
// the Dagger CLI or any supported Dagger SDK.

package main

import (
	"context"
	"dagger/docker/internal/dagger"
	"fmt"
)

type Docker struct {
	// +private
	BaseHadolintContainer *dagger.Container
	// +private
	BuildContainer *dagger.Container
}

func New(
	// base hadolint container
	// It need contain hadolint
	// +optional
	baseHadolintContainer *dagger.Container,
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

	return image
}

// GetBaseHadolintContainer return the default image for hadolint
func (m *Docker) GetBaseHadolintContainer() *dagger.Container {
	return dag.Container().
		From("ghcr.io/hadolint/hadolint:2.12.0")
}

func (m *Docker) BuildAndPush(
	ctx context.Context,
	// The source directory
	source *dagger.Directory,
	// The repository name
	repositoryName string,
	// tag
	tag string,
	// The registry username
	// +optional
	registryUsername *dagger.Secret,
	// The registry password
	// +optional
	registryPassword *dagger.Secret,
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

	// STEP 2: BUILD THE IMAGE
	builtImage := m.Build(source, dockerfile, withDirectories)
	if builtImage == nil {
		return "", fmt.Errorf("build failed: builtImage is nil")
	}

	// STEP 3: PUSH THE IMAGE TO THE REGISTRY
	imageRef := fmt.Sprintf("%s/%s:%s", registryUrl, repositoryName, tag)
	_, err := builtImage.Push(ctx, repositoryName, tag, registryUsername, registryPassword, registryUrl)
	if err != nil {
		return "", fmt.Errorf("push failed: %w", err)
	}

	return fmt.Sprintf("Successfully built and pushed", imageRef), nil
}
