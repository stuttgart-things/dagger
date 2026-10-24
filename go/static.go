package main

import (
	"context"
	"dagger/go/internal/dagger"
)

// Lint runs the linter on the provided source code
func (m *Go) Lint(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	// +default="500s"
	timeout string,
	// golangci-lint image tag; must be built with a Go >= the go directive in go.mod
	// +optional
	// +default="v2.15.0"
	lintVersion string,
	// Go image tag the linter runs in
	// +optional
	// +default="1.27.3"
	goVersion string,
) *dagger.Container {

	golangciLintRunOpts := dagger.GolangciLintRunOpts{
		Timeout: timeout,
	}

	return dag.GolangciLint(dagger.GolangciLintOpts{
		Version:   lintVersion,
		GoVersion: goVersion,
	}).Run(
		src,
		golangciLintRunOpts,
	)
}
