package main

import (
	"context"
	"dagger/helm/internal/dagger"
)

func (m *Helm) Render(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	valuesFile *dagger.File,
	// +optional
	registrySecret *dagger.Secret,
) (string, error) {

	projectDir := "/helm"
	dockerConfigPath := "/root/.docker/config.json"

	helmContainer := m.container()

	// CONDITIONALLY MOUNT THE SECRET IF PROVIDED
	if registrySecret != nil { // pragma: allowlist secret
		helmContainer = helmContainer.WithMountedSecret(dockerConfigPath, registrySecret)
	}

	updatedChart := m.DependencyUpdate(ctx, src)

	helmContainer = helmContainer.
		WithDirectory(projectDir, updatedChart).
		WithWorkdir(projectDir)

	args := []string{"helm", "template", projectDir}

	if valuesFile != nil {
		// Mount the values file into the container
		helmContainer = helmContainer.WithFile(projectDir+"/values-file.yaml", valuesFile)
		args = append(args, "--values", "values-file.yaml")
	}

	renderedManifests, err := helmContainer.
		WithExec(args).
		Stdout(ctx)
	if err != nil {
		return "", err
	}

	return renderedManifests, nil
}

func (m *Helm) RenderHelmfile(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	// +default="./"
	pathHelmfile string,
	// +optional
	// +default="helmfile.yaml"
	helmfileName string,
	// +optional
	registrySecret *dagger.Secret,
) (string, error) {

	projectDir := "/helmfiles"
	dockerConfigPath := "/root/.docker/config.json"

	helmContainer := m.container().
		WithDirectory(projectDir, src.Directory(pathHelmfile)).
		WithWorkdir(projectDir)

	// CONDITIONALLY MOUNT THE SECRET IF PROVIDED
	if registrySecret != nil { // pragma: allowlist secret
		helmContainer = helmContainer.WithMountedSecret(dockerConfigPath, registrySecret)
	}

	args := []string{
		"helmfile",
		"-f",
		helmfileName,
		"template"}

	renderedManifests, err := helmContainer.
		WithExec(args).
		Stdout(ctx)

	if err != nil {
		return "", err
	}

	return renderedManifests, nil
}
