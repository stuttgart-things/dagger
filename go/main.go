/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

// https://github.com/dagger/dagger/pull/5833/files#diff-42807a87b4d8f4c8adb3861609de1a2a6a6158cf11b00b9b1b342c0a23f1bc03

package main

import (
	"context"
	"dagger/go/internal/dagger"
	"fmt"
)

type Go struct {
	Src             *dagger.Directory
	GoLangContainer *dagger.Container
	KoContainer     *dagger.Container
}

// GetGoLangContainer return the default image for golang
func (m *Go) GetGoLangContainer(goVersion string) *dagger.Container {
	return dag.Container().
		From("golang:" + goVersion)
}

func (m *Go) GetKoContainer() *dagger.Container {
	return dag.Container().
		From("ghcr.io/ko-build/ko:v0.17.1")
}

func New(
	// golang container
	// It need contain golang
	// +optional
	goLangContainer *dagger.Container,
	// +optional
	koContainer *dagger.Container,

	// +defaultPath="/"
	src *dagger.Directory,

) *Go {
	golang := &Go{}

	if goLangContainer != nil {
		golang.GoLangContainer = goLangContainer
	} else {
		golang.GoLangContainer = golang.GetGoLangContainer()
	}

	if koContainer != nil {
		golang.KoContainer = koContainer
	} else {
		golang.KoContainer = golang.GetKoContainer()
	}

	golang.Src = src

	return golang
}

// Lint runs the linter on the provided source code
func (m *Go) Lint(ctx context.Context, src *dagger.Directory) *dagger.Container {
	return dag.GolangciLint().Run(src)
}

// RunPipeline orchestrates running both Lint and Build steps
func (m *Go) RunPipeline(ctx context.Context, src *dagger.Directory, goVersion string) (*dagger.Directory, error) {

	if goVersion == "" {
		goVersion = "1.23.4"
	}

	// STAGE 0: LINT
	fmt.Println("RUNNING LINTING...")
	lintOutput, err := m.Lint(ctx, src).Stdout(ctx)
	if err != nil {
		fmt.Println("ERROR RUNNING LINTER: ", err)
	}
	fmt.Print("LINT RESULT: ", "\n"+lintOutput)

	// STAGE 1: BUILD SOURCE CODE
	fmt.Println("RUNNING BUILD...")
	buildOutput := m.Build(ctx, src)

	// Returning the build output
	return buildOutput, nil
}

// Returns lines that match a pattern in the files of the provided Directory
func (m *Go) Build(ctx context.Context, src *dagger.Directory) *dagger.Directory {

	// MOUNT CLONED REPOSITORY INTO `GOLANG` IMAGE
	golang := m.GoLangContainer.WithDirectory("/src", src).WithWorkdir("/src")

	// DEFINE THE APPLICATION BUILD COMMAND
	path := "build/"
	golang = golang.WithExec([]string{"env", "GOOS=linux", "GOARCH=amd64", "go", "build", "-o", path, "./main.go"})

	// GET REFERENCE TO BUILD OUTPUT DIRECTORY IN CONTAINER
	outputDir := golang.Directory(path)

	return outputDir
}

// Returns lines that match a pattern in the files of the provided Directory
func (m *Go) KoBuild(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	// +default="GITHUB_TOKEN"
	tokenName string,
	token *dagger.Secret,
	// +optional
	// +default="ko.local"
	repo string,
	// +optional
	// +default="."
	buildArg string,
) (buildOutput string) {

	srcDir := "/src"
	// MOUNT CLONED REPOSITORY INTO `KO` IMAGE
	ko := m.KoContainer.WithDirectory(srcDir, src).WithWorkdir(srcDir)

	// DEFINE THE APPLICATION BUILD COMMAND W/ KO
	buildOutput, err := ko.
		WithEnvVariable("KO_DOCKER_REPO", repo).
		WithSecretVariable(tokenName, token).
		WithExec(
			[]string{"ko", "build", buildArg},
		).Stdout(ctx)

	if err != nil {
		fmt.Println("ERROR RUNNING KO", err)
	}

	fmt.Println("BUILD IMAGE: ", buildOutput)

	return buildOutput
}
