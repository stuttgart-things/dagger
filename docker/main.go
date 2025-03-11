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

// Build permit to build image from Dockerfile
func (m *Docker) Build(

	// the source directory
	source *dagger.Directory,

	// The dockerfile path
	// +optional
	// +default="Dockerfile"
	dockerfile string,

	// Set extra directories
	// +optional
	withDirectories []*dagger.Directory,
) *ImageBuild {

	if m.BuildContainer != nil {
		return &ImageBuild{
			Container: m.BuildContainer,
		}
	}

	for _, directory := range withDirectories {
		source = source.WithDirectory(fmt.Sprintf("%s", directory), directory)
	}

	return &ImageBuild{
		Container: source.DockerBuild(
			dagger.DirectoryDockerBuildOpts{
				Dockerfile: dockerfile,
			},
		),
	}
}

// BuildAndPush combines the Build and Push functionalities
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
	insecure bool,
) (string, error) {

	// Step 1: Build the Docker image
	imageBuild := m.Build(source, dockerfile, withDirectories)

	// Step 2: Push the Docker image
	return imageBuild.Push(ctx, repositoryName, version, withRegistryUsername, withRegistryPassword, registryUrl, insecure)
}
