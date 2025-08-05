package main

import (
	"context"
	"dagger/kyverno/internal/dagger"
	"fmt"
)

type Kyverno struct {
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
		WithExec([]string{
			"kubectl-kyverno",
			"apply",
			"/policy",
			"--resource",
			"/resource"}).
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
