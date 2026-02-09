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
