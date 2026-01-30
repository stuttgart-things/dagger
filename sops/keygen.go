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

	keyFile := "/tmp/age-key.txt"

	ctr = ctr.
		WithEntrypoint([]string{}).
		WithExec([]string{"apk", "add", "--no-cache", "age"}).
		WithExec([]string{"age-keygen", "-o", keyFile})

	// Sync to validate execution succeeded
	_, err = ctr.Sync(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate age key: %w", err)
	}

	return ctr.File(keyFile), nil
}
