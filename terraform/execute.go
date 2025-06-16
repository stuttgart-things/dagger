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

	return ctr.Directory(workDir), nil
}
