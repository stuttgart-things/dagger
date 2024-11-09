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

// RunPipeline method: Orchestrates running both Lint and Build steps
func (m *Helm) RunPipeline(ctx context.Context, src *dagger.Directory) {

	fmt.Println("Running linting...")

	lint, err := m.Lint(ctx, src)
	if err != nil {
		fmt.Println("Error running linter: ", err)
	}

	fmt.Print("Lint: ", lint)

	fmt.Println("Running templating...")

	templatedChart := m.Template(ctx, src)
	fmt.Println("Templated Chart: ", templatedChart)

}

// Lint
func (m *Helm) Lint(ctx context.Context, src *dagger.Directory) (string, error) {
	return dag.HelmDisaster37().Lint(ctx, src)
}

// Template
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
