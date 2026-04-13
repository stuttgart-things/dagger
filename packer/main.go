package main

import (
	"context"
	"dagger/packer/internal/dagger"
	"fmt"
	"path/filepath"
	"strings"
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
	// If true, force overwrite of existing template
	// +optional
	// +default=false
	force bool,
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
	// Comma-separated packer build vars, e.g. "name=tpl,username=root,password=foo" # pragma: allowlist secret
	// +optional
	vars string,
	// Path (relative to the packer build dir) to a plain YAML file with build vars
	// +optional
	varsFile string,
	// Path (relative to the packer build dir) to a SOPS-encrypted YAML file with secret build vars
	// +optional
	sopsFile string,
	// Age private key used to decrypt sopsFile (SOPS_AGE_KEY)
	// +optional
	sopsAgeKey *dagger.Secret,
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
	if vaultSecretID != nil { // pragma: allowlist secret
		base = base.WithSecretVariable("VAULT_SECRET_ID", vaultSecretID)
	}
	if sopsAgeKey != nil {
		base = base.WithSecretVariable("SOPS_AGE_KEY", sopsAgeKey)
	}

	// CONVERT PLAIN YAML VARS FILE -> JSON (packer -var-file accepts JSON)
	if varsFile != "" {
		base = base.WithExec([]string{
			"sh", "-c",
			fmt.Sprintf("yq -o=json '.' %s > /tmp/packer-vars.json", shellQuote(varsFile)),
		})
	}

	// DECRYPT SOPS YAML -> JSON
	if sopsFile != "" {
		base = base.WithExec([]string{
			"sh", "-c",
			fmt.Sprintf("sops -d %s | yq -o=json '.' > /tmp/packer-sops.json", shellQuote(sopsFile)),
		})
	}

	// RUN `PACKER INIT`
	initContainer := base.WithExec([]string{
		"packer",
		"init",
		packerFile,
	})

	// RUN `PACKER BUILD` UNLESS INITONLY IS TRUE
	var buildContainer *dagger.Container
	if !initOnly {
		buildArgs := []string{"packer", "build"}
		if force {
			buildArgs = append(buildArgs, "-force")
		}
		if varsFile != "" {
			buildArgs = append(buildArgs, "-var-file=/tmp/packer-vars.json")
		}
		if sopsFile != "" {
			buildArgs = append(buildArgs, "-var-file=/tmp/packer-sops.json")
		}
		for _, kv := range splitCSV(vars) {
			buildArgs = append(buildArgs, "-var", kv)
		}
		buildArgs = append(buildArgs, packerFile)

		buildContainer = initContainer.WithExec(
			buildArgs,
			dagger.ContainerWithExecOpts{
				Expect: dagger.ReturnTypeAny,
			},
		)
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

	// CHECK EXIT CODE AND INCLUDE IT IN OUTPUT IF FAILED
	exitCode, err := buildContainer.ExitCode(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to get exit code: %w", err))
	}

	if exitCode != 0 {
		return fmt.Sprintf("BUILD FAILED (exit code: %d)\n\n%s", exitCode, logContents)
	}

	return logContents
}

// splitCSV splits "k1=v1,k2=v2" into ["k1=v1","k2=v2"], skipping empties.
func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// shellQuote wraps a value in single quotes for safe shell interpolation.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
