package main

// dagger call -m packer build --repo-url https://github.com/stuttgart-things/stuttgart-things.git --branch main --token env:GITHUB_TOKEN  --progress plain

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

func (m *Packer) Build(
	ctx context.Context,
	// The Packer version to use
	// +optional
	// +default="1.12.0"
	packerVersion,
	// The Packer arch
	// +optional
	// +default="linux_amd64"
	arch,
	repoURL,
	// The Branch name
	// +optional
	// +default="main"
	branch string,
	// If true, only init packer w/out build
	// +optional
	// +default=false
	initOnly bool,
	token *dagger.Secret, // injected securely
) {

	repoContent, err := m.ClonePrivateRepo(ctx, repoURL, branch, token)
	if err != nil {
		fmt.Errorf("failed to clone repo: %w", err)
	}

	entries, err := repoContent.Entries(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("Top-level entries:", entries)

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

// ClonePrivateRepo clones a private GitHub repo using HTTPS and a personal access token
func (m *Packer) ClonePrivateRepo(
	ctx context.Context,
	repoURL, // e.g. "https://github.com/your-org/your-private-repo.git"
	branch string, // e.g. "main"
	token *dagger.Secret, // injected securely
) (*dagger.Directory, error) {
	src := dag.Git(repoURL).
		WithAuthToken(token).
		Branch(branch).Tree()

	return src, nil
}
