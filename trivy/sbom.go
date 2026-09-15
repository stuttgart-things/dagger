package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"dagger/trivy/internal/dagger"
)

// digestRef matches a reference pinned to a digest, with or without a tag in
// front of it.
var digestRef = regexp.MustCompile(`@sha256:[0-9a-f]{64}$`)

// sbomFiles maps each supported format to the name of the file it is written
// to, so an exported document says what it is.
var sbomFiles = map[string]string{
	"cyclonedx": "sbom.cdx.json",
	"spdx-json": "sbom.spdx.json",
}

// Sbom writes a software bill of materials for one platform of a published
// image -- the document a signed attestation on that image's digest carries.
//
// The result covers exactly one platform, linux/amd64 unless told otherwise.
// A multi-arch release is an index, and an SBOM describes an image, not an
// index: left to itself trivy resolves the index to the platform it happens to
// run on, so the same call would describe a different image on a different
// runner. For another platform, call again. A platform the index does not
// carry is an error.
//
// imageRef has to be pinned to a digest (repo@sha256:...). An inventory of
// whatever a tag pointed at while this ran is an inventory of nothing in
// particular, and could not be attached to the digest it claims to describe.
// Resolve a tag first, e.g. with the crane module's Digest.
//
// Unlike ScanImage, trivy's exit code is not discarded. An SBOM run that cannot
// pull the image must not hand back an empty document: attested and signed, an
// empty inventory is a signed claim that the image contains nothing.
//
// The image is always read from the registry (--image-src remote), never from
// a local daemon, so the document describes what was published.
func (m *Trivy) Sbom(
	ctx context.Context,
	imageRef string, // Image reference pinned to a digest (e.g., "ghcr.io/org/app@sha256:...")
	// +optional
	registryUser *dagger.Secret,
	// +optional
	registryPassword *dagger.Secret,
	// The one platform the document covers
	// +optional
	// +default="linux/amd64"
	platform string,
	// cyclonedx or spdx-json
	// +optional
	// +default="cyclonedx"
	format string,
	// +optional
	// +default="0.64.1"
	trivyVersion string,
) (*dagger.File, error) {
	if !digestRef.MatchString(imageRef) {
		return nil, fmt.Errorf(
			"%s is not pinned to a digest: an SBOM has to describe one image, "+
				"so pass repo@sha256:... (resolve a tag with the crane module's Digest)",
			imageRef)
	}

	if platform == "" {
		return nil, fmt.Errorf(
			"platform must not be empty: an SBOM describes one platform of %s, not its index",
			imageRef)
	}

	fileName, ok := sbomFiles[format]
	if !ok {
		return nil, fmt.Errorf("unsupported SBOM format %q: use cyclonedx or spdx-json", format)
	}
	sbomPath := "/tmp/" + fileName

	trivyContainer := dag.Container().
		From("aquasec/trivy:" + trivyVersion)

	if registryUser != nil && registryPassword != nil { // pragma: allowlist secret
		trivyContainer = trivyContainer.
			WithSecretVariable("TRIVY_USERNAME", registryUser).
			WithSecretVariable("TRIVY_PASSWORD", registryPassword)

		fmt.Println("✅ Trivy credentials configured")
	}

	sbom, err := trivyContainer.
		WithExec([]string{
			"trivy", "image",
			"--quiet",
			"--image-src", "remote",
			"--platform", platform,
			"--format", format,
			"--output", sbomPath,
			imageRef,
		}).
		File(sbomPath).
		Sync(ctx)
	if err != nil {
		var execErr *dagger.ExecError
		if errors.As(err, &execErr) {
			return nil, fmt.Errorf("trivy could not write an SBOM for %s (%s): %s",
				imageRef, platform, strings.TrimSpace(execErr.Stderr))
		}
		return nil, fmt.Errorf("trivy could not write an SBOM for %s (%s): %w",
			imageRef, platform, err)
	}

	return sbom, nil
}
