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

// Init Crossplane Package
func (m *Crossplane) InitPackage(ctx context.Context, name string) *dagger.Directory {

	// Define a simple template
	tmpl := `Hello {{.Title}} {{.Name}}!`

	// Data to be used with the template
	data := Data{
		Name:  "John Doe",
		Title: "Mr.",
	}

	rendered := RenderTemplate(tmpl, data)
	fmt.Println(rendered)

	output := m.XplaneContainer.
		WithNewFile("/templates/configmap.yaml", rendered).
		WithExec([]string{"crossplane", "xpkg", "init", name, "configuration-template", "-d", name}).
		WithExec([]string{"ls", "-lta", name}).
		WithExec([]string{"rm", "-rf", name + "/NOTES.txt"}).
		WithExec([]string{"cat", "/templates/configmap.yaml"})

	//fmt.Println(output)
	return output.Directory(name)
}

type Data struct {
	Name  string
	Title string
}

// GetXplaneContainer return the default image for helm
func (m *Crossplane) GetXplaneContainer() *dagger.Container {
	return dag.Container().
		From("ghcr.io/stuttgart-things/crossplane-cli:v1.18.0")
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
