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
}

// GetGoLangContainer return the default image for golang
func (m *Go) GetGoLangContainer() *dagger.Container {
	return dag.Container().
		From("golang:1.23.3")
}

func New(
	// helm container
	// It need contain helm
	// +optional
	goLangContainer *dagger.Container,

	// +defaultPath="/"
	src *dagger.Directory,

) *Go {
	golang := &Go{}

	if goLangContainer != nil {
		golang.GoLangContainer = goLangContainer
	} else {
		golang.GoLangContainer = golang.GetGoLangContainer()
	}

	golang.Src = src

	return golang
}

// Lint runs the linter on the provided source code
func (m *Go) Lint(ctx context.Context, src *dagger.Directory) *dagger.Container {
	return dag.GolangciLint().Run(src)
}

// RunPipeline orchestrates running both Lint and Build steps
func (m *Go) RunPipeline(ctx context.Context, src *dagger.Directory) (*dagger.Directory, error) {

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
