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
	if vaultRoleID != nil {
		ctr = ctr.WithSecretVariable("TF_VAR_vault_role_id", vaultRoleID)
	}
	if vaultSecretID != nil { // pragma: allowlist secret
		ctr = ctr.WithSecretVariable("TF_VAR_vault_secret_id", vaultSecretID)
	}
	if vaultToken != nil {
		ctr = ctr.WithSecretVariable("TF_VAR_vault_token", vaultToken)
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
