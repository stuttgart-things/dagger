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
	// +optional
	kubeConfig *dagger.Secret,
	// +optional
	vaultAppRoleID *dagger.Secret,
	// +optional
	vaultSecretID *dagger.Secret,
	// +optional
	vaultUrl *dagger.Secret,
	// +optional
	secretPathKubeconfig string,
	// +optional
	// +default="approle"
	vaultAuthMethod string,
) error {

	projectDir := "/helmfiles"
	dockerConfigPath := "/root/.docker/config.json"
	kubeConfigPath := "/root/.kube/config"

	helmContainer := m.container().
		WithDirectory(projectDir, src.Directory(pathHelmfile)).
		WithWorkdir(projectDir).
		WithEnvVariable("VAULT_SKIP_VERIFY", "TRUE").
		WithEnvVariable("VAULT_AUTH_METHOD", "approle")

	// CONDITIONALLY MOUNT THE SECRET IF PROVIDED
	if registrySecret != nil {
		helmContainer = helmContainer.WithMountedSecret(dockerConfigPath, registrySecret)
	}

	// OPTIONAL VAULT ENVS
	if vaultAppRoleID != nil {
		helmContainer = helmContainer.WithSecretVariable("VAULT_ROLE_ID", vaultAppRoleID)
	}

	if vaultSecretID != nil {
		helmContainer = helmContainer.WithSecretVariable("VAULT_SECRET_ID", vaultSecretID)
	}

	if vaultUrl != nil {
		helmContainer = helmContainer.WithSecretVariable("VAULT_ADDR", vaultUrl)
	}

	// DOWNLOAD SECRET FILES FROM VAULT WITH VALS
	if secretPathKubeconfig != "" {
		// Create .kube directory first
		helmContainer = helmContainer.WithExec([]string{"mkdir", "-p", "/root/.kube"})

		// Use proper shell execution for piped commands
		valsCommand := "vals get ref+vault://" + secretPathKubeconfig + " | base64 -d > " + kubeConfigPath

		helmContainer = helmContainer.WithExec([]string{
			"sh",
			"-c",
			valsCommand,
		})

		// Set proper permissions on kubeconfig
		helmContainer = helmContainer.WithExec([]string{"chmod", "600", kubeConfigPath})
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
