// A generated module for Kyverno functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/kyverno/internal/dagger"
	"fmt"
)

type Kyverno struct {
	// Base Wolfi image to use
	// +optional
	// +default="cgr.dev/chainguard/wolfi-base:latest"
	BaseImage string
}

func (m *Kyverno) Validate(
	ctx context.Context,
	policy *dagger.Directory,
	resource *dagger.Directory,
) error {
	kyverno := m.container().
		WithMountedDirectory("/policy", policy).
		WithMountedDirectory("/resource", resource).
		WithWorkdir("/")

	result, err := kyverno.
		WithExec([]string{"kubectl-kyverno", "apply", "/policy", "--resource", "/resource"}).
		Stdout(ctx)

	if err != nil {
		return fmt.Errorf("failed to validate: %w", err)
	}

	fmt.Println(result)
	return nil
}

func (m *Kyverno) Version(
	ctx context.Context) (version string) {
	kyverno := m.container()

	cmd := []string{"kubectl-kyverno", "version"}

	version, err := kyverno.WithExec(cmd).Stdout(ctx)
	if err != nil {
		fmt.Println("Error running kyverno version: ", err)
		return
	}
	fmt.Println("Kyverno version: ", version)

	return version
}

func (m *Kyverno) container() *dagger.Container {
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	ctr := dag.Container().From(m.BaseImage)

	pkg := "kyverno-cli"
	ctr = ctr.WithExec([]string{"apk", "add", "--no-cache", pkg})
	ctr = ctr.WithEntrypoint([]string{"kubectl-kyverno"})

	return ctr
}
