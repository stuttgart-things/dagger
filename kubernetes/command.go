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
	// +optional
	// +default=false
	ignoreErrors bool,
) (string, error) {

	kubeConfigPath := "/root/.kube/config"

	// Build kubectl command arguments
	parts := strings.Fields(resourceKind)
	args := append([]string{"kubectl", operation}, parts...)

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
		// Redirect stderr to stdout so we capture error messages
		// Use '|| true' to ignore exit codes and still get the output
		fullCommand := fmt.Sprintf("(%s 2>&1) || true; %s", strings.Join(args, " "), additionalCommand)
		kubectlContainer = kubectlContainer.WithExec([]string{"sh", "-c", fullCommand})
	} else {
		// Execute kubectl command directly through sh to capture stderr in stdout
		// Use '|| true' to ignore exit codes, so we can always capture the output
		fullCommand := fmt.Sprintf("(%s 2>&1) || true", strings.Join(args, " "))
		kubectlContainer = kubectlContainer.WithExec([]string{"sh", "-c", fullCommand})
	}

	// Run the command and capture output
	out, err := kubectlContainer.Stdout(ctx)

	if err != nil {
		if ignoreErrors {
			// If ignoring errors, return output with no error
			fmt.Printf("kubectl warning (ignoring error): %v\n", err)
			return out, nil
		}

		// Return the error but include the output for the caller to inspect
		// (The CheckResourceStatus function needs to check for "not found" in the output)
		return out, fmt.Errorf("kubectl error: %w", err)
	}

	return out, nil
}

// Kubectl applies or manages Kubernetes manifests from files or URLs
func (m *Kubernetes) Kubectl(
	ctx context.Context,
	// Kubectl operation (apply, delete, create, etc.)
	// +optional
	// +default="apply"
	operation string,
	// Source file (local file from Dagger)
	// +optional
	sourceFile *dagger.File,
	// URL source (e.g., https://raw.githubusercontent.com/org/repo/main/manifest.yaml)
	// +optional
	urlSource string,
	// Kustomize directory URL (e.g., https://github.com/org/repo/path/to/kustomize)
	// +optional
	kustomizeSource string,
	// Namespace for the operation
	// +optional
	namespace string,
	// Kubeconfig secret for authentication
	// +optional
	kubeConfig *dagger.Secret,
	// Use server-side apply (only valid with apply operation)
	// +optional
	// +default=false
	serverSide bool,
	// Additional kubectl flags (e.g., "--dry-run=client -o yaml")
	// +optional
	additionalFlags string,
) (string, error) {

	if operation == "" {
		operation = "apply"
	}

	kubeConfigPath := "/root/.kube/config"
	manifestPath := "/tmp/manifest.yaml"

	kubectlContainer := m.container().
		WithMountedSecret(kubeConfigPath, kubeConfig)

	// Build kubectl command
	var args []string

	// Handle source: either local file, URL, or kustomize source
	if sourceFile != nil {
		// Mount the file and use it
		kubectlContainer = kubectlContainer.WithMountedFile(manifestPath, sourceFile)
		args = []string{"kubectl", operation, "-f", manifestPath}
	} else if kustomizeSource != "" {
		// Use kustomize source with -k flag
		args = []string{"kubectl", operation, "-k", kustomizeSource}
	} else if urlSource != "" {
		// Use URL directly - kubectl supports this natively
		args = []string{"kubectl", operation, "-f", urlSource}
	} else {
		return "", fmt.Errorf("either sourceFile, urlSource, or kustomizeSource must be provided")
	}

	// Add server-side flag if specified (only valid with apply operation)
	if serverSide && operation == "apply" {
		args = append(args, "--server-side")
	}

	// Add namespace if specified
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	// Add additional flags if provided
	if additionalFlags != "" {
		// Split additional flags and append them
		flags := strings.Fields(additionalFlags)
		args = append(args, flags...)
	}

	// Execute kubectl command
	kubectlContainer = kubectlContainer.WithExec(args)

	// Run the command and capture output
	out, err := kubectlContainer.Stdout(ctx)
	if err != nil {
		// also capture stderr for debugging
		stderr, _ := kubectlContainer.Stderr(ctx)
		return "", fmt.Errorf("kubectl error: %w\nstderr: %s", err, stderr)
	}

	return out, nil
}
