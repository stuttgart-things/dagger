// Terraform is a Dagger module that runs Terraform commands inside a containerized environment.
//
// It uses a custom container built on the Wolfi base image with Terraform, SOPS, Git, and supporting tools pre-installed.
// The module supports securely decrypting a SOPS-encrypted `terraform.tfvars.json` using an optional AGE key,
// and mounts it into the working directory for use during plan or apply operations.
//
// Supported Terraform operations include `init`, `apply`, and `destroy`. The module always runs `terraform init` first,
// and then executes the specified operation. After execution, any sensitive files such as the decrypted tfvars file are deleted.
//
// The module also exposes helper methods:
//   - `Version`: returns the installed Terraform version.
//   - `Output`: retrieves Terraform outputs in JSON format.
//   - `DecryptSops`: decrypts a file using SOPS and returns its content as a string.
//
// It is designed to run Terraform commands reproducibly in CI pipelines or local development environments with secret handling and plugin caching.

package main

import (
	"context"
	"fmt"

	"dagger/terraform/internal/dagger"
)

type Terraform struct {
	BaseImage string
}

func (m *Terraform) Version(ctx context.Context) (string, error) {
	ctr, err := m.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	out, err := ctr.WithExec([]string{"terraform", "version"}).Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("terraform version failed: %w", err)
	}

	return out, nil
}

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

func (m *Terraform) container(ctx context.Context) (*dagger.Container, error) {
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	const terraformVersion = "1.12.1"
	terraformURL := fmt.Sprintf("https://releases.hashicorp.com/terraform/%s/terraform_%s_linux_amd64.zip", terraformVersion, terraformVersion)

	ctr := dag.Container().
		From(m.BaseImage).
		WithExec([]string{"apk", "add", "--no-cache", "curl", "unzip", "sops", "git"}).
		WithExec([]string{"sh", "-c", fmt.Sprintf("curl -sSL %s -o /tmp/terraform.zip", terraformURL)}).
		WithExec([]string{"unzip", "-d", "/usr/bin", "/tmp/terraform.zip"}).
		WithExec([]string{"chmod", "+x", "/usr/bin/terraform"}).
		WithEntrypoint([]string{"terraform"})

	return ctr, nil
}

func (m *Terraform) DecryptSops(
    ctx context.Context,
    sopsKey *dagger.Secret,
    encryptedFile *dagger.File,
) (string, error) {
    ctr, err := m.container(ctx)
    if err != nil {
        return "", fmt.Errorf("container init failed: %w", err)
    }

    workDir := "/src"
    fileName := "encrypted.json"

    // Mount encrypted file into container using string concatenation
    ctr = ctr.
        WithMountedFile(workDir + "/" + fileName, encryptedFile).
        WithWorkdir(workDir)

    // Add SOPS key secret if provided
    if sopsKey != nil {
        ctr = ctr.WithSecretVariable("SOPS_AGE_KEY", sopsKey)
    }

    // Decrypt file
    out, err := ctr.
        WithEntrypoint([]string{}). // Clear terraform entrypoint
        WithExec([]string{"sops", "-d", fileName}).
        Stdout(ctx)
    if err != nil {
        return "", fmt.Errorf("sops decryption failed: %w", err)
    }

    return out, nil
}


func (m *Terraform) Output(
	ctx context.Context,
	terraformDir *dagger.Directory,
) (string, error) {
	ctr, err := m.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	workDir := "/src"

	ctr = ctr.
		WithDirectory(workDir, terraformDir).
		WithWorkdir(workDir).
		WithExec([]string{"terraform", "output", "-json"})

	out, err := ctr.Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("terraform output failed: %w", err)
	}

	return out, nil
}