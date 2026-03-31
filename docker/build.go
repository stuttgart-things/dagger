package main

import (
	"dagger/docker/internal/dagger"
)

// Build permit to build image from Dockerfile
func (m *Docker) Build(
	// the source directory
	src *dagger.Directory,
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
		src = src.WithDirectory(".", directory)
	}

	return &ImageBuild{
		Container: src.DockerBuild(
			dagger.DirectoryDockerBuildOpts{
				Dockerfile: dockerfile,
			},
		),
	}
}
