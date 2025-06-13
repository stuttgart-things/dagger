package main

import (
	"context"
	"dagger/terraform/internal/dagger"
	"fmt"
)

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
