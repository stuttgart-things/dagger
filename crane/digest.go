package main

import (
	"context"
	"dagger/crane/internal/dagger"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Digest resolves a reference to the digest it currently points at and returns
// it bare ("sha256:..."), so a caller can write ref@<digest> without parsing.
//
// platform defaults to empty, unlike Copy, where it defaults to linux/amd64.
// That difference is deliberate: a signature on a multi-arch release is made on
// the index digest, and resolving to one platform's manifest returns a digest
// that nothing signed. Set platform only when one platform's manifest is what
// you are after. For a reference that is a single manifest rather than an
// index, crane does not check platform at all and returns that manifest's
// digest whatever platform was asked for.
//
// A reference that does not exist is an error carrying the registry's message
// (e.g. MANIFEST_UNKNOWN), never an empty string a caller would concatenate
// into "repo@".
//
// The answer is never taken from cache: a tag moves, and the digest it pointed
// at on an earlier run is not necessarily the one it points at now.
// +cache="never"
func (m *Crane) Digest(
	ctx context.Context,
	ref string,
	// Resolve to this platform's manifest instead of the index. Empty, unlike
	// Copy: a multi-arch release is signed on its index digest.
	// +optional
	platform string,
	// Registry to log in to; derived from ref when empty
	// +optional
	registry string,
	// +optional
	username string,
	// +optional
	password *dagger.Secret,
	// +optional
	// +flag
	// +default=false
	insecure bool,
) (string, error) {
	ctr := m.resolver(registry, username, password, insecure, ref)
	return resolveDigest(ctx, ctr, ref, platform, insecure)
}

// SameDigest checks that two references point at the same digest right now and
// returns that digest. When they do not, the error names both.
//
// This is the check for a moving tag: latest and a release tag can both report
// a successful push while naming different images. platform, the credentials
// and insecure apply to both references, and mean what they mean for Digest.
// +cache="never"
func (m *Crane) SameDigest(
	ctx context.Context,
	refA string,
	refB string,
	// +optional
	platform string,
	// Registry to log in to; derived from each ref when empty
	// +optional
	registry string,
	// +optional
	username string,
	// +optional
	password *dagger.Secret,
	// +optional
	// +flag
	// +default=false
	insecure bool,
) (string, error) {
	ctr := m.resolver(registry, username, password, insecure, refA, refB)

	digestA, err := resolveDigest(ctx, ctr, refA, platform, insecure)
	if err != nil {
		return "", err
	}
	digestB, err := resolveDigest(ctx, ctr, refB, platform, insecure)
	if err != nil {
		return "", err
	}

	if digestA != digestB {
		return "", fmt.Errorf("%s and %s point at different digests:\n  %s  %s\n  %s  %s",
			refA, refB, digestA, refA, digestB, refB)
	}

	return digestA, nil
}

// resolver returns a crane container logged in, with the same credentials, to
// every registry the references live on -- or only to registry, when given.
func (m *Crane) resolver(
	registry string,
	username string,
	password *dagger.Secret,
	insecure bool,
	refs ...string,
) *dagger.Container {
	ctr := m.container(insecure)

	// crane version prints a commit hash, so the image tag is what makes the
	// version behind an answer knowable.
	fmt.Println("Resolving digests with", craneImage(m.Version))

	if username == "" || password == nil { // pragma: allowlist secret
		return ctr
	}

	registries := []string{registry}
	if registry == "" {
		registries = nil
		for _, ref := range refs {
			if r := extractRegistry(ref); r != "" && !slices.Contains(registries, r) {
				registries = append(registries, r)
			}
		}
	}

	for _, r := range registries {
		ctr = authenticate(ctr, &RegistryAuth{
			URL:      r,
			Username: username,
			Password: password,
		}, insecure)
	}

	return ctr
}

// resolveDigest runs crane digest for one reference.
func resolveDigest(
	ctx context.Context,
	ctr *dagger.Container,
	ref string,
	platform string,
	insecure bool,
) (string, error) {
	cmd := []string{"crane", "digest"}
	if platform != "" {
		cmd = append(cmd, "--platform", platform)
	}
	if insecure {
		cmd = append(cmd, "--insecure")
	}
	cmd = append(cmd, ref)

	fmt.Println("Executing command:", strings.Join(cmd, " "))

	out, err := ctr.
		// CACHE_BUSTER forces a fresh lookup: an identical exec would
		// otherwise be answered from cache after the tag has moved.
		WithEnvVariable("CACHE_BUSTER", strconv.FormatInt(time.Now().UnixNano(), 10)).
		WithExec(cmd).
		Stdout(ctx)
	if err != nil {
		var execErr *dagger.ExecError
		if errors.As(err, &execErr) {
			return "", fmt.Errorf("cannot resolve %s: %s", ref, strings.TrimSpace(execErr.Stderr))
		}
		return "", fmt.Errorf("cannot resolve %s: %w", ref, err)
	}

	digest := strings.TrimSpace(out)
	if !strings.HasPrefix(digest, "sha256:") {
		return "", fmt.Errorf("crane returned %q for %s, not a sha256 digest", digest, ref)
	}

	return digest, nil
}
