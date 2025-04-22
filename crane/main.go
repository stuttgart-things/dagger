package main

import (
	"context"
	"dagger/crane/internal/dagger"
	"fmt"
	"strings"
)

// Crane installs Crane CLI on a Wolfi base image at runtime
// @module
type Crane struct {
	// Base Wolfi image to use
	// +optional
	// +default="cgr.dev/chainguard/wolfi-base:latest"
	BaseImage string

	// Crane version to install (e.g., "latest" or specific version)
	// +optional
	// +default="latest"
	Version string
}

// RegistryAuth contains authentication details for a registry
type RegistryAuth struct {
	URL      string
	Username string
	Password *dagger.Secret
}

// Copy copies an image between registries with authentication
// This is the primary function that will be called from the CLI
// +call
func (m *Crane) Copy(
	ctx context.Context,
	// Source image reference (e.g., "harbor.example.com/test/redis:latest")
	source string,
	// Target image reference (e.g., "ghcr.io/test/redis")
	target string,
	// Source registry URL (extracted from source if empty)
	// +optional
	sourceRegistry string,
	// Username for source registry
	// +optional
	sourceUsername string,
	// Password for source registry
	// +optional
	sourcePassword *dagger.Secret,
	// Target registry URL (extracted from target if empty)
	// +optional
	targetRegistry string,
	// Username for target registry
	// +optional
	targetUsername string,
	// Password for target registry
	// +optional
	targetPassword *dagger.Secret,
	// Allow insecure registry connections
	// +optional
	// +flag
	// +default=false
	insecure bool,
	// Image platform
	// +optional
	// +flag
	// +default="linux/amd64"
	platform string,
) (string, error) {
	if platform == "" {
		platform = "linux/amd64"
	}

	// If registry URLs weren't provided explicitly, extract them from image references
	if sourceRegistry == "" {
		sourceRegistry = extractRegistry(source)
	}
	if targetRegistry == "" {
		targetRegistry = extractRegistry(target)
	}

	// Set up auth configurations
	var sourceAuth, targetAuth *RegistryAuth

	if sourceRegistry != "" && sourceUsername != "" && sourcePassword != nil {
		sourceAuth = &RegistryAuth{
			URL:      sourceRegistry,
			Username: sourceUsername,
			Password: sourcePassword,
		}
	}

	if targetRegistry != "" && targetUsername != "" && targetPassword != nil {
		targetAuth = &RegistryAuth{
			URL:      targetRegistry,
			Username: targetUsername,
			Password: targetPassword,
		}
	}

	return m.copyImage(ctx, source, target, sourceAuth, targetAuth, insecure, platform)
}

// Helper function to extract registry from image reference
func extractRegistry(imageRef string) string {
	parts := strings.Split(imageRef, "/")
	if len(parts) > 1 && (strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":")) {
		return parts[0]
	}
	return ""
}

// Internal function that performs the actual copy operation
func (m *Crane) copyImage(
	ctx context.Context,
	source string,
	target string,
	sourceAuth *RegistryAuth,
	targetAuth *RegistryAuth,
	insecure bool,
	platform string,
) (string, error) {
	ctr := m.container(insecure)

	// Authenticate with source registry if needed
	if sourceAuth != nil {
		ctr = authenticate(ctr, sourceAuth, insecure)
	}

	// Authenticate with target registry if needed
	if targetAuth != nil {
		ctr = authenticate(ctr, targetAuth, insecure)
	}

	// Build copy command
	cmd := []string{"crane", "copy", "--platform", platform}
	if insecure {
		cmd = append(cmd, "--insecure")
	}
	cmd = append(cmd, source, target)

	fmt.Println("Executing command:", strings.Join(cmd, " "))

	// Execute the copy
	result := ctr.WithExec(cmd)

	out, err := result.Stdout(ctx)
	if err != nil {
		stderr, _ := result.Stderr(ctx)
		return "", fmt.Errorf("copy failed: %w\nStdout: %s\nStderr: %s", err, out, stderr)
	}

	return out, nil
}

// authenticate adds registry authentication to the container
func authenticate(ctr *dagger.Container, registry *RegistryAuth, insecure bool) *dagger.Container {
	loginCmd := []string{
		"sh", "-c",
		fmt.Sprintf(`echo "$CRANE_PASSWORD" | crane auth login %s --username %s --password-stdin %s`,
			registry.URL,
			registry.Username,
			ifThenElse(insecure, "--insecure", ""),
		),
	}

	fmt.Printf("Authenticating with registry: %s as user: %s\n", registry.URL, registry.Username)

	return ctr.
		WithSecretVariable("CRANE_PASSWORD", registry.Password).
		WithExec(loginCmd)
}

// container returns a Wolfi-based container with Crane CLI installed
func (m *Crane) container(insecure bool) *dagger.Container {
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	ctr := dag.Container().From(m.BaseImage)

	pkg := "crane"
	ctr = ctr.WithExec([]string{"apk", "add", "--no-cache", pkg})
	ctr = ctr.WithEntrypoint([]string{"crane"})

	if insecure {
		ctr = ctr.WithEnvVariable("SSL_CERT_DIR", "/nonexistent")
	}

	return ctr
}

// Helper function for conditional string selection
func ifThenElse(condition bool, a string, b string) string {
	if condition {
		return a
	}
	return b
}
