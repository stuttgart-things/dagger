package main

import (
	"context"
	"dagger/helm/internal/dagger"
)

// DependencyUpdate updates the dependencies of a chart.
func (m *Helm) DependencyUpdate(
	ctx context.Context,
	src *dagger.Directory) *dagger.Directory {

	projectDir := "/helm"

	chartDir := m.container().
		WithDirectory(projectDir, src).
		WithWorkdir(projectDir).
		WithExec(
			[]string{
				"helm",
				"dependency",
				"update",
			})

	return chartDir.Directory(projectDir)
}
