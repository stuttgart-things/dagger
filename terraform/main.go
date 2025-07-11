// Terraform module for Dagger-based infrastructure automation
//
// This module provides functionality to run Terraform operations inside a
// containerized Dagger pipeline. It supports applying, destroying, and
// initializing Terraform configurations with optional variable injection.
//
// The module executes Terraform in a controlled container environment,
// optionally injecting HCL-style variables and secret JSON variables. It
// is especially useful in CI/CD scenarios where reproducibility and
// secure secret handling are required.
//
// Supported features include:
//   - Running `terraform init`, `apply`, or `destroy` inside a Dagger container
//   - Supplying key-value `-var` arguments via a comma-separated string
//   - Injecting a `terraform.tfvars.json` file via a mounted Dagger secret
//   - Fetching outputs from the Terraform state in JSON format
//
// This module is intended to encapsulate Terraform workflows inside a
// container runtime, simplifying automation across cloud or local environments.

package main

type Terraform struct {
	BaseImage string
}
