// A Dagger module for SOPS encryption and decryption
//
// This module provides functionality for working with [Mozilla SOPS](https://github.com/getsops/sops)
// in a Dagger pipeline. It supports encrypting and decrypting files using an AGE key.
// Files are mounted into a container and processed using the `sops` CLI tool.
//
// The `DecryptSops` function decrypts a file with the given SOPS key and returns its plaintext content.
// The `EncryptSops` function encrypts a plaintext file using the same SOPS key and returns the encrypted content.

package main

type Sops struct {
	BaseImage string
}
