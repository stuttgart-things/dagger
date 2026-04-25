package main

import (
	"context"
	"dagger/helm/internal/dagger"
)

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
	// Wait for resources to be ready (--wait)
	// +optional
	wait bool,
	// Timeout for --wait, e.g. "5m", "300s" (--timeout)
	// +optional
	timeout string,
	// Roll back the release if the install/upgrade fails (--atomic)
	// +optional
	atomic bool,
	// Render templates and validate against the cluster but do not
	// apply changes (--dry-run)
	// +optional
	dryRun bool,
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

		// LIFECYCLE FLAGS
		if wait {
			args = append(args, "--wait")
		}
		if timeout != "" {
			args = append(args, "--timeout", timeout)
		}
		if atomic {
			args = append(args, "--atomic")
		}
		if dryRun {
			args = append(args, "--dry-run")
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
