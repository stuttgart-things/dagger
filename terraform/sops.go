package main

import (
	"context"
	"dagger/terraform/internal/dagger"
	"fmt"
)

func (m *Terraform) DecryptSops(
	ctx context.Context,
	sopsKey *dagger.Secret,
	encryptedFile *dagger.File,
) (string, error) {
	ctr, err := m.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	workDir := "/src"
	fileName := "encrypted.json"

	// Mount encrypted file into container using string concatenation
	ctr = ctr.
		WithMountedFile(workDir+"/"+fileName, encryptedFile).
		WithWorkdir(workDir)

	// Add SOPS key secret if provided
	if sopsKey != nil {
		ctr = ctr.WithSecretVariable("SOPS_AGE_KEY", sopsKey)
	}

	// Decrypt file
	out, err := ctr.
		WithEntrypoint([]string{}). // Clear terraform entrypoint
		WithExec([]string{"sops", "-d", fileName}).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("sops decryption failed: %w", err)
	}

	return out, nil
}
