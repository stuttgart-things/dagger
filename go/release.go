package main

import (
	"context"
	"dagger/go/internal/dagger"
	"fmt"
	"strings"
)

func (m *Go) Release(
	ctx context.Context,
	// Source code directory
	src *dagger.Directory,
	// GitHub token for authentication
	token *dagger.Secret,
	// +optional
	// +default="v2.13.2"
	releaserVersion string,
	// +optional
	// +default=false
	// Perform a dry run (snapshot) without publishing
	snapshot bool,
	// +optional
	// +default=false
	// Skip validation checks
	skipValidate bool,
	// +optional
	// +default=false
	// Run only goreleaser check without releasing
	checkOnly bool,
) (*dagger.File, error) {

	container := m.GetReleaserContainer(releaserVersion).
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithSecretVariable("GITHUB_TOKEN", token)

	logFile := "/tmp/goreleaser.log"

	// Run goreleaser check and log output
	checkCmd := fmt.Sprintf("goreleaser check > %s 2>&1 || true", logFile)
	container = container.WithExec([]string{"sh", "-c", checkCmd})

	// If checkOnly is true, return early with just the check results
	if checkOnly {
		return container.File(logFile), nil
	}

	// Build release command
	args := []string{"goreleaser", "--clean"}
	if snapshot {
		args = append(args, "--snapshot")
	}
	if skipValidate {
		args = append(args, "--skip=validate")
	}

	// Run release and append to log
	releaseCmd := fmt.Sprintf("%s >> %s 2>&1 || true", strings.Join(args, " "), logFile)
	container = container.WithExec([]string{"sh", "-c", releaseCmd})

	// Return the log file
	return container.File(logFile), nil
}
