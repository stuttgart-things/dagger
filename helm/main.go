/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"context"
	"dagger/helm/internal/dagger"
	reg "dagger/helm/registry"

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
func (m *Helm) RunPipeline(ctx context.Context, src *dagger.Directory, values *dagger.File) {

	// STAGE 0: LINT
	fmt.Println("RUNNING CHART LINTING...")
	lint, err := m.Lint(ctx, src)
	if err != nil {
		fmt.Println("Error running linter: ", err)
	}
	fmt.Print("LINT RESULT: ", lint)

	// STAGE 0: TEST-TEMPLATE
	fmt.Println("RUNNING TEST-TEMPLATING OF CHART...")
	templatedChart := m.Render(ctx, src, values)
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

// PUSH HELM CHART TO TARGET REGISTRY
func (m *Helm) Push(
	ctx context.Context,
	// +default="ghcr.io"
	registry string,
	repository string,
	username string,
	password *dagger.Secret,
	src *dagger.Directory) string {

	passwordPlaintext, err := password.Plaintext(ctx)
	projectDir := "/helm"

	configJSON, err := reg.CreateDockerConfigJSON(username, passwordPlaintext, registry)
	if err != nil {
		fmt.Printf("ERROR CREATING DOCKER config.json: %v\n", err)
	}

	// PACKAGE CHART
	packedChart := m.Package(ctx, src)

	archiveFileName, err := packedChart.Name(ctx)
	if err != nil {
		fmt.Println("ERROR GETTING ARCHIVE NAME: ", err)
	}

	status, err := m.HelmContainer.
		WithFile(projectDir+"/"+archiveFileName, packedChart).
		WithNewFile("/root/.docker/config.json", configJSON).
		WithDirectory(projectDir, src).
		WithWorkdir(projectDir).
		WithExec(
			[]string{"helm", "push", projectDir + "/" + archiveFileName, "oci://" + registry + "/" + repository}).
		Stdout(ctx)

	if err != nil {
		fmt.Println("ERROR PUSHING CHART: ", err)
	}

	return status
}

// Renders a chart as Kubernetes manifests
func (m *Helm) Render(ctx context.Context, chart *dagger.Directory, values *dagger.File) (templatedChart string) {

	projectDir := "/project"
	valuesFileName := "test-values.yaml"

	templatedChart, err := m.HelmContainer.
		WithDirectory(projectDir, chart).
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
