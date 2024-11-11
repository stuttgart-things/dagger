/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"context"
	"dagger/helm/internal/dagger"
	"fmt"
)

type Helm struct {
	HelmContainer *dagger.Container
}

func New(
	// helm container
	// It need contain helm
	// +optional
	helmContainer *dagger.Container,

) *Helm {
	helm := &Helm{}

	if helmContainer != nil {
		helm.HelmContainer = helmContainer
	} else {
		helm.HelmContainer = helm.GetHelmContainer()
	}
	return helm
}

// GetHelmContainer return the default image for helm
func (m *Helm) GetHelmContainer() *dagger.Container {
	return dag.Container().
		From("alpine/helm:3.14.3")
}

// RunPipeline orchestrates all pipeline steps
func (m *Helm) RunPipeline(ctx context.Context, src *dagger.Directory) {

	// STAGE 0: LINT
	fmt.Println("RUNNING CHART LINTING...")
	lint, err := m.Lint(ctx, src)
	if err != nil {
		fmt.Println("Error running linter: ", err)
	}
	fmt.Print("LINT RESULT: ", lint)

	// STAGE 0: TEST-TEMPLATE
	fmt.Println("RUNNING TEST-TEMPLATING OF CHART...")
	templatedChart := m.Template(ctx, src)
	fmt.Println("TEMPLATED MANIFESTS: ", templatedChart)

	// STAGE 1: PACKAGE CHART
	fmt.Println("RUNNING CHART PACKAGING...")
	packedChart := m.Package(ctx, src)
	fmt.Println("PACKAGED CHART: ", packedChart)
}

// Lints a chart
func (m *Helm) Lint(ctx context.Context, src *dagger.Directory) (string, error) {
	return dag.HelmDisaster37().Lint(ctx, src)
}

// Packages a chart into a versioned chart archive file (.tgz)
func (m *Helm) Package(ctx context.Context, src *dagger.Directory) *dagger.File {
	return dag.HelmOci().Package(src)
}

// Renders a chart as Kubernetes manifests
func (m *Helm) Template(ctx context.Context, src *dagger.Directory) (templatedChart string) {

	templatedChart, err := m.HelmContainer.
		WithDirectory("/project", src).
		WithWorkdir("/project").
		WithExec(
			[]string{"helm", "template", "."}).
		Stdout(ctx)

	fmt.Println(err)

	return templatedChart
}
