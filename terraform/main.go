package main

import (
	"context"
	"fmt"

	"dagger.io/dagger"
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

func (m *Terraform) container(ctx context.Context) (*dagger.Container, error) {
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(nil))
	if err != nil {
		return nil, err
	}
	// DO NOT defer client.Close() here—let the caller handle it

	const terraformVersion = "1.12.1"
	terraformURL := fmt.Sprintf("https://releases.hashicorp.com/terraform/%s/terraform_%s_linux_amd64.zip", terraformVersion, terraformVersion)

	ctr := client.Container().
		From(m.BaseImage).
		WithExec([]string{"apk", "add", "--no-cache", "curl", "unzip", "sops"}).
		WithExec([]string{"sh", "-c", fmt.Sprintf("curl -sSL %s -o /tmp/terraform.zip", terraformURL)}).
		WithExec([]string{"unzip", "-d", "/usr/bin", "/tmp/terraform.zip"}).
		WithExec([]string{"chmod", "+x", "/usr/bin/terraform"}).
		WithEntrypoint([]string{"terraform"})

	return ctr, nil
}
