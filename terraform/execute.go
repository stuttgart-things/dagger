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
	// Run terraform output --json after the operation and write result to output.json
	// +optional
	exportTfOutput bool,
	// Resource addresses to limit apply/destroy to, comma-separated, one
	// -target each (e.g. 'vault_kv_secret_v2.this["kv/a"],null_resource.b').
	// Use it when the configuration manages more than the caller owns: an
	// untargeted apply would also plan everything the caller's inputs omit.
	// +optional
	targets string,
	// Apply only if the plan deletes nothing. Plans to a file, aborts if any
	// resource change contains "delete" (a replace does too), then applies
	// exactly that plan -- not a fresh one that could differ. apply only.
	// +optional
	refuseDestroy bool,
	// A service to bind under bindServiceAlias -- from the CLI
	// `tcp://<ip>:<port>`, which the caller's host forwards. Pins a hostname
	// to an address while TLS and SNI still see the real name. For names the
	// engine's resolver cannot answer reliably, e.g. a lab zone without
	// public NS. (/etc/hosts is read-only inside a Dagger exec.)
	// +optional
	bindService *dagger.Service,
	// Hostname bindService is reachable under, e.g. the host in VAULT_ADDR.
	// +optional
	bindServiceAlias string,
) (*dagger.Directory, error) {
	if operation == "" {
		operation = "init"
	}
	if (bindService == nil) != (bindServiceAlias == "") {
		return nil, fmt.Errorf("bindService and bindServiceAlias go together")
	}
	if refuseDestroy && operation != "apply" {
		return nil, fmt.Errorf("refuseDestroy only applies to operation apply, got %q", operation)
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

	// PIN A HOSTNAME TO THE CALLER'S SERVICE
	if bindService != nil {
		ctr = ctr.WithServiceBinding(bindServiceAlias, bindService)
	}

	workDir := "/src"
	ctr = ctr.WithDirectory(workDir, terraformDir).
		WithWorkdir(workDir).
		WithEnvVariable("VAULT_SKIP_VERIFY", "TRUE")

	// ALWAYS RUN INIT FIRST WITH --UPGRADE
	ctr = ctr.WithExec([]string{"terraform", "init", "-upgrade", "-input=false", "-no-color"})

	targetArgs := []string{}
	for _, t := range strings.Split(targets, ",") {
		if t = strings.TrimSpace(t); t != "" {
			targetArgs = append(targetArgs, "-target="+t)
		}
	}

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
		if refuseDestroy {
			ctr = ctr.WithExec(append(append([]string{
				"terraform", "plan", "-out=tfplan", "-input=false", "-no-color"},
				varArgs...), targetArgs...))
			// THE PLAN FILES HOLD VALUES, SO THEY ARE REMOVED ON BOTH PATHS
			ctr = ctr.WithExec([]string{"sh", "-c", `set -eu
terraform show -json tfplan > tfplan.json
deletes=$(jq -r '.resource_changes[]? | select(.change.actions | index("delete")) | "\(.address) \(.change.actions | join("+"))"' tfplan.json)
if [ -n "$deletes" ]; then
  rm -f tfplan tfplan.json
  echo "REFUSING TO APPLY: the plan deletes or replaces:" >&2
  echo "$deletes" >&2
  exit 1
fi`})
			ctr = ctr.WithExec([]string{"terraform", "apply", "-input=false", "-no-color", "tfplan"}).
				WithExec([]string{"rm", "-f", "tfplan", "tfplan.json"})
		} else {
			ctr = ctr.WithExec(append(append([]string{
				"terraform",
				"apply",
				"-auto-approve",
				"-input=false",
				"-no-color"},
				varArgs...,
			), targetArgs...))
		}
	case "destroy":
		ctr = ctr.WithExec(append(append([]string{
			"terraform",
			"destroy",
			"-auto-approve",
			"-input=false",
			"-no-color"},
			varArgs...,
		), targetArgs...))
	default:
		return nil, fmt.Errorf("unsupported terraform operation: %s", operation)
	}

	if exportTfOutput {
		ctr = ctr.WithExec([]string{"sh", "-c", "terraform output --json > output.json"})
	}

	return ctr.Directory(workDir), nil
}
