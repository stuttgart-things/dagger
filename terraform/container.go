package main

import (
	"context"
	"fmt"

	"dagger/terraform/internal/dagger"
)

func (m *Terraform) container(ctx context.Context) (*dagger.Container, error) {
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	const terraformVersion = "1.12.1"
	terraformURL := fmt.Sprintf("https://releases.hashicorp.com/terraform/%s/terraform_%s_linux_amd64.zip", terraformVersion, terraformVersion)

	ctr := dag.Container().
		From(m.BaseImage).
		WithExec([]string{"apk", "add", "--no-cache", "curl", "unzip", "git"}).
		WithExec([]string{"sh", "-c", fmt.Sprintf("curl -sSL %s -o /tmp/terraform.zip", terraformURL)}).
		WithExec([]string{"unzip", "-d", "/usr/bin", "/tmp/terraform.zip"}).
		WithExec([]string{"chmod", "+x", "/usr/bin/terraform"}).
		WithEntrypoint([]string{"terraform"})

	return ctr, nil
}
