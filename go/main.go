/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"context"
	"dagger/go/internal/dagger"
)

func New(
	// +defaultPath="./"
	src *dagger.Directory,
) *Go {
	return &Go{
		Src: src,
	}
}

type Go struct {
	Src *dagger.Directory
}

// Execute Dev pipeline for sthings-golang application
func (m *Go) DevBuild(ctx context.Context, src *dagger.Directory) *dagger.Directory {

	// LINT THE APPLICATION
	m.Lint(ctx, m.Src)

	// BUILD THE APPLICATION
	outputDir := m.Build(ctx, src)
	return outputDir

}

// Returns lines that match a pattern in the files of the provided Directory
func (m *Go) Build(ctx context.Context, src *dagger.Directory) *dagger.Directory {

	// GET `GOLANG` IMAGE
	golang := dag.Container().From("sthings-golang:latest")

	// MOUNT CLONED REPOSITORY INTO `GOLANG` IMAGE
	golang = golang.WithDirectory("/src", src).WithWorkdir("/src")

	// DEFINE THE APPLICATION BUILD COMMAND
	path := "build/"
	golang = golang.WithExec([]string{"env", "GOOS=linux", "GOARCH=amd64", "go", "build", "-o", path, "./main.go"})

	// GET REFERENCE TO BUILD OUTPUT DIRECTORY IN CONTAINER
	outputDir := golang.Directory(path)

	return outputDir
}

func (m *Go) Lint(ctx context.Context, src *dagger.Directory) *dagger.Container {
	return dag.GolangciLint().Run(src)
}
