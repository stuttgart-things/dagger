package main

import (
	"context"
	"dagger/helm/internal/dagger"
)

func (m *Helm) ValidateChart(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	valuesFile *dagger.File,
	// +optional
	registrySecret *dagger.Secret,
	// +optional
	// +default="error"
	severity string,
) (*dagger.File, error) {

	// Render the Helm chart to get the manifests
	renderedManifests, err := m.Render(
		ctx,
		src,
		valuesFile,
		registrySecret,
	)
	if err != nil {
		return nil, err
	}

	// Create a container to run Polaris
	polarisContainer := m.container().
		WithWorkdir("/manifests").
		WithNewFile("rendered.yaml", renderedManifests)

	// Run Polaris and capture the output
	polarisContainer = polarisContainer.
		WithExec([]string{
			"polaris",
			"audit",
			"--audit-path",
			"/manifests/rendered.yaml",
			"--severity",
			severity,
			"--format",
			"json",
			"--output-file",
			"/manifests/validation.json",
		})

	if err != nil {
		return nil, err
	}

	return polarisContainer.File("/manifests/validation.json"), nil
}
