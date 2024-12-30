/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"context"
	"dagger/crossplane/internal/dagger"
	reg "dagger/crossplane/registry"
	"dagger/crossplane/templates"
	"fmt"
	"strings"

	"dagger.io/dagger/dag"
	strcase "github.com/stoewer/go-strcase"
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

// Push Crossplane Package
func (m *Crossplane) Push(
	ctx context.Context,
	src *dagger.Directory,
	// +default="ghcr.io"
	registry string,
	username string,
	password *dagger.Secret,
	destination string) string {

	dirWithPackage := m.Package(ctx, src)

	passwordPlaintext, err := password.Plaintext(ctx)

	configJSON, err := reg.CreateDockerConfigJSON(username, passwordPlaintext, registry)
	if err != nil {
		fmt.Printf("ERROR CREATING DOCKER config.json: %v\n", err)
	}

	status, err := m.XplaneContainer.
		WithNewFile("/root/.docker/config.json", configJSON).
		WithDirectory("/src", dirWithPackage).
		WithWorkdir("/src").
		WithExec([]string{"crossplane", "xpkg", "push", destination}).
		Stdout(ctx)

	if err != nil {
		fmt.Println("ERROR PUSHING PACKAGE: ", err)
	}

	fmt.Println("PACKAGE STATUS: ", status)

	return status
}

// GetXplaneContainer return the default image for helm
func (m *Crossplane) GetXplaneContainer() *dagger.Container {
	return dag.Container().
		From("ghcr.io/stuttgart-things/crossplane-cli:v1.18.0")
}

// Init Crossplane Package based on custom templates and a configuration file
func (m *Crossplane) InitCustomPackage(ctx context.Context, name string) *dagger.Directory {

	// DEFINE INTERFACE MAP FOR TEMPLATE DATA INLINE - LATER LOAD AS YAML FILE
	// DEFINE A STRUCT WITH THE NEEDED PACKAGE FOLDER STRUCTURE AND TARGET PATHS
	// RENDER THE TEMPLATES WITH THE DATA
	// COPY TO CONTAINER AND RETURN OR TRY TO RETURN FOR EXPORTING WITHOUT USING A CONTAINER

	functions := []templates.FunctionPackage{
		{
			Name:       "function-go-templating",
			PackageURL: "xpkg.upbound.io/crossplane-contrib/function-go-templating",
			Version:    "v0.7.0",
			ApiVersion: "pkg.crossplane.io/v1beta1",
		},
		{
			Name:       "function-patch-and-transform",
			PackageURL: "xpkg.upbound.io/crossplane-contrib/function-patch-and-transform",
			Version:    "v0.1.4",
			ApiVersion: "pkg.crossplane.io/v1beta1",
		},
	}

	xplane := m.XplaneContainer

	packageName := strings.ToLower(name)
	workingDir := "/" + packageName + "/"

	// Data to be used with the template
	data := map[string]interface{}{
		"name":                  packageName,
		"namespace":             "crossplane-system",
		"compositionApiVersion": "apiextensions.crossplane.io/v1",
		"claimName":             "incluster",
		"apiGroup":              "resources.stuttgart-things.com",
		"claimApiVersion":       "v1alpha1",
		"maintainer":            "patrick.hermann@sva.de",
		"source":                "github.com/stuttgart-things/stuttgart-things",
		"license":               "Apache-2.0",
		"crossplaneVersion":     ">=v1.14.1-0",
		"kindLower":             strings.ToLower(strcase.LowerCamelCase(name)),
		"kindLowerX":            "x" + packageName,
		"kind":                  "X" + ToTitle(strcase.LowerCamelCase(name)),
		"plural":                strings.ToLower("x" + strcase.LowerCamelCase(name) + "s"),
		"claimKind":             ToTitle(strcase.LowerCamelCase(name)),
		"claimPlural":           strings.ToLower(strcase.LowerCamelCase(name)) + "s",
		"compositeApiVersion":   "apiextensions.crossplane.io/v1",
		"functions":             functions,
	}

	fmt.Println("KINDS: ", data)

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

func ToTitle(str string) string {
	letters := strings.Split(str, "")
	return strings.ToUpper(letters[0]) + strings.Join(letters[1:], "")
}
