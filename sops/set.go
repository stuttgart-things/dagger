package main

import (
	"context"
	"crypto/sha256"
	"dagger/sops/internal/dagger"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Set changes one value in a SOPS-encrypted file without writing the rest out
// in plaintext (`sops set`). Returns the updated encrypted file.
//
// The value never becomes an operation argument: it is JSON-encoded here,
// wrapped as a Secret and handed to sops with --value-file.
func (m *Sops) Set(
	ctx context.Context,
	// SOPS-encrypted file to change
	encryptedFile *dagger.File,
	// sops path of the value, e.g. '["stringData"]["KEY"]'
	path string,
	// The new value. Treated as a string unless --json-value is set.
	value *dagger.Secret,
	// AGE private key that can decrypt the file
	ageKey *dagger.Secret,
	// The value is already JSON (an object, a number, ...) and is set as is
	// +optional
	jsonValue bool,
) (*dagger.File, error) {
	if path == "" {
		return nil, fmt.Errorf("path is required, e.g. '[\"stringData\"][\"KEY\"]'")
	}

	encoded, err := encodeSetValue(ctx, value, jsonValue)
	if err != nil {
		return nil, err
	}

	ctr, err := m.container(ctx)
	if err != nil {
		return nil, fmt.Errorf("container init failed: %w", err)
	}

	fileName, err := encryptedFile.Name(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get file name: %w", err)
	}

	valuePath := "/run/sops/value.json"
	ctr = ctr.
		WithFile("/src/"+fileName, encryptedFile).
		WithWorkdir("/src").
		WithMountedSecret(valuePath, encoded).
		WithSecretVariable("SOPS_AGE_KEY", ageKey).
		WithEntrypoint([]string{}).
		WithExec([]string{"sops", "set", "--value-file", fileName, path, valuePath})

	return ctr.File("/src/" + fileName), nil
}

// encodeSetValue returns the value as the JSON sops set expects, as a Secret.
func encodeSetValue(ctx context.Context, value *dagger.Secret, jsonValue bool) (*dagger.Secret, error) {
	plaintext, err := value.Plaintext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read value: %w", err)
	}

	encoded := []byte(plaintext)
	if jsonValue {
		if !json.Valid(encoded) {
			return nil, fmt.Errorf("--json-value is set, but the value is not valid JSON")
		}
	} else {
		encoded, err = json.Marshal(plaintext)
		if err != nil {
			return nil, fmt.Errorf("failed to encode value: %w", err)
		}
	}

	// Content-derived name, so two values never share one secret name.
	sum := sha256.Sum256(encoded)
	return dag.SetSecret("sops-set-value-"+hex.EncodeToString(sum[:8]), string(encoded)), nil
}
