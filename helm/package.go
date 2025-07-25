package main

import (
	"context"
	"dagger/helm/internal/dagger"
	"fmt"
)

// Package updates the dependencies of a chart and packages the Helm chart.
func (m *Helm) Package(
	ctx context.Context,
	src *dagger.Directory) (*dagger.File, error) {

	helmContainer := m.container()

	updatedChart := m.DependencyUpdate(ctx, src)

	projectDir := "/helm"

	chartDir := helmContainer.
		WithDirectory(projectDir, updatedChart).
		WithWorkdir(projectDir).
		WithExec([]string{
			"helm",
			"package",
			"."})

	// List files to find the packaged .tgz
	files, err := chartDir.Directory(projectDir).Entries(ctx)
	if err != nil {
		return nil, err
	}

	var chartFile string
	for _, f := range files {
		if len(f) > 4 && f[len(f)-4:] == ".tgz" {
			chartFile = f
			break
		}
	}

	if chartFile == "" {
		return nil, fmt.Errorf("packaged chart (.tgz) not found in %s", projectDir)
	}

	return chartDir.File(fmt.Sprintf("%s/%s", projectDir, chartFile)), nil
}
