package main

import (
	"context"
	"dagger/helm/internal/dagger"
	reg "dagger/helm/registry"
	"fmt"
)

func (m *Helm) Push(
	ctx context.Context,
	src *dagger.Directory,
	registry string,
	repository string,
	username string,
	password *dagger.Secret,
) (string, error) {

	projectDir := "/helm"
	helmContainer := m.container()

	passwordPlaintext, err := password.Plaintext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read password secret: %w", err)
	}

	configJSON, err := reg.CreateDockerConfigJSON(username, passwordPlaintext, registry)
	if err != nil {
		return "", fmt.Errorf("failed to create Docker config.json: %w", err)
	}

	// Package chart
	packedChart, err := m.Package(ctx, src)
	if err != nil {
		return "", fmt.Errorf("failed to package chart: %w", err)
	}

	archiveFileName, err := packedChart.Name(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get packaged chart filename: %w", err)
	}

	// Push the chart
	result, err := helmContainer.
		WithFile(projectDir+"/"+archiveFileName, packedChart).
		WithNewFile("/root/.docker/config.json", configJSON).
		WithDirectory(projectDir, src).
		WithWorkdir(projectDir).
		WithExec([]string{
			"helm",
			"push",
			projectDir + "/" + archiveFileName,
			"oci://" + registry + "/" + repository,
		}).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to push chart: %w", err)
	}

	return result, nil
}
