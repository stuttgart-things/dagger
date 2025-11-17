package main

import (
	"context"
	"dagger/kubernetes/internal/dagger"
	"fmt"
	"strings"
)

func (m *Kubernetes) Command(
	ctx context.Context,
	// src *dagger.Directory,
	// +optional
	// +default="get"
	operation string,
	// +optional
	// +default="pods"
	resourceKind string,
	// +optional
	namespace string,
	// +optional
	kubeConfig *dagger.Secret,
	// +optional
	additionalCommand string,
) (string, error) {

	kubeConfigPath := "/root/.kube/config"

	// Build kubectl command arguments
	args := []string{"kubectl", operation, resourceKind}

	// Handle namespace options:
	// - Empty string: no namespace flag (for cluster-wide resources)
	// - Specific namespace: use -n flag
	// - Special values for all namespaces: use -A flag
	if namespace != "" {
		// Check if user wants all namespaces
		if namespace == "ALL" || namespace == "all" || namespace == "*" {
			args = append(args, "-A")
		} else {
			args = append(args, "-n", namespace)
		}
	}

	kubectlContainer := m.container().
		WithMountedSecret(kubeConfigPath, kubeConfig)

	// If additional command is provided, pipe kubectl output through it
	if additionalCommand != "" {
		// Build a shell command that pipes kubectl through the additional command
		fullCommand := fmt.Sprintf("%s | %s", strings.Join(args, " "), additionalCommand)
		kubectlContainer = kubectlContainer.WithExec([]string{"sh", "-c", fullCommand})
	} else {
		// Execute kubectl command directly
		kubectlContainer = kubectlContainer.WithExec(args)
	}

	// Run the command and capture output
	out, err := kubectlContainer.Stdout(ctx)
	if err != nil {
		// also capture stderr for debugging
		stderr, _ := kubectlContainer.Stderr(ctx)
		return "", fmt.Errorf("kubectl error: %w\nstderr: %s", err, stderr)
	}

	return out, nil
}
