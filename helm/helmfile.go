package main

import (
	"context"
	"dagger/helm/internal/dagger"
	"strings"
)

func (m *Helm) HelmfileOperation(
	ctx context.Context,
	// +optional
	src *dagger.Directory,
	// +optional
	// +default="helmfile.yaml"
	helmfileRef string,
	// +optional
	// +default="apply"
	operation string,
	// Comma-separated key=value pairs for --state-values-set
	// (e.g., "issuerName=cluster-issuer-approle,domain=demo.example.com")
	// +optional
	stateValues string,
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
		WithEnvVariable("VAULT_SKIP_VERIFY", "TRUE").
		WithEnvVariable("VAULT_AUTH_METHOD", vaultAuthMethod).
		WithEnvVariable("HELMFILE_INTERACTIVE", "false")

	// CONDITIONALLY MOUNT SOURCE DIRECTORY IF PROVIDED
	if src != nil {
		helmContainer = helmContainer.
			WithDirectory(projectDir, src).
			WithWorkdir(projectDir)
	} else {
		helmContainer = helmContainer.
			WithExec([]string{"mkdir", "-p", projectDir}).
			WithWorkdir(projectDir)
	}

	// CONDITIONALLY MOUNT THE SECRET IF PROVIDED
	if registrySecret != nil { // pragma: allowlist secret
		helmContainer = helmContainer.WithMountedSecret(dockerConfigPath, registrySecret)
	}

	helmContainer = dag.Vault().WithAppRoleEnv(helmContainer, dagger.VaultWithAppRoleEnvOpts{
		RoleID:   vaultAppRoleID,
		SecretID: vaultSecretID, // pragma: allowlist secret
		Addr:     vaultUrl,
	})

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

	// BUILD HELMFILE COMMAND
	args := []string{"helmfile", "-f", helmfileRef, operation}

	// ADD STATE VALUES IF PROVIDED
	if stateValues != "" {
		statePairs := splitValues(stateValues)
		for _, pair := range statePairs {
			args = append(args, "--state-values-set", pair)
		}
	}

	// FOR DESTROY, USE YES PIPING WITH DEBUG
	if operation == "destroy" {
		cmdString := "yes | " + strings.Join(args, " ") + " --debug"
		if _, err := helmContainer.WithExec([]string{"sh", "-c", cmdString}).Sync(ctx); err != nil {
			return err
		}
	} else {
		if _, err := helmContainer.WithExec(args).Sync(ctx); err != nil {
			return err
		}
	}

	return nil
}
