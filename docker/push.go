package main

import (
	"bytes"
	"context"
	"dagger/docker/internal/dagger"
	"fmt"

	"emperror.dev/errors"
	"github.com/coreos/go-semver/semver"
)

type ImageBuild struct {
	// +private
	Container *dagger.Container
}

// GetContainer permit to get the container
func (m *ImageBuild) GetContainer() *dagger.Container {
	return m.Container
}

// Push permits pushing an image to a registry, with support for insecure registries
func (m *ImageBuild) Push(
	ctx context.Context,
	// The repository name
	repositoryName string,
	// The tag
	tag string,
	// The registry username
	// +optional
	registryUsername *dagger.Secret,
	// The registry password
	// +optional
	registryPassword *dagger.Secret,
	// The registry URL
	registryUrl string,
) (string, error) {

	// Mitigate semver tag
	semtag, err := semver.NewVersion(tag)
	if err == nil {
		var buffer bytes.Buffer

		fmt.Fprintf(&buffer, "%d.%d.%d", semtag.Major, semtag.Minor, semtag.Patch)

		if semtag.PreRelease != "" {
			fmt.Fprintf(&buffer, "-%s", semtag.PreRelease)
		}
		if semtag.Metadata != "" {
			fmt.Fprintf(&buffer, "-%s", semtag.Metadata)
		}

		tag = buffer.String()
	}

	// Configure registry authentication (if credentials are provided)
	if registryUsername != nil && registryPassword != nil {
		username, err := registryUsername.Plaintext(ctx)
		if err != nil {
			return "", errors.Wrap(err, "Error when getting registry username")
		}
		m.Container = m.Container.WithRegistryAuth(registryUrl, username, registryPassword)
	}

	// Publish the image
	return m.Container.Publish(
		ctx,
		fmt.Sprintf(
			"%s/%s:%s",
			registryUrl,
			repositoryName,
			tag),
	)
}
