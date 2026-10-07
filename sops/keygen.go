package main

import (
	"context"
	"dagger/sops/internal/dagger"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// GenerateAgeKey generates a new AGE key pair using age-keygen.
// Returns the key file containing both the public key (in a comment) and the private key.
//
// Never cached: with Dagger's function and exec cache, every call on the same
// engine used to return the same private key.
// +cache="never"
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
		// A per-call value, so the exec is never served from cache either.
		WithEnvVariable("KEYGEN_NONCE", strconv.FormatInt(time.Now().UnixNano(), 10)).
		WithExec([]string{"age-keygen", "-o", keyFile})

	// Sync to validate execution succeeded
	_, err = ctr.Sync(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate age key: %w", err)
	}

	return ctr.File(keyFile), nil
}

// AgePublicKey derives the public key (age1...) from an AGE private key
// (`age-keygen -y`).
func (m *Sops) AgePublicKey(
	ctx context.Context,
	// AGE private key (AGE-SECRET-KEY-1...) or a key file with comments
	ageKey *dagger.Secret,
) (string, error) {
	ctr, err := m.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	out, err := ctr.
		WithMountedSecret("/run/sops/age-key.txt", ageKey).
		WithEntrypoint([]string{}).
		WithExec([]string{"age-keygen", "-y", "/run/sops/age-key.txt"}).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to derive public key: %w", err)
	}

	return strings.TrimSpace(out), nil
}
