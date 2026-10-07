package main

import (
	"context"
	"dagger/sops/internal/dagger"
	"fmt"
)

// UpdateKeys re-encrypts a SOPS file's data key for the recipients the
// .sops.yaml now lists (`sops updatekeys -y`), e.g. after adding a recipient.
// Returns the updated encrypted file.
func (m *Sops) UpdateKeys(
	ctx context.Context,
	// SOPS-encrypted file to update
	encryptedFile *dagger.File,
	// .sops.yaml with the new recipients. Its creation rule must match the
	// file's name.
	sopsConfig *dagger.File,
	// AGE private key that can decrypt the file
	ageKey *dagger.Secret,
) (*dagger.File, error) {
	ctr, err := m.container(ctx)
	if err != nil {
		return nil, fmt.Errorf("container init failed: %w", err)
	}

	fileName, err := encryptedFile.Name(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get file name: %w", err)
	}

	ctr = ctr.
		WithFile("/src/"+fileName, encryptedFile).
		WithWorkdir("/src").
		WithSecretVariable("SOPS_AGE_KEY", ageKey)

	cmd := withSopsConfig(&ctr, sopsConfig, "updatekeys", "-y", fileName)

	ctr = ctr.
		WithEntrypoint([]string{}).
		WithExec(cmd)

	return ctr.File("/src/" + fileName), nil
}
