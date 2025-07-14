package main

import (
	"context"
	"dagger/helm/internal/dagger"
	reg "dagger/helm/registry"
	"fmt"
)

// DEPENDENCYUPDATE UPDATES THE DEPENDENCIES OF A CHART
func (m *Helm) DependencyUpdate(
	ctx context.Context,
	src *dagger.Directory) *dagger.Directory {

	projectDir := "/helm"

	chartDir := m.container(m.BaseImage).
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

// PACKAGES A CHART INTO A VERSIONED CHART ARCHIVE FILE (.tgz)
func (m *Helm) Package(
	ctx context.Context,
	src *dagger.Directory) *dagger.File {
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
