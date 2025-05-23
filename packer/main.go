package main

import (
	"context"
	"dagger/packer/internal/dagger"
	"fmt"
	"path/filepath"
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
	// +optional
	repoURL,
	// The Branch name
	// +optional
	// +default="main"
	branch string,
	// If true, only init packer w/out build
	// +optional
	// +default=false
	initOnly bool,
	// vaultAddr
	// +optional
	vaultAddr string,
	// vaultRoleID
	// +optional
	vaultRoleID string,
	// vaultSecretID
	// +optional
	vaultSecretID *dagger.Secret,
	// vaultToken
	// +optional
	vaultToken *dagger.Secret,
	buildPath string,
	// +optional
	token *dagger.Secret, // injected securely
	// +optional
	localDir *dagger.Directory, // NEW: optional local directory
) {

	workingDir := filepath.Dir(buildPath)
	packerFile := filepath.Base(buildPath)

	var repoContent *dagger.Directory
	var err error

	if localDir != nil {
		repoContent = localDir
	} else {
		repoContent, err = m.ClonePrivateRepo(ctx, repoURL, branch, token)
		if err != nil {
			panic(fmt.Errorf("failed to clone repo: %w", err))
		}
	}
	buildDir := repoContent.Directory(workingDir)

	entries1, err := buildDir.Entries(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("Files in buildPath:", entries1)

	// MOUNT BUILDDIR AND SET WORKING DIRECTORY
	base := m.container(packerVersion, arch).
		WithMountedDirectory("/src", buildDir).
		WithEnvVariable("VAULT_ADDR", vaultAddr).
		WithEnvVariable("VAULT_ROLE_ID", vaultRoleID).
		WithEnvVariable("VAULT_SKIP_VERIFY", "TRUE").
		WithSecretVariable("VAULT_TOKEN", vaultToken).
		WithSecretVariable("VAULT_SECRET_ID", vaultSecretID).
		WithWorkdir("/src")

	// RUN PACKER INIT AND PERSIST CONTAINER STATE
	packerContainer := base.WithExec([]string{"packer", "init", packerFile})

	// NOW RUN BUILD ON THE RESULT OF INIT
	if !initOnly {
		buildOut, err := packerContainer.
			WithExec([]string{"packer", "build", packerFile}).
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

	// Install base packages + Ansible dependencies
	ctr = ctr.WithExec([]string{"apk", "add", "--no-cache",
		"wget",
		"unzip",
		"bash",
		"coreutils",
		"python3",
		"py3-pip",
		"openssh-client",
		"ca-certificates-bundle",
		"cdrkit",
		"git",
		"sshpass",
	})

	// Install Ansible via pip
	ctr = ctr.WithExec([]string{"pip3", "install", "--no-cache-dir", "ansible", "hvac", "passlib"})

	// Install Packer
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
