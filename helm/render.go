package main

import (
	"context"
	"dagger/helm/internal/dagger"
)

func (m *Helm) Render(
	ctx context.Context,
	src *dagger.Directory,
	// Values file
	// +optional
	// +default="values.yaml"
	valuesFile string,
) (string, error) {

	helmContainer := m.container()

	updatedChart := m.DependencyUpdate(ctx, src)

	args := []string{"helm", "template", "."}

	if valuesFile != "" {
		args = append(args, "-f", valuesFile)
	}

	renderedManifests, err := helmContainer.
		WithDirectory("/helm", updatedChart).
		WithWorkdir("/helm").
		WithExec(args).
		Stdout(ctx)
	if err != nil {
		return "", err
	}

	return renderedManifests, nil
}
