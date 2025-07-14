package main

import (
	"context"
	"dagger/helm/internal/dagger"
)

// LINTS A CHART
func (m *Helm) Lint(
	ctx context.Context,
	chart *dagger.Directory) (string, error) {

	updatedChart := m.DependencyUpdate(ctx, chart)

	return dag.
		HelmDisaster37().
		Lint(
			ctx,
			updatedChart,
		)
}
