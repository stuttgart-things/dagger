package main

// dagger call -m packer build --repo-url https://github.com/stuttgart-things/stuttgart-things.git --branch "feat/packer-hello" --token env:GITHUB_TOKEN --build-path packer/builds/hello  --progress plain -vv

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
	buildPath string,
	token *dagger.Secret, // injected securely
) {

	repoContent, err := m.ClonePrivateRepo(ctx, repoURL, branch, token)
	if err != nil {
		fmt.Errorf("failed to clone repo: %w", err)
	}

	// entries, err := repoContent.Entries(ctx)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Top-level entries:", entries)

	buildDir := repoContent.Directory(buildPath)

	entries1, err := buildDir.Entries(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("Files in buildPath:", entries1)

	// Mount buildDir and set working directory
	base := m.container(packerVersion, arch).
		WithMountedDirectory("/src", buildDir).
		WithWorkdir("/src")

	// Run packer init and persist container state
	initContainer := base.WithExec([]string{"packer", "init", "hello.pkr.hcl"})

	// Optionally get init output (from a separate execution)
	initOut, err := initContainer.WithExec([]string{"packer", "version"}).Stdout(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to verify init: %w", err))
	}
	fmt.Println("Init complete - Packer version:", initOut)

	// Now run build on the result of init
	if !initOnly {
		buildOut, err := initContainer.
			WithExec([]string{"packer", "build", "hello.pkr.hcl"}).
			Stdout(ctx)
		if err != nil {
			panic(fmt.Errorf("failed to build: %w", err))
		}
		fmt.Println(buildOut)
	}

	if err != nil {
		fmt.Errorf("failed to initialize: %w", err)
	}

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

	ctr = ctr.WithExec([]string{"apk", "add", "--no-cache", "wget", "unzip", "bash", "coreutils"})
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
