package main

import (
	"context"
	"dagger/sops/internal/dagger"
	"fmt"
)

// Decrypt decrypts a SOPS-encrypted file using an AGE key.
// Returns the decrypted file.
func (m *Sops) Decrypt(
	ctx context.Context,
	sopsKey *dagger.Secret,
	encryptedFile *dagger.File,
) (*dagger.File, error) {
	ctr, err := m.container(ctx)
	if err != nil {
		return nil, fmt.Errorf("container init failed: %w", err)
	}

	workDir := "/src"
	fileName, err := encryptedFile.Name(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get file name: %w", err)
	}

	decryptedFile := "decrypted-" + fileName

	// Mount encrypted file into container
	ctr = ctr.
		WithMountedFile(workDir+"/"+fileName, encryptedFile).
		WithWorkdir(workDir)

	// Add SOPS key secret if provided
	if sopsKey != nil {
		ctr = ctr.WithSecretVariable("SOPS_AGE_KEY", sopsKey)
	}

	// Decrypt file to output file
	ctr = ctr.
		WithEntrypoint([]string{}).
		WithExec([]string{"sops", "-d", "--output", decryptedFile, fileName})

	return ctr.File(decryptedFile), nil
}
