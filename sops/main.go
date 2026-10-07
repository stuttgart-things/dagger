// A Dagger module for SOPS encryption and decryption
//
// This module provides functionality for working with [Mozilla SOPS](https://github.com/getsops/sops)
// in a Dagger pipeline. It supports generating AGE keys, encrypting and decrypting files.
// Files are mounted into a container and processed using the `sops` CLI tool.
//
// Functions:
//   - GenerateAgeKey: Generates a new AGE key pair
//   - GenerateSopsConfig: Generates a .sops.yaml configuration file
//   - Encrypt: Encrypts a plaintext file using SOPS with an AGE key
//   - Decrypt: Decrypts a SOPS-encrypted file and returns the decrypted file

package main

import "dagger/sops/internal/dagger"

// sopsConfigPath is where a caller's .sops.yaml is mounted. sops looks for
// .sops.yaml from the working directory upwards, so the old mount at
// /root/.sops.yaml was never found and --sops-config was ignored (#404). It
// is also passed explicitly with --config.
const sopsConfigPath = "/src/.sops.yaml"

// withSopsConfig mounts the optional config and returns the sops command
// with the matching --config flag, followed by args.
func withSopsConfig(ctr **dagger.Container, sopsConfig *dagger.File, args ...string) []string {
	cmd := []string{"sops"}
	if sopsConfig != nil {
		*ctr = (*ctr).WithMountedFile(sopsConfigPath, sopsConfig)
		cmd = append(cmd, "--config", sopsConfigPath)
	}
	return append(cmd, args...)
}

type Sops struct {
	BaseImage string
}
