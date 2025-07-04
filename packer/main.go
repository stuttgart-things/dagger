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

func (m *Packer) Bake(
	ctx context.Context,
	// The Packer version to use
	// +optional
	// +default="1.12.0"
	packerVersion string,
	// The Packer arch
	// +optional
	// +default="linux_amd64"
	arch string,
	// If true, only init packer w/out build
	// +optional
	// +default=false
	initOnly bool,
	// vaultAddr
	// +optional
	vaultAddr string,
	// vaultRoleID
	// +optional
	vaultRoleID *dagger.Secret,
	// vaultSecretID
	// +optional
	vaultSecretID *dagger.Secret,
	// vaultToken
	// +optional
	vaultToken *dagger.Secret,
	buildPath string,
	localDir *dagger.Directory,
) string {
	workingDir := filepath.Dir(buildPath)
	packerFile := filepath.Base(buildPath)

	repoContent := localDir
	buildDir := repoContent.Directory(workingDir)

	logFilePath := "/src/packer.log"

	// PREPARE BASE CONTAINER WITH ENVIRONMENT AND SECRETS
	base := m.container(packerVersion, arch).
		WithMountedDirectory("/src", buildDir).
		WithWorkdir("/src").
		WithEnvVariable("VAULT_ADDR", vaultAddr).
		WithEnvVariable("VAULT_SKIP_VERIFY", "TRUE").
		WithEnvVariable("PACKER_LOG", "1").
		WithEnvVariable("PACKER_LOG_PATH", logFilePath)

	if vaultToken != nil {
		base = base.WithSecretVariable("VAULT_TOKEN", vaultToken)
	}
	if vaultRoleID != nil {
		base = base.WithSecretVariable("VAULT_ROLE_ID", vaultRoleID)
	}
	if vaultSecretID != nil {
		base = base.WithSecretVariable("VAULT_SECRET_ID", vaultSecretID)
	}

	// RUN `PACKER INIT`
	initContainer := base.WithExec([]string{"packer", "init", packerFile})

	// RUN `PACKER BUILD` UNLESS INITONLY IS TRUE
	var buildContainer *dagger.Container
	if !initOnly {
		buildContainer = initContainer.WithExec([]string{"packer", "build", packerFile})
	} else {
		buildContainer = initContainer
	}

	// READ THE PACKER LOG FROM THE CONTAINER
	logContents, err := buildContainer.
		File(logFilePath).
		Contents(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to read packer log: %w", err))
	}

	return logContents
}
