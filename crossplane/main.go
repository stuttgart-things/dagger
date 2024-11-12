/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"context"
	"dagger/crossplane/internal/dagger"
	"fmt"
)

type Crossplane struct {
	XplaneContainer *dagger.Container
}

// Init Crossplane Package
func (m *Crossplane) InitPackage(ctx context.Context, name string) {

	output, err := m.XplaneContainer.
		WithWorkdir("/project").
		WithExec(
			[]string{"crossplane", "xpkg", "init", name, "configuration-template"}).
		Stdout(ctx)

	fmt.Println(err)

	fmt.Println(output)
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
