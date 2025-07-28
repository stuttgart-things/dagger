/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"dagger/go/internal/dagger"
)

type Go struct {
	Src             *dagger.Directory
	GoLangContainer *dagger.Container
	KoContainer     *dagger.Container
}

type GoBuildOpts struct {
	GoVersion   string
	Os          string
	Arch        string
	GoMainFile  string
	BinName     string
	Ldflags     string // Add this field for linker flags
	PackageName string
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
