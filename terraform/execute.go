package main

import (
	"context"
	"fmt"
	"strings"

	"dagger/terraform/internal/dagger"
)

func (m *Terraform) Execute(
	ctx context.Context,
	terraformDir *dagger.Directory,
	// +optional
	// +default="apply"
	operation string,
	// +optional
	// e.g., "name=patrick,food=schnitzel"
	variables string,
	// AWS S3/MinIO credentials
	// +optional
	awsAccessKeyID *dagger.Secret,
	// +optional
	awsSecretAccessKey *dagger.Secret,
	// +optional
	secretJsonVariables *dagger.Secret,
	// vaultRoleID
	// +optional
	vaultRoleID *dagger.Secret,
	// vaultSecretID
	// +optional
	vaultSecretID *dagger.Secret,
	// vaultToken
	// +optional
	vaultToken *dagger.Secret,
	// Vault address (e.g. "https://vault.example.com")
	// +optional
	vaultAddr string,
	// Kubeconfig secret for Kubernetes backend access
	// +optional
	kubeConfig *dagger.Secret,
	// Path to mount the kubeconfig inside the container (must match backend config_path)
	// +optional
	// +default="/root/.kube/config"
	kubeConfigPath string,
) (*dagger.Directory, error) {
	if operation == "" {
		operation = "init"
	}

	// GET THE BASE CONTAINER WITH TERRAFORM
	ctr, err := m.container(ctx)
	if err != nil {
		return nil, fmt.Errorf("container init failed: %w", err)
	}

	// INJECT VAULT SECRETS AS ENVIRONMENT VARIABLES
	// INJECT AWS SECRETS AS ENVIRONMENT VARIABLES FOR S3 BACKEND
	if awsAccessKeyID != nil { // pragma: allowlist secret
		ctr = ctr.WithSecretVariable("AWS_ACCESS_KEY_ID", awsAccessKeyID)
	}
	if awsSecretAccessKey != nil { // pragma: allowlist secret
		ctr = ctr.WithSecretVariable("AWS_SECRET_ACCESS_KEY", awsSecretAccessKey)
	}

	// INJECT VAULT SECRETS AS ENVIRONMENT VARIABLES
	if vaultRoleID != nil {
		ctr = ctr.WithSecretVariable("VAULT_ROLE_ID", vaultRoleID)
	}
	if vaultSecretID != nil { // pragma: allowlist secret
		ctr = ctr.WithSecretVariable("VAULT_SECRET_ID", vaultSecretID)
	}
	if vaultToken != nil {
		ctr = ctr.WithSecretVariable("VAULT_TOKEN", vaultToken)
	}
	if vaultAddr != "" {
		ctr = ctr.WithEnvVariable("VAULT_ADDR", vaultAddr)
	}

	// MOUNT KUBECONFIG FOR KUBERNETES BACKEND
	if kubeConfig != nil {
		ctr = ctr.WithMountedSecret(kubeConfigPath, kubeConfig)
	}

	workDir := "/src"
	ctr = ctr.WithDirectory(workDir, terraformDir).
		WithWorkdir(workDir).
		WithEnvVariable("VAULT_SKIP_VERIFY", "TRUE")

	// ALWAYS RUN INIT FIRST WITH --UPGRADE
	ctr = ctr.WithExec([]string{"terraform", "init", "-upgrade", "-input=false", "-no-color"})

	// PARSE VARIABLES STRING INTO -VAR ARGUMENTS
	varArgs := []string{}
	if variables != "" {
		pairs := strings.Split(variables, ",")
		for _, pair := range pairs {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}
			varArgs = append(varArgs, "-var", pair)
		}
	}

	if secretJsonVariables != nil { // pragma: allowlist secret
		// MOUNT THE SECRET VARIABLES AS A FILE
		ctr = ctr.WithMountedSecret(workDir+"/terraform.tfvars.json", secretJsonVariables)
	}

	switch operation {
	case "init":
		// Nothing more to do
	case "apply":
		ctr = ctr.WithExec(append([]string{
			"terraform",
			"apply",
			"-auto-approve",
			"-input=false",
			"-no-color"},
			varArgs...,
		))
	case "destroy":
		ctr = ctr.WithExec(append([]string{
			"terraform",
			"destroy",
			"-auto-approve",
			"-input=false",
			"-no-color"},
			varArgs...,
		))
	default:
		return nil, fmt.Errorf("unsupported terraform operation: %s", operation)
	}

	return ctr.Directory(workDir), nil
}
