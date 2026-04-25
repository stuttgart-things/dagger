package main

import (
	"context"
	"dagger/helm/internal/dagger"
)

// Diff runs `helmfile diff` against the cluster pointed to by
// kubeConfig, showing what would change for the releases declared
// in the helmfile. Read-only; does not apply changes.
func (m *Helm) Diff(
	ctx context.Context,
	// +optional
	src *dagger.Directory,
	// +optional
	// +default="helmfile.yaml"
	helmfileRef string,
	// Comma-separated key=value pairs for --state-values-set
	// +optional
	stateValues string,
	// +optional
	registrySecret *dagger.Secret,
	kubeConfig *dagger.Secret,
) (string, error) {

	projectDir := "/helmfiles"
	dockerConfigPath := "/root/.docker/config.json"
	kubeConfigPath := "/root/.kube/config"

	helmContainer := m.container().
		WithEnvVariable("HELMFILE_INTERACTIVE", "false")

	if src != nil {
		helmContainer = helmContainer.
			WithDirectory(projectDir, src).
			WithWorkdir(projectDir)
	} else {
		helmContainer = helmContainer.
			WithExec([]string{"mkdir", "-p", projectDir}).
			WithWorkdir(projectDir)
	}

	if registrySecret != nil { // pragma: allowlist secret
		helmContainer = helmContainer.WithMountedSecret(dockerConfigPath, registrySecret)
	}

	helmContainer = helmContainer.WithMountedSecret(kubeConfigPath, kubeConfig)

	args := []string{"helmfile", "-f", helmfileRef, "diff"}
	if stateValues != "" {
		for _, pair := range splitValues(stateValues) {
			args = append(args, "--state-values-set", pair)
		}
	}

	return helmContainer.WithExec(args).Stdout(ctx)
}
