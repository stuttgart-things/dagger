package main

import (
	"context"
	"dagger/helm/internal/dagger"
)

func (m *Helm) HelmfileOperation(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	// +default="./"
	pathHelmfile string,
	// +optional
	// +default="helmfile.yaml"
	helmfileName string,
	// +optional
	// +default="apply"
	operation string,
	// +optional
	registrySecret *dagger.Secret,
	kubeConfig *dagger.Secret,
) error {

	projectDir := "/helmfiles"
	dockerConfigPath := "/root/.docker/config.json"
	kubeConfigPath := "/root/.kube/config"

	helmContainer := m.container().
		WithDirectory(projectDir, src.Directory(pathHelmfile)).
		WithWorkdir(projectDir)

	// CONDITIONALLY MOUNT THE SECRET IF PROVIDED
	if registrySecret != nil {
		helmContainer = helmContainer.WithMountedSecret(dockerConfigPath, registrySecret)
	}

	// MOUNT KUBECONFIG FILE
	if kubeConfig != nil {
		helmContainer = helmContainer.WithMountedSecret(kubeConfigPath, kubeConfig)
	}

	args := []string{
		"helmfile",
		"-f",
		helmfileName,
		operation}

	if _, err := helmContainer.WithExec(args).Sync(ctx); err != nil {
		return err
	}

	return nil
}
