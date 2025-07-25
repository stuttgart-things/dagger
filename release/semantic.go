package main

import (
	"context"
	"dagger/release/internal/dagger"
)

// Semantic runs semantic-release using the specified folder and GitHub token.
// It first performs a dry run, then a real run with --no-ci.
func (m *Release) Semantic(
	ctx context.Context,
	// +optional
	// +default="1.0.18-light"
	semanticReleaseVersion string,
	// Source folder (e.g. ".")
	src *dagger.Directory,
	// +optional
	// +default="GITHUB_TOKEN"
	tokenName string,
	// +optional
	// +default=false
	dryRunOnly bool,
	// +optional
	token *dagger.Secret,
) (string, error) {

	var dryRun string
	releaseRun := "ONLY DRYRUN WAS EXECUTED"

	// Create container with all required plugins
	releaseContainer := m.container(semanticReleaseVersion).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithSecretVariable(tokenName, token)

	// Dry-run semantic-release
	dryRun, err := releaseContainer.
		WithExec([]string{
			"npx",
			"semantic-release",
			"--dry-run"}).
		Stdout(ctx)
	if err != nil {
		return "", err
	}

	if !dryRunOnly {
		// Real run with --no-ci
		releaseRun, err = releaseContainer.
			WithExec([]string{
				"npx",
				"semantic-release",
				"--debug",
				"--no-ci"}).
			Stdout(ctx)
		if err != nil {
			return "", err
		}
	}

	return dryRun + "\n\n" + releaseRun, nil
}
