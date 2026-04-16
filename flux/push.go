package main

import (
	"context"
	"dagger/flux/internal/dagger"
	reg "dagger/flux/registry"
	"fmt"
)

// PushArtifact builds and pushes an OCI artifact from a directory to a container registry
func (m *Flux) PushArtifact(
	ctx context.Context,
	// Source directory containing the artifact files
	src *dagger.Directory,
	// OCI artifact address (e.g., oci://ghcr.io/org/repo:tag)
	artifact string,
	// Registry URL for authentication (e.g., ghcr.io)
	registry string,
	// Registry username
	username string,
	// Registry password
	password *dagger.Secret, // pragma: allowlist secret
	// Source URL metadata (e.g., git remote URL)
	// +optional
	source string,
	// Revision metadata (e.g., branch@sha1:commit)
	// +optional
	revision string,
) (string, error) {

	passwordPlaintext, err := password.Plaintext(ctx) // pragma: allowlist secret
	if err != nil {
		return "", fmt.Errorf("failed to read password secret: %w", err)
	}

	configJSON, err := reg.CreateDockerConfigJSON(username, passwordPlaintext, registry)
	if err != nil {
		return "", fmt.Errorf("failed to create Docker config.json: %w", err)
	}

	fluxContainer := m.container()

	cmd := []string{
		"flux", "push", "artifact",
		artifact,
		"--path=/workspace",
	}

	if source != "" {
		cmd = append(cmd, fmt.Sprintf("--source=%s", source))
	}

	if revision != "" {
		cmd = append(cmd, fmt.Sprintf("--revision=%s", revision))
	}

	result, err := fluxContainer.
		WithNewFile("/root/.docker/config.json", configJSON).
		WithDirectory("/workspace", src).
		WithWorkdir("/workspace").
		WithExec(cmd).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to push artifact: %w", err)
	}

	return result, nil
}

// PushArtifacts builds and pushes OCI artifacts from multiple directories to a container registry.
// Each directory name is appended to the base artifact address as a tag.
func (m *Flux) PushArtifacts(
	ctx context.Context,
	// Source directory containing subdirectories, each becoming an OCI artifact
	src *dagger.Directory,
	// Base OCI artifact address without tag (e.g., oci://ghcr.io/org/repo)
	artifact string,
	// Tag to use for all artifacts
	tag string,
	// Registry URL for authentication (e.g., ghcr.io)
	registry string,
	// Registry username
	username string,
	// Registry password
	password *dagger.Secret, // pragma: allowlist secret
	// Source URL metadata (e.g., git remote URL)
	// +optional
	source string,
	// Revision metadata (e.g., branch@sha1:commit)
	// +optional
	revision string,
) (string, error) {

	passwordPlaintext, err := password.Plaintext(ctx) // pragma: allowlist secret
	if err != nil {
		return "", fmt.Errorf("failed to read password secret: %w", err)
	}

	configJSON, err := reg.CreateDockerConfigJSON(username, passwordPlaintext, registry)
	if err != nil {
		return "", fmt.Errorf("failed to create Docker config.json: %w", err)
	}

	entries, err := src.Entries(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list source directories: %w", err)
	}

	fluxContainer := m.container().
		WithNewFile("/root/.docker/config.json", configJSON)

	var output string

	for _, entry := range entries {
		subDir := src.Directory(entry)

		artifactAddr := fmt.Sprintf("%s/%s:%s", artifact, entry, tag)

		cmd := []string{
			"flux", "push", "artifact",
			artifactAddr,
			"--path=/workspace",
		}

		if source != "" {
			cmd = append(cmd, fmt.Sprintf("--source=%s", source))
		}

		if revision != "" {
			cmd = append(cmd, fmt.Sprintf("--revision=%s", revision))
		}

		result, err := fluxContainer.
			WithDirectory("/workspace", subDir).
			WithWorkdir("/workspace").
			WithExec(cmd).
			Stdout(ctx)
		if err != nil {
			return output, fmt.Errorf("failed to push artifact %s: %w", artifactAddr, err)
		}

		output += fmt.Sprintf("pushed %s\n%s\n", artifactAddr, result)
	}

	return output, nil
}
