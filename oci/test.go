package main

import (
	"context"
	"fmt"
	"strings"
)

// TestPushArtifact starts a local zot registry with TLS and pushes a test artifact using flux to verify the push workflow.
// No external registry credentials are required.
func (m *Oci) TestPushArtifact(
	ctx context.Context,
	// Zot registry image to use
	// +optional
	// +default="ghcr.io/project-zot/zot-linux-amd64:latest"
	registryImage string,
) (string, error) {
	tlsDir := m.tlsCerts()
	registry := m.RegistryService(registryImage, 5000)

	testDir := dag.Directory().
		WithNewFile("test.yaml", "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n")

	fluxContainer := m.container().
		WithServiceBinding("registry", registry).
		WithFile("/usr/local/share/ca-certificates/registry.crt", tlsDir.File("cert.pem")).
		WithExec([]string{"update-ca-certificates"})

	result, err := fluxContainer.
		WithDirectory("/workspace", testDir).
		WithWorkdir("/workspace").
		WithExec([]string{
			"flux", "push", "artifact",
			"oci://registry:5000/test/artifact:v1.0.0",
			"--path=/workspace",
			"--source=local",
			"--revision=test",
		}).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to push test artifact: %w", err)
	}

	return fmt.Sprintf("test push-artifact: OK\n%s", result), nil
}

// TestPushArtifacts starts a local zot registry with TLS and pushes multiple test artifacts using flux to verify the batch push workflow.
// No external registry credentials are required.
func (m *Oci) TestPushArtifacts(
	ctx context.Context,
	// Zot registry image to use
	// +optional
	// +default="ghcr.io/project-zot/zot-linux-amd64:latest"
	registryImage string,
) (string, error) {
	tlsDir := m.tlsCerts()
	registry := m.RegistryService(registryImage, 5000)

	testDir := dag.Directory().
		WithNewFile("app-config/deployment.yaml", "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: test\n").
		WithNewFile("monitoring/prometheus.yaml", "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: prometheus\n")

	fluxContainer := m.container().
		WithServiceBinding("registry", registry).
		WithFile("/usr/local/share/ca-certificates/registry.crt", tlsDir.File("cert.pem")).
		WithExec([]string{"update-ca-certificates"})

	entries, err := testDir.Entries(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list test directories: %w", err)
	}

	var output string

	for _, entry := range entries {
		subDir := testDir.Directory(entry)
		entryName := strings.TrimRight(entry, "/")
		artifactAddr := fmt.Sprintf("oci://registry:5000/test/artifacts/%s:v1.0.0", entryName)

		result, err := fluxContainer.
			WithDirectory("/workspace", subDir).
			WithWorkdir("/workspace").
			WithExec([]string{
				"flux", "push", "artifact",
				artifactAddr,
				"--path=/workspace",
				"--source=local",
				"--revision=test",
			}).
			Stdout(ctx)
		if err != nil {
			return output, fmt.Errorf("failed to push test artifact %s: %w", artifactAddr, err)
		}

		output += fmt.Sprintf("pushed %s\n%s\n", artifactAddr, result)
	}

	return fmt.Sprintf("test push-artifacts: OK\n%s", output), nil
}
