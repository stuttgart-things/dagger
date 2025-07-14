package main

import (
	"context"
	"dagger/helm/internal/dagger"
)

func (m *Helm) Lint(
	ctx context.Context,
	src *dagger.Directory) (string, error) {

	helmContainer := m.container()

	updatedChart := m.DependencyUpdate(
		ctx,
		src,
	)

	lintResult, err := helmContainer.
		WithDirectory("/helm", updatedChart).
		WithWorkdir("/helm").
		WithExec([]string{"helm", "lint", "."}).
		Stdout(ctx)
	if err != nil {
		return "", err
	}

	return lintResult, nil
}
