package main

import (
	"context"
	"dagger/helm/internal/dagger"
)

func (m *Helm) Render(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	valuesFile *dagger.File, // accept dagger.File instead of string
) (string, error) {

	helmContainer := m.container()

	updatedChart := m.DependencyUpdate(ctx, src)

	projectDir := "/helm"
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
