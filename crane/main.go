// Crane module for cross-registry image transfers
//
// This module provides functionality for copying container images between
// registries, and for resolving a tag to the digest it currently points at,
// using Google’s `crane` CLI, wrapped in a Dagger pipeline.
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
	// Crane version to install, as a go-containerregistry release (e.g.,
	// "0.22.1") or "latest". The binary comes from that release's image, so
	// the version does not depend on the day the container is built.
	// +optional
	// +default="0.22.1"
	Version string
}

// defaultVersion is the crane release used when Version is empty. Keep it in
// step with the +default above, which has to be a literal.
const defaultVersion = "0.22.1"

// craneImage returns the go-containerregistry image of a crane release. Its
// tags are v-prefixed; a bare "0.22.1" is accepted as well.
func craneImage(version string) string {
	if version == "" {
		version = defaultVersion
	}
	if version[0] >= '0' && version[0] <= '9' {
		version = "v" + version
	}
	return "gcr.io/go-containerregistry/crane:" + version
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

	// Copied out of the release image rather than installed with apk, which
	// took whatever crane Wolfi shipped that day and ignored Version.
	crane := dag.Container().From(craneImage(m.Version)).File("/ko-app/crane")

	ctr := dag.Container().From(m.BaseImage).
		WithFile("/usr/bin/crane", crane, dagger.ContainerWithFileOpts{Permissions: 0o755}).
		WithEntrypoint([]string{"crane"})

	if insecure {
		ctr = ctr.WithEnvVariable("SSL_CERT_DIR", "/nonexistent")
	}

	return ctr
}
