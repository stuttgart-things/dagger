package main

import (
	"dagger/docker/internal/dagger"
	"fmt"
)

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
