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

	_, err := m.Command(
		ctx,
		"get",
		resource,
		namespace,
		kubeConfig,
		"", // no additionalCommand needed
	)

	if err != nil {
		// kubectl returns an error if resource does not exist
		// we can check stderr or simply treat any error as "not found"
		// optional: you could inspect err to differentiate "not found" vs other errors
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
