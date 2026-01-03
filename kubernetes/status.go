package main

import (
	"context"
	"dagger/kubernetes/internal/dagger"
	"fmt"
	"strings"
)

func (m *Kubernetes) CheckResourceStatus(
	ctx context.Context,
	kind string, // e.g., "secret", "pod"
	name string, // resource name
	namespace string, // optional namespace
	kubeConfig *dagger.Secret,
) (bool, error) {
	// Combine kind and name as kubectl expects
	resource := fmt.Sprintf("%s %s", kind, name)

	out, err := m.Command(
		ctx,
		"get",
		resource,
		namespace,
		kubeConfig,
		"",
		false, // Don't ignore errors, we need to detect if resource doesn't exist
	)

	// Check if the output contains "not found" - this means the resource doesn't exist
	if strings.Contains(out, "not found") || strings.Contains(out, "NotFound") {
		return false, nil
	}

	// If there's an error but no "not found" message, it's a different kind of error
	if err != nil {
		return false, err
	}

	// If we got here with no error and no "not found", the resource exists
	return true, nil
}
