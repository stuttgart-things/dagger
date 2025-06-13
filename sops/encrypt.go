package main

import (
	"context"
	"dagger/sops/internal/dagger"
	"fmt"
	"strings"
)

func (m *Sops) Encrypt(
	ctx context.Context,
	ageKey *dagger.Secret,
	plaintextFile *dagger.File,
	fileExtension string, // e.g., "yaml", "json", "env"
	sopsConfig *dagger.File, // Optional: ~/.sops.yaml config file
) (*dagger.File, error) {
	// Set default file extension to "yaml" if none provided
	if fileExtension == "" {
		fileExtension = "yaml"
	}
	fileExtension = strings.TrimPrefix(fileExtension, ".") // Sanitize extension

	ctr, err := m.container(ctx)
	if err != nil {
		return nil, fmt.Errorf("container init failed: %w", err)
	}

	workDir := "/src"
	plainFile := "plaintext." + fileExtension
	encryptedFile := "encrypted." + fileExtension

	// Mount the plaintext file
	ctr = ctr.
		WithMountedFile(workDir+"/"+plainFile, plaintextFile).
		WithWorkdir(workDir)

	// Mount the optional .sops.yaml config file
	if sopsConfig != nil {
		ctr = ctr.WithMountedFile("/root/.sops.yaml", sopsConfig)
	}

	// Provide the SOPS secret key (required for encryption)
	if ageKey != nil {
		ctr = ctr.WithSecretVariable("SOPS_AGE_RECIPIENTS", ageKey)
	} else {
		return nil, fmt.Errorf("ageKey is required for encryption")
	}

	// Copy file and encrypt it using sops
	ctr = ctr.
		WithEntrypoint([]string{}).
		WithExec([]string{"cp", plainFile, encryptedFile}).
		WithExec([]string{"sops", "--encrypt", "--in-place", encryptedFile})

	// Return the encrypted file
	return ctr.File(encryptedFile), nil
}
