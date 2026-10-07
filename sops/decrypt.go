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
	ageKey *dagger.Secret,
	encryptedFile *dagger.File,
	// .sops.yaml to use
	// +optional
	sopsConfig *dagger.File,
	// Return a single value instead of the whole file, as a sops path,
	// e.g. '["stringData"]["KEY"]'
	// +optional
	extract string,
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

	cmd := withSopsConfig(&ctr, sopsConfig, "-d", "--output", decryptedFile)
	if extract != "" {
		cmd = append(cmd, "--extract", extract)
	}

	// Provide the SOPS secret key (required for decryption)
	if ageKey != nil {
		ctr = ctr.WithSecretVariable("SOPS_AGE_KEY", ageKey)
	} else {
		return nil, fmt.Errorf("ageKey is required for decryption")
	}

	// Decrypt file to output file
	ctr = ctr.
		WithEntrypoint([]string{}).
		WithExec(append(cmd, fileName))

	return ctr.File(decryptedFile), nil
}
