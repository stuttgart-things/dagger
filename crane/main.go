// A generated module for Crane functions
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

	// Allow insecure registry connections
	// +optional
	Insecure bool

	// Registry to authenticate to
	// +optional
	Registry string

	// Username for registry authentication
	// +optional
	Username string

	// Password for registry authentication
	// +optional
	Password *dagger.Secret
}

func New(
	// +optional
	baseImage string,
	// +optional
	version string,
	// +optional
	insecure bool,
	// +optional
	registry string,
	// +optional
	username string,
	// +optional
	password *dagger.Secret,
) *Crane {
	return &Crane{
		BaseImage: baseImage,
		Version:   version,
		Insecure:  insecure,
		Registry:  registry,
		Username:  username,
		Password:  password,
	}
}

func (m *Crane) Test(
	ctx context.Context,
	password *dagger.Secret,
) {
	crane := New("cgr.dev/chainguard/wolfi-base:latest", "latest", true, "harbor.fluxdev-3.sthings-vsphere.labul.sva.de", "admin", password)
	output, err := crane.CopyImage(ctx, "redis:latest", "harbor.fluxdev-3.sthings-vsphere.labul.sva.de/test/redis")
	fmt.Println("Output:", output)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Copy successful")
}

// Container returns a Wolfi-based container with Crane CLI installed
func (w *Crane) Container() *dagger.Container {
	// Start from Wolfi base image
	ctr := dag.Container().From(w.BaseImage)

	// Determine package to install
	pkg := "crane"
	if w.Version != "latest" {
		pkg = fmt.Sprintf("crane-%s", w.Version)
	}

	// Install crane using apk (Wolfi's package manager)
	ctr = ctr.WithExec([]string{"apk", "add", "--no-cache", pkg})

	// Set crane as entrypoint
	ctr = ctr.WithEntrypoint([]string{"crane"})

	// Optional: Configure for insecure registries
	if w.Insecure {
		ctr = ctr.WithEnvVariable("SSL_CERT_DIR", "/nonexistent")
	}

	return ctr
}

func (m *Crane) CopyImage(
	ctx context.Context,
	source string,
	target string,
) (string, error) {
	ctr := m.Container()

	// Debug: Print current auth configuration
	fmt.Printf("Attempting auth with registry: %s, user: %s\n", m.Registry, m.Username)

	if m.Registry != "" && m.Username != "" && m.Password != nil {
		// Better approach using password-stdin
		loginCmd := []string{
			"sh", "-c",
			fmt.Sprintf(`echo "$CRANE_PASSWORD" | crane auth login %s --username %s --password-stdin %s`,
				m.Registry,
				m.Username,
				"--insecure",
			),
		}

		fmt.Println("Login command:", strings.Join(loginCmd, " "))

		ctr = ctr.
			WithSecretVariable("CRANE_PASSWORD", m.Password).
			WithExec(loginCmd)
	}

	// Debug command that will be executed
	cmd := []string{"crane", "copy"}
	if m.Insecure {
		cmd = append(cmd, "--insecure")
	}
	cmd = append(cmd, source, target)
	fmt.Println("Copy command:", strings.Join(cmd, " "))

	// Execute with debug output
	result := ctr.
		WithExec(cmd).
		WithExec([]string{"sh", "-c", "ls -la ~/.docker/config.json"}) // Debug: Check config file

	out, err := result.Stdout(ctx)
	if err != nil {
		// Get stderr for more detailed error
		stderr, _ := result.Stderr(ctx)
		fmt.Printf("Error details:\nStdout: %s\nStderr: %s\n", out, stderr)
		return "", fmt.Errorf("copy failed: %w", err)
	}

	return out, nil
}

func (m *Crane) TestAuth(
	ctx context.Context,
) (string, error) {
	// Test authentication independently
	if m.Registry == "" || m.Username == "" || m.Password == nil {
		return "", fmt.Errorf("registry credentials not provided")
	}

	ctr := m.Container().
		WithSecretVariable("CRANE_PASSWORD", m.Password).
		WithExec([]string{
			"sh", "-c",
			fmt.Sprintf(`echo "$CRANE_PASSWORD" | crane auth login %s --username %s --password-stdin %s`,
				m.Registry,
				m.Username,
				"--insecure",
			),
		})

	return ctr.Stdout(ctx)
}
