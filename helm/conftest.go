package main

import (
	"context"
	"dagger/helm/internal/dagger"
)

// Conftest renders a chart and evaluates the caller-supplied Rego policy
// directory with `conftest test`. The policy set is deferred; this function
// exists now so the signature is frozen and consumers can wire calls before
// policies land.
func (m *Helm) Conftest(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	valuesFile *dagger.File,
	// +optional
	policyDir *dagger.Directory,
	// +optional
	registrySecret *dagger.Secret,
) (string, error) {

	renderedManifests, err := m.Render(
		ctx,
		src,
		valuesFile,
		registrySecret,
	)
	if err != nil {
		return "", err
	}

	ctr := m.container().
		WithWorkdir("/manifests").
		WithNewFile("rendered.yaml", renderedManifests)

	args := []string{"conftest", "test", "--output", "json"}
	if policyDir != nil {
		ctr = ctr.WithDirectory("/policy", policyDir)
		args = append(args, "--policy", "/policy")
	}
	args = append(args, "rendered.yaml")

	return ctr.WithExec(args).Stdout(ctx)
}
