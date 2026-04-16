// Flux module for building and pushing OCI artifacts
//
// This module provides functionality for packaging directories into OCI
// artifacts and pushing them to container registries using the Flux CLI.
// It is designed for Flux GitOps workflows where Kubernetes manifests,
// Kustomize overlays, or other configuration are stored as OCI artifacts.

package main

// Flux builds and pushes OCI artifacts using the Flux CLI
type Flux struct {
	// Base image to use for the Flux container
	// +optional
	// +default="cgr.dev/chainguard/wolfi-base:latest"
	BaseImage string
}
