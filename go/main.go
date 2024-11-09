/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"context"
	"dagger/go/internal/dagger"
	"fmt"
)

type Go struct {
	Src *dagger.Directory
}

func New(
	// +defaultPath="/"
	src *dagger.Directory,
) *Go {
	return &Go{
		Src: src,
	}
}

// Lint
func (m *Go) Lint(ctx context.Context, src *dagger.Directory) *dagger.Container {
	return dag.GolangciLint().Run(src)
}

// RunPipeline method: Orchestrates running both Lint and Build steps
func (m *Go) RunPipeline(ctx context.Context, src *dagger.Directory) (*dagger.Directory, error) {
	// Create a container for the Go build environment
	container := dag.Container().From("golang:latest")

	// Step 1: Lint the source code
	fmt.Println("Running linting...")
	//dag.GolangciLint().Run(src)
	// run linter
	lintOutput, err := m.Lint(ctx, src).Stdout(ctx)
	if err != nil {
		fmt.Println("Error running linter: ", err)
	}

	output := "\n" + lintOutput
	fmt.Println("LINT!!!!", output)

	// You can check the lint result or logs here if necessary

	// Step 2: Build the source code
	fmt.Println("Running build...")
	buildOutput := m.Build(ctx, src, container)

	// Returning the build output
	return buildOutput, nil
}

// Returns lines that match a pattern in the files of the provided Directory
func (m *Go) Build(ctx context.Context, src *dagger.Directory, container *dagger.Container) *dagger.Directory {

	// GET `GOLANG` IMAGE
	// golang := dag.Container().From("golang:latest")

	// MOUNT CLONED REPOSITORY INTO `GOLANG` IMAGE
	golang := container.WithDirectory("/src", src).WithWorkdir("/src")

	// DEFINE THE APPLICATION BUILD COMMAND
	path := "build/"
	golang = golang.WithExec([]string{"env", "GOOS=linux", "GOARCH=amd64", "go", "build", "-o", path, "./main.go"})

	// GET REFERENCE TO BUILD OUTPUT DIRECTORY IN CONTAINER
	outputDir := golang.Directory(path)

	return outputDir
}

// func (m *Go) Lint(ctx context.Context, src *dagger.Directory) {

// 	lint := dag.GolangciLint().Run(src)

// 	fmt.Println(lint)
// }
