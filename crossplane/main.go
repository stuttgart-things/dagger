/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"context"
	"dagger/crossplane/internal/dagger"
	"dagger/crossplane/templates"

	"fmt"
)

type Crossplane struct {
	XplaneContainer *dagger.Container
}

// Package Crossplane Package
func (m *Crossplane) Package(ctx context.Context, src *dagger.Directory) *dagger.Directory {

	xplane := m.XplaneContainer.
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"crossplane", "xpkg", "build"})

	buildArtifact, err := xplane.WithExec(
		[]string{"find", "-maxdepth", "1", "-name", "*.xpkg", "-exec", "basename", "{}", ";"}).
		Stdout(ctx)

	if err != nil {
		fmt.Println("ERROR GETTING BUILD ARTIFACT: ", err)
	}

	fmt.Println("BUILD PACKAGE: ", buildArtifact)

	return xplane.Directory("/src")
}

// GetXplaneContainer return the default image for helm
func (m *Crossplane) GetXplaneContainer() *dagger.Container {
	return dag.Container().
		From("ghcr.io/stuttgart-things/crossplane-cli:v1.18.0")
}

// Init Crossplane Package based on custom templates and a configuration file
func (m *Crossplane) InitCustomPackage(ctx context.Context) *dagger.Directory {

	// DEFINE INTERFACE MAP FOR TEMPLATE DATA INLINE - LATER LOAD AS YAML FILE
	// DEFINE A STRUCT WITH THE NEEDED PACKAGE FOLDER STRUCTURE AND TARGET PATHS
	// RENDER THE TEMPLATES WITH THE DATA
	// COPY TO CONTAINER AND RETURN OR TRY TO RETURN FOR EXPORTING WITHOUT USING A CONTAINER

	xplane := m.XplaneContainer

	packageName := "test"
	workingDir := "/" + packageName + "/"

	// Data to be used with the template
	data := map[string]interface{}{
		"kind":      "test",
		"namespace": "crossplane-system",
		"claimName": "incluster",
	}

	for _, template := range templates.PackageFiles {
		rendered := templates.RenderTemplate(template.Template, data)
		xplane = xplane.WithNewFile(workingDir+template.Destination, rendered)
	}

	return xplane.Directory(workingDir)
}

// Init Crossplane Package
func (m *Crossplane) InitPackage(ctx context.Context, name string) *dagger.Directory {

	output := m.XplaneContainer.
		WithExec([]string{"crossplane", "xpkg", "init", name, "configuration-template", "-d", name}).
		WithExec([]string{"ls", "-lta", name}).
		WithExec([]string{"rm", "-rf", name + "/NOTES.txt"})

	return output.Directory(name)
}

func New(
	// xplane container
	// It need contain xplane
	// +optional
	xplaneContainer *dagger.Container,

) *Crossplane {
	xplane := &Crossplane{}

	if xplaneContainer != nil {
		xplane.XplaneContainer = xplaneContainer
	} else {
		xplane.XplaneContainer = xplane.GetXplaneContainer()
	}
	return xplane
}
