// A generated module for Release functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/release/internal/dagger"
)

type Release struct {
	// Base Wolfi image to use
	// +optional
	// +default="1.0.18-light"
	SemanticReleaseVersion string
	// +optional
	// +default="hoppr/semantic-release"
	BaseImage string
}

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
	token *dagger.Secret,
) (string, error) {

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

	// Real run with --no-ci
	finalRun, err := releaseContainer.
		WithExec([]string{
			"npx",
			"semantic-release",
			"--debug",
			"--no-ci"}).
		Stdout(ctx)
	if err != nil {
		return "", err
	}

	return dryRun + "\n\n" + finalRun, nil
}
