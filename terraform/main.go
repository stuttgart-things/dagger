// Terraform is a Dagger module that runs Terraform commands inside a containerized environment.
//
// It uses a custom container built on the Wolfi base image with Terraform, SOPS, Git, and supporting tools pre-installed.
// The module supports securely decrypting a SOPS-encrypted `terraform.tfvars.json` using an optional AGE key,
// and mounts it into the working directory for use during plan or apply operations.
//
// Supported Terraform operations include `init`, `apply`, and `destroy`. The module always runs `terraform init` first,
// and then executes the specified operation. After execution, any sensitive files such as the decrypted tfvars file are deleted.
//
// The module also exposes helper methods:
//   - `Version`: returns the installed Terraform version.
//   - `Output`: retrieves Terraform outputs in JSON format.
//   - `DecryptSops`: decrypts a file using SOPS and returns its content as a string.
//
// It is designed to run Terraform commands reproducibly in CI pipelines or local development environments with secret handling and plugin caching.

package main

type Terraform struct {
	BaseImage string
}
