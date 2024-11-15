/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"bytes"
	"context"
	"dagger/crossplane/internal/dagger"
	"fmt"
	"html/template"
	"log"
)

type Crossplane struct {
	XplaneContainer *dagger.Container
}

type Data struct {
	Name  string
	Title string
}

// Init Crossplane Package
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
func (m *Crossplane) InitCustomPackage(ctx context.Context, name string) *dagger.Directory {

	// DEFINE INTERFACE MAP FOR TEMPLATE DATA INLINE - LATER LOAD AS YAML FILE
	// DEFINE A STRUCT WITH THE NEEDED PACKAGE FOLDER STRUCTURE AND TARGET PATHS
	// RENDER THE TEMPLATES WITH THE DATA
	// COPY TO CONTAINER AND RETURN OR TRY TO RETURN FOR EXPORTING WITHOUT USING A CONTAINER

	// Define a simple template
	tmpl := `Hello {{.Title}} {{.Name}}!`

	// Data to be used with the template
	data := Data{
		Name:  "John Doe",
		Title: "Mr.",
	}

	rendered := RenderTemplate(tmpl, data)

	output := m.XplaneContainer.
		WithNewFile("/templates/configmap.yaml", rendered).
		WithExec([]string{"cat", "/templates/configmap.yaml"})

	return output.Directory(name)
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

func RenderTemplate(templateData string, data Data) (renderTemplate string) {

	// PARSE THE TEMPLATE
	t, err := template.New("dagger-xplane").Parse(templateData)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	// EXECUTE THE TEMPLATE AND WRITE THE OUTPUT TO A BUFFER
	var output bytes.Buffer
	err = t.Execute(&output, data)
	if err != nil {
		log.Fatalf("Error executing template: %v", err)
	}

	renderTemplate = output.String()

	return renderTemplate
}
