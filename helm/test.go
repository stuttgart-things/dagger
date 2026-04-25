package main

import (
	"context"
	"dagger/helm/internal/dagger"
)

// Test runs `helm test <release>` against an existing release.
// Requires a kubeconfig pointing at the cluster the release lives in.
func (m *Helm) Test(
	ctx context.Context,
	releaseName string,
	namespace string,
	kubeConfig *dagger.Secret,
	// +optional
	// Container timeout for the test pods (e.g. "5m", "300s")
	timeout string,
	// +optional
	// Show pod logs even on success
	logs bool,
) (string, error) {

	args := []string{"helm", "test", releaseName, "--namespace", namespace}
	if timeout != "" {
		args = append(args, "--timeout", timeout)
	}
	if logs {
		args = append(args, "--logs")
	}

	return m.container().
		WithMountedSecret("/root/.kube/config", kubeConfig).
		WithExec(args).
		Stdout(ctx)
}
