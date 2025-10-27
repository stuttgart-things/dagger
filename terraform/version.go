package main

import (
	"context"
	"fmt"

)

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
