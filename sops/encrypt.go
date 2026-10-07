package main

import (
	"context"
	"dagger/sops/internal/dagger"
	"fmt"
	"strings"
)

// Encrypt encrypts a file with SOPS for the given AGE recipient(s), or for the
// recipients a --sops-config names. With --sops-config or --encrypted-regex
// only matching values are encrypted, so a Kubernetes Secret keeps
// apiVersion, kind and metadata readable.
func (m *Sops) Encrypt(
	ctx context.Context,
	// AGE public key(s), comma-separated. Optional with --sops-config: then
	// the recipients of the matching creation rule are used. When both are
	// given, these keys replace the config's recipients (sops treats keys
	// from the environment like --age).
	// +optional
	ageKey *dagger.Secret,
	plaintextFile *dagger.File,
	// +optional
	// +default="yaml"
	fileExtension string, // e.g., "yaml", "json", "env"
	// .sops.yaml whose creation_rules apply (recipients, encrypted_regex, ...)
	// +optional
	sopsConfig *dagger.File,
	// Encrypt only the values whose keys match, e.g. '^(data|stringData)$'
	// for a Kubernetes Secret that Flux applies. Overrides the config.
	// +optional
	encryptedRegex string,
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

	cmd := withSopsConfig(&ctr, sopsConfig, "--encrypt", "--in-place")
	if encryptedRegex != "" {
		cmd = append(cmd, "--encrypted-regex", encryptedRegex)
	}

	// Recipients: from --age-key, or else from the config's creation rule.
	switch {
	case ageKey != nil:
		ctr = ctr.WithSecretVariable("SOPS_AGE_RECIPIENTS", ageKey)
	case sopsConfig == nil:
		return nil, fmt.Errorf("ageKey or sopsConfig is required for encryption")
	}

	// Copy file and encrypt it using sops
	ctr = ctr.
		WithEntrypoint([]string{}).
		WithExec([]string{"cp", plainFile, encryptedFile}).
		WithExec(append(cmd, encryptedFile))

	// Return the encrypted file
	return ctr.File(encryptedFile), nil
}
