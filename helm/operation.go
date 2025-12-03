package main

import (
	"context"
	"dagger/helm/internal/dagger"
	"strings"
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
	if registrySecret != nil { // pragma: allowlist secret
		helmContainer = helmContainer.WithMountedSecret(dockerConfigPath, registrySecret)
	}

	// OPTIONAL VAULT ENVS
	if vaultAppRoleID != nil {
		helmContainer = helmContainer.WithSecretVariable("VAULT_ROLE_ID", vaultAppRoleID)
	}

	if vaultSecretID != nil { // pragma: allowlist secret
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

func (m *Helm) Execute(
	ctx context.Context,
	// +optional
	src *dagger.Directory,
	releaseName string,
	// Chart path (not needed for uninstall operation)
	// +optional
	chartPath string,
	// +optional
	valuesFile *dagger.File,
	// +optional
	// +default="install"
	operation string,
	// +optional
	registrySecret *dagger.Secret,
	namespace string,
	// +optional
	kubeConfig *dagger.Secret,
	// Comma-separated values (e.g., "key1=val1,key2=val2")
	// +optional
	values string,
	// Helm repository URL for traditional charts (e.g., "https://helm.cilium.io")
	// Not needed for OCI charts (oci://...) or local charts
	// +optional
	repoURL string,
	// Repository name for traditional charts (e.g., "cilium")
	// +optional
	repoName string,
	// Chart version (e.g., "1.2.3")
	// +optional
	version string,
) error {

	projectDir := "/helm"
	dockerConfigPath := "/root/.docker/config.json"
	kubeConfigPath := "/root/.kube/config"

	helmContainer := m.container()

	if src != nil {
		helmContainer = helmContainer.WithDirectory(projectDir, src).WithWorkdir(projectDir)
	} else {
		helmContainer = helmContainer.
			WithExec([]string{"mkdir", "-p", projectDir}).
			WithWorkdir(projectDir)
	}

	// CONDITIONALLY MOUNT REGISTRY SECRET
	if registrySecret != nil { // pragma: allowlist secret
		helmContainer = helmContainer.WithMountedSecret(dockerConfigPath, registrySecret)
	}

	// MOUNT KUBECONFIG
	if kubeConfig != nil {
		helmContainer = helmContainer.WithMountedSecret(kubeConfigPath, kubeConfig)
	}

	// ADD HELM REPO IF TRADITIONAL REPOSITORY IS PROVIDED
	if repoURL != "" && repoName != "" {
		repoAddArgs := []string{"helm", "repo", "add", repoName, repoURL}
		helmContainer = helmContainer.WithExec(repoAddArgs)
		helmContainer = helmContainer.WithExec([]string{"helm", "repo", "update"})
	}

	// BUILD HELM COMMAND BASED ON OPERATION
	var args []string

	switch operation {
	case "install", "upgrade":
		args = []string{"helm", "upgrade", "--install", releaseName, chartPath, "--namespace", namespace, "--create-namespace"}

		// ADD VERSION IF PROVIDED
		if version != "" {
			args = append(args, "--version", version)
		}

		// ADD VALUES FILE IF PROVIDED
		if valuesFile != nil {
			helmContainer = helmContainer.WithMountedFile(projectDir+"/values.yaml", valuesFile)
			args = append(args, "-f", "values.yaml")
		}

		// ADD SET VALUES IF PROVIDED (COMMA-SEPARATED)
		if values != "" {
			valuePairs := splitValues(values)
			for _, pair := range valuePairs {
				args = append(args, "--set", pair)
			}
		}

	case "uninstall":
		args = []string{"helm", "uninstall", releaseName, "--namespace", namespace}

	default:
		args = []string{"helm", "upgrade", "--install", releaseName, chartPath, "--namespace", namespace, "--create-namespace"}
	}

	if _, err := helmContainer.WithExec(args).Sync(ctx); err != nil {
		return err
	}

	return nil
}

// splitValues splits comma-separated key=value pairs
func splitValues(values string) []string {
	var result []string
	for _, pair := range strings.Split(values, ",") {
		pair = strings.TrimSpace(pair)
		if pair != "" {
			result = append(result, pair)
		}
	}
	return result
}
