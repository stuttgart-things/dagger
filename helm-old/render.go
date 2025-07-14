package main

import (
	"context"
	"dagger/helm/internal/dagger"
	"fmt"
)

// Renders a chart as Kubernetes manifests
func (m *Helm) Render(
	ctx context.Context,
	chart *dagger.Directory,
	values *dagger.File) (templatedChart string) {

	dependencyUpdatedChartDir := m.DependencyUpdate(ctx, chart)

	projectDir := "/project"
	valuesFileName := "test-values.yaml"

	templatedChart, err := m.HelmContainer.
		WithDirectory(projectDir, dependencyUpdatedChartDir).
		WithFile(projectDir+"/"+valuesFileName, values).
		WithWorkdir(projectDir).
		WithExec(
			[]string{"helm", "template", ".", "--values", valuesFileName}).
		Stdout(ctx)

	if err != nil {
		fmt.Println("ERROR RUNNING VERSION COMMAND: ", err)
	}

	return templatedChart
}
