// Crane module for cross-registry image transfers
//
// This module provides functionality for copying container images between
// registries using Google’s `crane` CLI, wrapped in a Dagger pipeline.
// It supports authentication, platform targeting, and insecure registry access.
//
// The module is ideal for scenarios where images need to be promoted between
// environments (e.g., dev → staging → production) or mirrored across
// different registry backends.
//
// Typical usage includes:
//   - Copying an image from one registry to another (e.g., Harbor to GHCR)
//   - Providing credentials for source and/or target registries
//   - Optionally specifying platform (e.g., "linux/amd64")
//   - Allowing insecure registries in air-gapped or self-hosted setups
//
// This module is designed to be used as part of a CI/CD pipeline via the
// Dagger CLI or SDKs.

package main

import (
	"dagger/crane/internal/dagger"
)

// Crane installs Crane CLI on a Wolfi base image at runtime
// @module
type Crane struct {
	// Base Wolfi image to use
	// +optional
	// +default="cgr.dev/chainguard/wolfi-base:latest"
	BaseImage string
	// Crane version to install (e.g., "latest" or specific version)
	// +optional
	// +default="latest"
	Version string
}

// RegistryAuth contains authentication details for a registry
type RegistryAuth struct {
	URL      string
	Username string
	Password *dagger.Secret
}

// container returns a Wolfi-based container with Crane CLI installed
func (m *Crane) container(insecure bool) *dagger.Container {
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	ctr := dag.Container().From(m.BaseImage)

	pkg := "crane"
	ctr = ctr.WithExec([]string{"apk", "add", "--no-cache", pkg})
	ctr = ctr.WithEntrypoint([]string{"crane"})

	if insecure {
		ctr = ctr.WithEnvVariable("SSL_CERT_DIR", "/nonexistent")
	}

	return ctr
}
