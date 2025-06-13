package main

import (
	"context"
	"fmt"

	"dagger/terraform/internal/dagger"
)


func (m *Terraform) Execute(
	ctx context.Context,
	terraformDir *dagger.Directory,
	// +optional
	// +default="apply"
	operation string,
	// +optional
	encryptedFile *dagger.File,
	// +optional
	sopsKey *dagger.Secret,
) (*dagger.Directory, error) {
	if operation == "" {
		operation = "init"
	}

	// Get the base container with Terraform
	ctr, err := m.container(ctx)
	if err != nil {
		return nil, fmt.Errorf("container init failed: %w", err)
	}

	workDir := "/src"
	ctr = ctr.WithDirectory(workDir, terraformDir).WithWorkdir(workDir)

	// If encrypted tfvars is provided, decrypt and mount it
	if encryptedFile != nil {
		// Decrypt to string
		decryptedContent, err := m.DecryptSops(ctx, sopsKey, encryptedFile)
		if err != nil {
			return nil, fmt.Errorf("decrypting sops file failed: %w", err)
		}

		// Create a file with the decrypted content and mount it as terraform.tfvars.json
		ctr = ctr.WithNewFile(fmt.Sprintf("%s/terraform.tfvars.json", workDir), decryptedContent)
	}

	// Always run init first with --upgrade
	ctr = ctr.WithExec([]string{"terraform", "init", "-upgrade", "-input=false", "-no-color"})

	switch operation {
	case "init":
		// Nothing more to do
	case "apply":
		ctr = ctr.WithExec([]string{"terraform", "apply", "-auto-approve", "-input=false", "-no-color"})
	case "destroy":
		ctr = ctr.WithExec([]string{"terraform", "destroy", "-auto-approve", "-input=false", "-no-color"})
	default:
		return nil, fmt.Errorf("unsupported terraform operation: %s", operation)
	}

		// Delete the tfvars file
	if encryptedFile != nil {
		ctr = ctr.WithExec([]string{"rm", "-f", "terraform.tfvars.json"})
	}

	return ctr.Directory(workDir), nil
}