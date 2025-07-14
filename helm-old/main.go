package main

import (
	"context"
	"dagger/helm/internal/dagger"

	"fmt"
)

type Helm struct {
	// +optional
	// +default="cgr.dev/chainguard/wolfi-base:latest"
	BaseImage string
}

// RunPipeline orchestrates all pipeline steps
func (m *Helm) RunPipeline(
	ctx context.Context,
	src *dagger.Directory,
	values *dagger.File) {

	// STAGE 0: DEPENDENCY UPDATE
	fmt.Println("RUNNING CHART DEPENDENCY UPDATE...")
	chartDirectory := m.DependencyUpdate(ctx, src)

	// STAGE 0: LINT
	fmt.Println("RUNNING CHART LINTING...")
	lint, err := m.Lint(ctx, chartDirectory)
	if err != nil {
		fmt.Println("Error running linter: ", err)
	}
	fmt.Print("LINT RESULT: ", lint)

	// STAGE 0: TEST-TEMPLATE
	fmt.Println("RUNNING TEST-TEMPLATING OF CHART...")
	templatedChart := m.Render(ctx, chartDirectory, values)
	fmt.Println("TEMPLATED MANIFESTS: ", templatedChart)

	// STAGE 1: PACKAGE CHART
	fmt.Println("RUNNING CHART PACKAGING...")
	packedChart := m.Package(ctx, chartDirectory)
	fmt.Println("PACKAGED CHART: ", packedChart)
}
