package main

import (
	"context"
	"dagger/terraform/internal/dagger"
	"fmt"
)

func (m *Terraform) Output(
	ctx context.Context,
	terraformDir *dagger.Directory,
	// +optional
	awsAccessKeyID *dagger.Secret,
	// +optional
	awsSecretAccessKey *dagger.Secret,
	// Kubeconfig secret for Kubernetes backend access
	// +optional
	kubeConfig *dagger.Secret,
	// Path to mount the kubeconfig inside the container (must match backend config_path)
	// +optional
	// +default="/root/.kube/config"
	kubeConfigPath string,
) (string, error) {
	ctr, err := m.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	// Inject AWS creds for S3-compatible backend
	if awsAccessKeyID != nil { // pragma: allowlist secret
		ctr = ctr.WithSecretVariable("AWS_ACCESS_KEY_ID", awsAccessKeyID)
	}
	if awsSecretAccessKey != nil { // pragma: allowlist secret
		ctr = ctr.WithSecretVariable("AWS_SECRET_ACCESS_KEY", awsSecretAccessKey)
	}
	// Prevent attempts to use IMDS, which can cause noisy errors in CI
	ctr = ctr.WithEnvVariable("AWS_EC2_METADATA_DISABLED", "true")

	// MOUNT KUBECONFIG FOR KUBERNETES BACKEND
	if kubeConfig != nil {
		ctr = ctr.WithMountedSecret(kubeConfigPath, kubeConfig)
	}

	workDir := "/src"

	ctr = ctr.
		WithDirectory(workDir, terraformDir, dagger.ContainerWithDirectoryOpts{
			Exclude: []string{}, // Don't exclude anything - we need .terraform if it exists
		}).
		WithWorkdir(workDir).
		WithExec([]string{"terraform", "init", "-reconfigure"}). // Reinitialize backend to restore state connection
		WithExec([]string{"terraform", "output", "--json"})

	out, err := ctr.Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("terraform output failed: %w", err)
	}

	return out, nil
}
