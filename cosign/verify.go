package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"dagger/cosign/internal/dagger"
)

// Verify checks the signatures on ref against an OIDC issuer and exactly one of
// an identity or an identity regexp, and returns the verified payloads as
// cosign prints them (JSON). It is an error when no signature verifies.
//
// There is no default identity. For GitHub Actions the issuer is
// https://token.actions.githubusercontent.com and the identity is the signing
// workflow, e.g.
// ^https://github\.com/<org>/<repo>/\.github/workflows/<file>\.yaml@refs/
// Anchor and escape a regexp, or it matches more repositories than yours.
//
// No token and no local cosign install are needed, so this is also how to check
// an artefact by hand. A tag is accepted here and means whatever it points at
// right now.
//
// A verification that has never refused anything cannot be told apart from one
// that cannot refuse. Pair it with VerifyRefuses.
// +cache="never"
func (m *Cosign) Verify(
	ctx context.Context,
	// Artefact to verify, ideally pinned to a digest
	ref string,
	// OIDC issuer the certificate must name
	certificateOidcIssuer string,
	// Exact identity (certificate SAN) the signature must carry
	// +optional
	certificateIdentity string,
	// Regexp the identity must match; anchor and escape it
	// +optional
	certificateIdentityRegexp string,
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
	flags, err := identityFlags(certificateOidcIssuer, certificateIdentity, certificateIdentityRegexp)
	if err != nil {
		return "", err
	}

	ctr := m.container(cosignVersion, registryAuth{registry, registryUsername, registryPassword}, ref)

	stdout, stderr, err := run(ctx, ctr, append(append([]string{"verify"}, flags...), ref)...)
	if err != nil {
		return "", cosignError("cannot verify "+ref, err)
	}

	fmt.Println(strings.TrimSpace(stderr))
	return stdout, nil
}

// VerifyAttestation checks the attestations of predicateType on ref against an
// OIDC issuer and an identity, as Verify does for signatures, and returns the
// predicate they carry, e.g. the CycloneDX document of a "cyclonedx"
// attestation.
//
// When several attestations of that type verify and carry different
// predicates -- a digest attested twice without replace -- this is an error
// rather than a pick of one of them.
// +cache="never"
func (m *Cosign) VerifyAttestation(
	ctx context.Context,
	// Artefact to verify, ideally pinned to a digest
	ref string,
	// cosign predicate type name (cyclonedx, spdxjson, slsaprovenance, ...) or URI
	predicateType string,
	// OIDC issuer the certificate must name
	certificateOidcIssuer string,
	// Exact identity (certificate SAN) the attestation must carry
	// +optional
	certificateIdentity string,
	// Regexp the identity must match; anchor and escape it
	// +optional
	certificateIdentityRegexp string,
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
) (*dagger.File, error) {
	if predicateType == "" {
		return nil, errors.New("predicateType must not be empty")
	}
	flags, err := identityFlags(certificateOidcIssuer, certificateIdentity, certificateIdentityRegexp)
	if err != nil {
		return nil, err
	}

	ctr := m.container(cosignVersion, registryAuth{registry, registryUsername, registryPassword}, ref)

	args := append([]string{"verify-attestation", "--type", predicateType}, flags...)
	stdout, stderr, err := run(ctx, ctr, append(args, ref)...)
	if err != nil {
		return nil, cosignError(fmt.Sprintf("cannot verify a %s attestation on %s", predicateType, ref), err)
	}
	fmt.Println(strings.TrimSpace(stderr))

	predicate, err := singlePredicate(stdout)
	if err != nil {
		return nil, fmt.Errorf("%s attestation on %s: %w", predicateType, ref, err)
	}

	return dag.Directory().
		WithNewFile("predicate.json", string(predicate)).
		File("predicate.json"), nil
}

// VerifyRefuses proves that verification can refuse. It runs the check Verify
// runs (VerifyAttestation's, when predicateType is set) with an identity that
// must NOT match, and succeeds only when cosign refuses it for that reason.
//
// It is an error when the verification passes: the identity is broader than it
// looks, or the check is not a check. It is also an error when cosign fails for
// any other reason -- an artefact with no signature at all, a registry that
// cannot be reached, a typo in the issuer -- because none of those shows that a
// signature which does exist would be refused.
//
// Use an identity that differs from the real one in the part that matters,
// such as another repository of the same organisation, with the same issuer.
// +cache="never"
func (m *Cosign) VerifyRefuses(
	ctx context.Context,
	// Artefact that is signed, by an identity other than the one below
	ref string,
	// OIDC issuer the real signature was made with
	certificateOidcIssuer string,
	// Exact identity that must not match
	// +optional
	certificateIdentity string,
	// Identity regexp that must not match
	// +optional
	certificateIdentityRegexp string,
	// Check attestations of this type instead of signatures
	// +optional
	predicateType string,
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
	flags, err := identityFlags(certificateOidcIssuer, certificateIdentity, certificateIdentityRegexp)
	if err != nil {
		return "", err
	}

	what := "signature"
	args := []string{"verify"}
	if predicateType != "" {
		what = predicateType + " attestation"
		args = []string{"verify-attestation", "--type", predicateType}
	}
	args = append(append(args, flags...), ref)

	ctr := m.container(cosignVersion, registryAuth{registry, registryUsername, registryPassword}, ref)

	_, _, err = run(ctx, ctr, args...)
	if err == nil {
		return "", fmt.Errorf(
			"the %s on %s verified for %s, an identity that must not match: "+
				"a verification that accepts this is not a gate",
			what, ref, strings.Join(flags[2:], " "))
	}

	var execErr *dagger.ExecError
	if !errors.As(err, &execErr) {
		return "", fmt.Errorf("cannot run cosign against %s: %w", ref, err)
	}

	failure := cosignFailure(execErr.Stderr)
	if !strings.Contains(failure, identityMismatch) {
		return "", fmt.Errorf(
			"cosign failed on the %s on %s, but not because the identity did not "+
				"match, so this proves no refusal: %s",
			what, ref, failure)
	}

	return fmt.Sprintf("refused, as it must: %s", failure), nil
}
