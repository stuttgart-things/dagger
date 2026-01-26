package main

import (
	"context"
	"dagger/sops/internal/dagger"
	"fmt"
)

// GenerateAgeKey generates a new AGE key pair using age-keygen.
// Returns the key file containing both the public key (in a comment) and the private key.
func (m *Sops) GenerateAgeKey(
	ctx context.Context,
) (*dagger.File, error) {
	ctr, err := m.container(ctx)
	if err != nil {
		return nil, fmt.Errorf("container init failed: %w", err)
	}

	ctr = ctr.
		WithExec([]string{"apk", "add", "--no-cache", "age"}).
		WithEntrypoint([]string{}).
		WithExec([]string{"age-keygen", "-o", "/tmp/age-key.txt"})

	return ctr.File("/tmp/age-key.txt"), nil
}
