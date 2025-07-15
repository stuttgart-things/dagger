package main

import (
	"dagger/crane/internal/dagger"
	"fmt"
)

// authenticate adds registry authentication to the container
func authenticate(
	ctr *dagger.Container,
	registry *RegistryAuth,
	insecure bool) *dagger.Container {

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

// Helper function for conditional string selection
func ifThenElse(condition bool, a string, b string) string {
	if condition {
		return a
	}
	return b
}
