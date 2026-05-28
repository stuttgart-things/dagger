package main

import (
	"context"
	"dagger/dependencies/internal/dagger"
	_ "embed"
	"fmt"
	"strings"
)

//go:embed scripts/check-ansible-updates.sh
var checkAnsibleUpdatesScript string

//go:embed scripts/update-ansible-requirements.sh
var updateAnsibleRequirementsScript string

// GalaxyCollection represents an Ansible Galaxy collection metadata
type GalaxyCollection struct {
	Name           string `json:"name"`
	HighestVersion struct {
		Version string `json:"version"`
	} `json:"highest_version"`
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

// UpdateAnsibleRequirements checks for updates to Ansible collections in requirements.yaml
// and returns a report of available updates
func (m *Dependencies) UpdateAnsibleRequirements(
	ctx context.Context,
	// Path to the requirements.yaml file to check
	requirementsFile *dagger.File,
	// Optional: GitHub token for checking custom collection releases
	// +optional
	githubToken *dagger.Secret,
) (string, error) {
	// Create a container with required tools using Wolfi
	container := dag.Container().
		From("cgr.dev/chainguard/wolfi-base:latest").
		WithExec([]string{"apk", "add", "--no-cache", "bash", "curl", "jq", "yq", "python-3", "py3-pip"}).
		WithExec([]string{"pip", "install", "--break-system-packages", "ansible"})

	// Mount the requirements file
	container = container.
		WithMountedFile("/work/requirements.yaml", requirementsFile).
		WithWorkdir("/work")

	// Add GitHub token if provided
	if githubToken != nil {
		container = container.WithSecretVariable("GITHUB_TOKEN", githubToken)
	}

	// Execute the script
	result, err := container.
		WithNewFile("/tmp/check-updates.sh", checkAnsibleUpdatesScript, dagger.ContainerWithNewFileOpts{
			Permissions: 0755,
		}).
		WithExec([]string{"/bin/bash", "/tmp/check-updates.sh"}).
		Stdout(ctx)

	if err != nil {
		// Try to get stderr for debugging
		stderr, _ := container.Stderr(ctx)
		if stderr != "" {
			return result + "\n\nSTDERR:\n" + stderr, err
		}
		return result, err
	}

	return result, nil
}

// UpdateAnsibleRequirementsAndApply checks for updates and applies them to the requirements file
func (m *Dependencies) UpdateAnsibleRequirementsAndApply(
	ctx context.Context,
	// Path to the requirements.yaml file to update
	requirementsFile *dagger.File,
	// Collections to update (comma-separated, e.g., "community.general,kubernetes.core")
	// Use "all" to update all collections
	collectionsToUpdate string,
	// Optional: GitHub token for checking custom collection releases
	// +optional
	githubToken *dagger.Secret,
) (*dagger.File, error) {
	// Create a container with required tools using Wolfi
	container := dag.Container().
		From("cgr.dev/chainguard/wolfi-base:latest").
		WithExec([]string{"apk", "add", "--no-cache", "bash", "curl", "jq", "yq", "python-3", "py3-pip"}).
		WithExec([]string{"pip", "install", "--break-system-packages", "ansible"})

	// Mount the requirements file and copy it to a writable location
	container = container.
		WithMountedFile("/tmp/input/requirements.yaml", requirementsFile).
		WithWorkdir("/work").
		WithExec([]string{"cp", "/tmp/input/requirements.yaml", "/work/requirements.yaml"})

	// Add GitHub token if provided
	if githubToken != nil {
		container = container.WithSecretVariable("GITHUB_TOKEN", githubToken)
	}

	// Validate and prepare collections list
	updateAll := strings.ToLower(collectionsToUpdate) == "all"
	collections := []string{}
	if !updateAll {
		collections = strings.Split(collectionsToUpdate, ",")
	}

	updateScript := fmt.Sprintf(updateAnsibleRequirementsScript, updateAll, strings.Join(collections, " "))

	// Execute the update script
	container = container.
		WithNewFile("/tmp/update.sh", updateScript, dagger.ContainerWithNewFileOpts{
			Permissions: 0755,
		}).
		WithExec([]string{"/bin/bash", "/tmp/update.sh"})

	// Return the updated file
	return container.File("/work/requirements.yaml"), nil
}

// ApplyAnsibleUpdates checks for updates and applies ALL available updates to the requirements file
// This is a convenience function that automatically updates all collections to their latest versions
func (m *Dependencies) ApplyAnsibleUpdates(
	ctx context.Context,
	// Path to the requirements.yaml file to update
	requirementsFile *dagger.File,
	// Optional: GitHub token for checking custom collection releases
	// +optional
	githubToken *dagger.Secret,
) (*dagger.File, error) {
	return m.UpdateAnsibleRequirementsAndApply(ctx, requirementsFile, "all", githubToken)
}
