/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"context"
	"dagger/crossplane/internal/dagger"
)

type Crossplane struct {
	XplaneContainer *dagger.Container
}

// Constructor creates and returns a new Crossplane module instance
func (m *Crossplane) Constructor(
	ctx context.Context,
	// xplane container
	// It need contain xplane
	// +optional
	xplaneContainer *dagger.Container,
) *Crossplane {
	if xplaneContainer != nil {
		m.XplaneContainer = xplaneContainer
	} else {
		m.XplaneContainer = m.GetXplaneContainer(ctx)
	}
	return m
}
