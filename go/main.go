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

func (m *Go) GetKoContainer(
	// +optional
	// +default="v0.17.1"
	koVersion string,
) *dagger.Container {
	return dag.Container().
		From("ghcr.io/ko-build/ko:" + koVersion)
}

func New(
	// golang container
	// It need contain golang
	// +optional
	goLangContainer *dagger.Container,
	// +optional
	koContainer *dagger.Container,
	// +optional
	// +default="1.23.6"
	goLangVersion string,
	// +defaultPath="/"
	src *dagger.Directory,

) *Go {
	golang := &Go{}

	if goLangContainer != nil {
		golang.GoLangContainer = goLangContainer
	} else {
		golang.GoLangContainer = golang.GetGoLangContainer(goLangVersion)
	}

	if koContainer != nil {
		golang.KoContainer = koContainer
	} else {
		golang.KoContainer = golang.GetKoContainer("v0.17.1")
	}

	golang.Src = src

	return golang
}

// Lint runs the linter on the provided source code
func (m *Go) Lint(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	// +default="500s"
	timeout string,
) *dagger.Container {

	golangciLintRunOpts := dagger.GolangciLintRunOpts{
		Timeout: timeout,
	}

	return dag.GolangciLint().Run(src, golangciLintRunOpts)
}

// Lint runs the linter on the provided source code
func (m *Go) ScanTarBallImage(
	ctx context.Context,
	file *dagger.File,
) (*dagger.File, error) {
	scans := []*dagger.TrivyScan{
		dag.Trivy().ImageTarball(file),
	}

	// Grab the report as a file
	reportFile, err := scans[0].Report("json").Sync(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting report: %w", err)
	}

	return reportFile, nil
}

// RunPipeline orchestrates running both Lint and Build steps
func (m *Go) RunPipeline(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	// +default="1.23.6"
	goVersion string,
	// +optional
	// +default="500s"
	lintTimeout string,
	// +optional
	// +default="linux"
	os string,
	// +optional
	// +default="amd64"
	arch string,
	// +optional
	// +default="main.go"
	goMainFile string,
	// +optional
	// +default="main"
	binName string,
) (*dagger.Directory, error) {

	// STAGE 0: LINT
	fmt.Println("RUNNING LINTING...")
	lintOutput, err := m.Lint(ctx, src, lintTimeout).Stdout(ctx)
	if err != nil {
		fmt.Println("ERROR RUNNING LINTER: ", err)
	}
	fmt.Print("LINT RESULT: ", "\n"+lintOutput)

	// STAGE 1: BUILD SOURCE CODE
	fmt.Println("RUNNING BUILD...")
	buildOutput := m.Build(ctx, goVersion, os, arch, goMainFile, binName, src)

	// Returning the build output
	return buildOutput, nil
}

// Returns lines that match a pattern in the files of the provided Directory
func (m *Go) Build(
	ctx context.Context,
	// +optional
	// +default="1.23.6"
	goVersion string,
	// +optional
	// +default="linux"
	os string,
	// +optional
	// +default="amd64"
	arch string,
	// +optional
	// +default="main.go"
	goMainFile string,
	// +optional
	// +default="main"
	binName string,
	src *dagger.Directory) *dagger.Directory {

	// MOUNT CLONED REPOSITORY INTO `GOLANG` IMAGE

	PrintDirectoryInfo(ctx, src)

	golang := m.
		GetGoLangContainer(goVersion).
		WithDirectory("/src", src).
		WithWorkdir("/src")

	fmt.Println("DIR", src)
	// DEFINE THE APPLICATION BUILD COMMAND
	path := "build/"
	golang = golang.WithExec([]string{
		"env",
		"GOOS=" + os,
		"GOARCH=" + arch,
		"go",
		"build",
		"-o",
		path + "/" + binName,
		goMainFile,
	})

	// GET REFERENCE TO BUILD OUTPUT DIRECTORY IN CONTAINER
	outputDir := golang.Directory(path)

	return outputDir
}

func PrintDirectoryInfo(ctx context.Context, src *dagger.Directory) {
	if src == nil {
		fmt.Println("Directory is nil")
		return
	}

	id, err := src.ID(ctx)
	if err != nil {
		fmt.Println("Error getting directory ID:", err)
		return
	}
	fmt.Println("Dagger Directory ID:", id)

	// List files inside the directory
	entries, err := src.Entries(ctx)
	if err != nil {
		fmt.Println("Error retrieving directory entries:", err)
		return
	}

	fmt.Println("Directory contains:", entries)
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
	// +optional
	// +default="v0.17.1"
	koVersion string,
	// +optional
	// +default="true"
	push string,
) *dagger.Directory {

	srcDir := "/src"

	ko := m.
		GetKoContainer(koVersion).
		WithDirectory(srcDir, src).
		WithWorkdir(srcDir)

	// DEFINE THE APPLICATION BUILD COMMAND W/ KO
	ko = ko.
		WithEnvVariable("KO_DOCKER_REPO", repo).
		WithSecretVariable(tokenName, token).
		WithExec(
			[]string{"ko", "build", "--push=" + push, buildArg},
		)

	outputDir := ko.Directory(srcDir)

	return outputDir
}
