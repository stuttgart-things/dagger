// Cosign module for keyless signing, attestation and verification
//
// This module signs container images and OCI artefacts by digest with
// Sigstore's cosign, attaches predicates such as a CycloneDX SBOM as signed
// attestations, and verifies both against an OIDC issuer and an identity.
//
// Signing is keyless: cosign gets a short-lived certificate from Fulcio for the
// identity in an OIDC token, and the signature is recorded in Rekor. The token
// is requested by the calling repository's own workflow and handed in as a
// secret, so the identity in the certificate is that repository's workflow and
// not a shared template's.
//
// Verification needs no token and no local cosign install, so anyone can check
// an artefact by hand.

package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"dagger/cosign/internal/dagger"
)

// Cosign signs, attests and verifies OCI artefacts with a pinned cosign release
type Cosign struct {
	// Base image the cosign binary is copied into. It needs a shell, which
	// the registry login runs in.
	// +optional
	// +default="cgr.dev/chainguard/wolfi-base:latest"
	BaseImage string
}

// defaultCosignVersion is the cosign release used when cosignVersion is empty.
// Keep it in step with the +default annotations, which have to be literals.
//
// It stays on 2.x on purpose. cosign 3 is a major release, and nobody has
// checked yet that Kyverno's verifier reads what it writes by default.
const defaultCosignVersion = "2.6.5"

// identityMismatch is how cosign says that a signature or attestation exists
// but was made by an identity (or issuer) other than the one asked for.
const identityMismatch = "none of the expected identities matched"

// digestRef matches a reference pinned to a digest, with or without a tag in
// front of it. A tag does not end up in what is signed: cosign's payload names
// the repository and the digest only.
var digestRef = regexp.MustCompile(`@sha256:[0-9a-f]{64}$`)

// cosignScript logs in to the registry when credentials are set, then runs
// cosign with the arguments it was given. Both happen in one exec, with the
// docker config on a temporary mount, so the credentials never become part of
// a cached layer.
const cosignScript = `set -e
if [ -n "${REGISTRY_USERNAME:-}" ]; then
  printenv REGISTRY_PASSWORD \
    | cosign login "${REGISTRY}" --username "${REGISTRY_USERNAME}" --password-stdin >/dev/null
fi
exec cosign "$@"`

// cosignImage returns the image of a cosign release. Its tags are v-prefixed;
// a bare "2.6.5" is accepted as well.
func cosignImage(version string) string {
	if version == "" {
		version = defaultCosignVersion
	}
	if version[0] >= '0' && version[0] <= '9' {
		version = "v" + version
	}
	return "gcr.io/projectsigstore/cosign:" + version
}

// registryAuth is what cosign needs to read or write the signature and
// attestation images stored next to an artefact.
type registryAuth struct {
	// registry to log in to; derived from the reference when empty
	registry string
	username string
	password *dagger.Secret
}

// container returns BaseImage with the cosign binary of the given release,
// set up to log in to ref's registry when auth carries credentials.
//
// The binary is copied out of cosign's release image, so the version does not
// depend on the day the container is built.
func (m *Cosign) container(cosignVersion string, auth registryAuth, ref string) *dagger.Container {
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	image := cosignImage(cosignVersion)
	fmt.Println("Using", image)

	cli := dag.Container().From(image).File("/ko-app/cosign")

	ctr := dag.Container().From(m.BaseImage).
		WithFile("/usr/bin/cosign", cli, dagger.ContainerWithFileOpts{Permissions: 0o755}).
		WithMountedTemp("/cosign-auth").
		WithEnvVariable("DOCKER_CONFIG", "/cosign-auth")

	if auth.username == "" || auth.password == nil { // pragma: allowlist secret
		return ctr
	}

	registry := auth.registry
	if registry == "" {
		registry = registryOf(ref)
	}
	fmt.Printf("Authenticating with registry: %s as user: %s\n", registry, auth.username)

	return ctr.
		WithEnvVariable("REGISTRY", registry).
		WithEnvVariable("REGISTRY_USERNAME", auth.username).
		WithSecretVariable("REGISTRY_PASSWORD", auth.password)
}

// registryOf returns the registry host of a reference, and index.docker.io for
// one without a host, which is how docker and cosign read it.
func registryOf(ref string) string {
	first, _, found := strings.Cut(ref, "/")
	if found && (strings.ContainsAny(first, ".:") || first == "localhost") {
		return first
	}
	return "index.docker.io"
}

// requireDigest refuses a reference that is not pinned to a digest.
func requireDigest(ref string) error {
	if digestRef.MatchString(ref) {
		return nil
	}
	return fmt.Errorf(
		"%s is not pinned to a digest: a signature on a tag is a signature on a "+
			"name somebody can move, so pass repo@sha256:... "+
			"(resolve a tag with the crane module's Digest)",
		ref)
}

// identityFlags returns cosign's certificate flags for an issuer and exactly
// one of an identity or an identity regexp.
//
// There is no default for either. cosign refuses keyless verification without
// an identity because anybody can get a certificate from a public issuer; a
// module default would be an identity nobody chose.
func identityFlags(issuer, identity, identityRegexp string) ([]string, error) {
	if issuer == "" {
		return nil, errors.New("certificateOidcIssuer must not be empty")
	}

	flags := []string{"--certificate-oidc-issuer", issuer}

	switch {
	case identity != "" && identityRegexp != "":
		return nil, errors.New("pass certificateIdentity or certificateIdentityRegexp, not both")
	case identity != "":
		return append(flags, "--certificate-identity", identity), nil
	case identityRegexp != "":
		return append(flags, "--certificate-identity-regexp", identityRegexp), nil
	}

	return nil, fmt.Errorf(
		"pass certificateIdentity or certificateIdentityRegexp: without one, "+
			"any certificate %s ever issued would verify",
		issuer)
}

// run executes cosign with args in ctr and returns its stdout and stderr.
//
// Every run is fresh (CACHE_BUSTER). Signing writes to a registry and a
// transparency log, and a verification is only as current as the registry
// state it read, so an identical earlier run answers nothing.
func run(ctx context.Context, ctr *dagger.Container, args ...string) (string, string, error) {
	fmt.Println("Executing command: cosign", strings.Join(args, " "))

	res := ctr.
		WithEnvVariable("CACHE_BUSTER", strconv.FormatInt(time.Now().UnixNano(), 10)).
		WithExec(append([]string{"sh", "-c", cosignScript, "cosign"}, args...))

	stdout, err := res.Stdout(ctx)
	if err != nil {
		return "", "", err
	}
	stderr, err := res.Stderr(ctx)
	if err != nil {
		return "", "", err
	}

	return stdout, stderr, nil
}

// cosignError wraps a failed run in an error that says what cosign said.
func cosignError(what string, err error) error {
	var execErr *dagger.ExecError
	if errors.As(err, &execErr) {
		return fmt.Errorf("%s: %s", what, cosignFailure(execErr.Stderr))
	}
	return fmt.Errorf("%s: %w", what, err)
}

// cosignFailure picks cosign's own error line out of its stderr, which also
// carries progress messages and repeats the error once more at the end.
func cosignFailure(stderr string) string {
	for _, line := range strings.Split(stderr, "\n") {
		if msg, ok := strings.CutPrefix(line, "Error: "); ok {
			return msg
		}
	}
	return strings.TrimSpace(stderr)
}

// Version returns `cosign version` for the release the other functions use.
func (m *Cosign) Version(
	ctx context.Context,
	// cosign release
	// +optional
	// +default="2.6.5"
	cosignVersion string,
) (string, error) {
	stdout, _, err := run(ctx, m.container(cosignVersion, registryAuth{}, ""), "version")
	if err != nil {
		return "", cosignError("cosign version failed", err)
	}
	return stdout, nil
}
