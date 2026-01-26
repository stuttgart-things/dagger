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

type Sops struct {
	BaseImage string
}
