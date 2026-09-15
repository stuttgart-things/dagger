package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"dagger/cosign/internal/dagger"
)

// Sign signs the artefact at ref keylessly: cosign gets a short-lived
// certificate from Fulcio for the identity in identityToken, records the
// signature in Rekor and pushes it next to the artefact. It returns cosign's
// report, which names the Rekor entry and where the signature went.
//
// ref has to be pinned to a digest (repo@sha256:...). A signature on a tag is a
// signature on a name somebody can move, so a tag is refused rather than
// resolved here. Resolve it first with the crane module's Digest; for a
// multi-arch release that is the index digest.
//
// identityToken is an OIDC token with audience "sigstore", requested by the
// workflow that calls this function. In GitHub Actions the job needs
// `id-token: write`, and Fulcio writes the calling workflow's job_workflow_ref
// into the certificate, which is what Verify matches. Requested from a shared
// template's workflow instead, every repository using that template would sign
// as one and the same identity.
//
// The token reaches cosign as SIGSTORE_ID_TOKEN, never as an argument, and the
// credentials that request a token (ACTIONS_ID_TOKEN_REQUEST_*) never enter the
// container. Without such a token nothing here works locally; Verify does.
//
// The registry credentials need push access: the signature is stored in the
// artefact's repository.
// +cache="never"
func (m *Cosign) Sign(
	ctx context.Context,
	// Artefact to sign, pinned to a digest (repo@sha256:...)
	ref string,
	// OIDC token with audience "sigstore" for the identity to sign as
	identityToken *dagger.Secret,
	// Registry to log in to; derived from ref when empty
	// +optional
	registry string,
	// +optional
	registryUsername string,
	// +optional
	registryPassword *dagger.Secret,
	// cosign release
	// +optional
	// +default="2.6.5"
	cosignVersion string,
) (string, error) {
	if err := requireDigest(ref); err != nil {
		return "", err
	}

	ctr := m.container(cosignVersion, registryAuth{registry, registryUsername, registryPassword}, ref).
		WithSecretVariable("SIGSTORE_ID_TOKEN", identityToken)

	_, stderr, err := run(ctx, ctr, "sign", "--yes", ref)
	if err != nil {
		return "", cosignError("cannot sign "+ref, err)
	}

	return strings.TrimSpace(stderr), nil
}

// Attest attaches predicate to the artefact at ref as a signed in-toto
// attestation of predicateType, e.g. a CycloneDX SBOM from the trivy module's
// Sbom as "cyclonedx". The token, the registry credentials and the digest
// requirement work as they do for Sign.
//
// predicateType is one of cosign's names (cyclonedx, spdxjson, slsaprovenance,
// vuln, openvex, custom, ...) or a predicate type URI. It has no default: the
// type is a claim about what the predicate is.
//
// An empty predicate is refused. Attested and signed, an empty SBOM is a signed
// claim that the image contains nothing.
//
// cosign adds an attestation next to those already there. VerifyAttestation
// refuses to choose between attestations of one type that carry different
// predicates, so a job that may attest the same digest again, such as a
// re-run, should pass replace: it drops the earlier attestations of the same
// type before adding this one.
// +cache="never"
func (m *Cosign) Attest(
	ctx context.Context,
	// Artefact to attest, pinned to a digest (repo@sha256:...)
	ref string,
	// Predicate document, e.g. a CycloneDX SBOM
	predicate *dagger.File,
	// cosign predicate type name (cyclonedx, spdxjson, slsaprovenance, ...) or URI
	predicateType string,
	// OIDC token with audience "sigstore" for the identity to sign as
	identityToken *dagger.Secret,
	// Replace earlier attestations of the same type on ref
	// +optional
	// +default=false
	replace bool,
	// Registry to log in to; derived from ref when empty
	// +optional
	registry string,
	// +optional
	registryUsername string,
	// +optional
	registryPassword *dagger.Secret,
	// cosign release
	// +optional
	// +default="2.6.5"
	cosignVersion string,
) (string, error) {
	if err := requireDigest(ref); err != nil {
		return "", err
	}
	if predicateType == "" {
		return "", errors.New("predicateType must not be empty: it is a claim about what the predicate is")
	}

	size, err := predicate.Size(ctx)
	if err != nil {
		return "", fmt.Errorf("cannot read the predicate for %s: %w", ref, err)
	}
	if size == 0 {
		return "", fmt.Errorf(
			"the predicate for %s is empty: attested and signed, it would be a "+
				"signed claim about nothing",
			ref)
	}

	ctr := m.container(cosignVersion, registryAuth{registry, registryUsername, registryPassword}, ref).
		WithSecretVariable("SIGSTORE_ID_TOKEN", identityToken).
		WithMountedFile("/work/predicate", predicate)

	args := []string{"attest", "--yes", "--type", predicateType, "--predicate", "/work/predicate"}
	if replace {
		args = append(args, "--replace")
	}

	_, stderr, err := run(ctx, ctr, append(args, ref)...)
	if err != nil {
		return "", cosignError(fmt.Sprintf("cannot attest %s to %s", predicateType, ref), err)
	}

	return strings.TrimSpace(stderr), nil
}
