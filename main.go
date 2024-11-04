/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"context"
	"golang/internal/dagger"
)

type Golang struct{}

func (m *Golang) Pipeline(ctx context.Context, src *dagger.Directory) *dagger.Directory {

	// BUILD THE APPLICATION
	outputDir := m.Build(ctx, src)

	return outputDir

}

func (m *Golang) Build(ctx context.Context, src *dagger.Directory) *dagger.Directory {

	// GET `GOLANG` IMAGE
	golang := dag.Container().From("golang:latest")

	// MOUNT CLONED REPOSITORY INTO `GOLANG` IMAGE
	golang = golang.WithDirectory("/src", src).WithWorkdir("/src")

	// DEFINE THE APPLICATION BUILD COMMAND
	path := "build/"
	golang = golang.WithExec([]string{"env", "GOOS=linux", "GOARCH=amd64", "go", "build", "-o", path, "./main.go"})

	// GET REFERENCE TO BUILD OUTPUT DIRECTORY IN CONTAINER
	outputDir := golang.Directory(path)

	return outputDir
}
