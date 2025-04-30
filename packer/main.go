package main

import (
	"context"
	"dagger/packer/internal/dagger"
	"fmt"
)

type Packer struct {
	// Base Wolfi image to use
	// +optional
	// +default="cgr.dev/chainguard/wolfi-base:latest"
	BaseImage string
}

func (m *Packer) Init(
	ctx context.Context,
	// The Packer version to use
	// +optional
	// +default="1.12.0"
	packerVersion,
	// The Packer arch
	// +optional
	// +default="linux_amd64"
	arch string,
) {
	packer, err := m.container(packerVersion, arch).
		WithExec([]string{"packer", "version"}).Stdout(ctx)

	if err != nil {
		fmt.Errorf("failed to initialize: %w", err)
	}

	fmt.Println("Packer version: ", packer)
}

func (m *Packer) container(
	// The Packer version to use
	// +optional
	// +default="1.12.0"
	packerVersion,
	// The Packer arch
	// +optional
	// +default="linux_amd64"
	arch string) *dagger.Container {

	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	var (
		packerBin   = "packer"
		packerZip   = fmt.Sprintf("%s_%s_%s.zip", packerBin, packerVersion, arch)
		packerURL   = fmt.Sprintf("https://releases.hashicorp.com/%s/%s/%s", packerBin, packerVersion, packerZip)
		destBinPath = "/usr/bin/" + packerBin
	)

	ctr := dag.Container().From(m.BaseImage)

	ctr = ctr.WithExec([]string{"apk", "add", "--no-cache", "wget", "unzip"})
	ctr = ctr.WithExec([]string{"wget", packerURL})
	ctr = ctr.WithExec([]string{"unzip", packerZip})
	ctr = ctr.WithExec([]string{"mv", packerBin, destBinPath})
	ctr = ctr.WithExec([]string{"chmod", "+x", destBinPath})

	return ctr
}
