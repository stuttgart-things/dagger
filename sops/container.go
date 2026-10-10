package main

import (
	"context"
	"dagger/sops/internal/dagger"
	"fmt"
	"strings"
)

// Pinned tool versions. The release binaries are downloaded instead of
// `apk add sops`, whose version depended on the day of the run (#404).
// renovate: datasource=github-releases depName=getsops/sops extractVersion=^v(?<version>.*)$
const sopsVersion = "3.13.3"

// renovate: datasource=github-releases depName=FiloSottile/age extractVersion=^v(?<version>.*)$
const ageVersion = "1.3.2"

// sha256 of the age release tarballs for ageVersion. age publishes Sigsum
// proofs (.proof) rather than a checksum file, so these were taken from the
// GitHub release on 2026-10-07 and pinned. A Renovate bump of ageVersion
// leaves them stale on purpose: the download then fails the check, and the
// new sums are added by hand in the same PR.
var ageSHA256 = map[string]string{
	"amd64": "cbe24006683f8eb669266162894b9a522a1af52f2665fbc63a4bb032ed26ac10",
	"arm64": "6b8dc4333c53a5a57c9e5834e3a48f92605d7154014cd07269ff3327db5d37f4",
}

// The base image, pinned by digest: only sh, grep, sha256sum, tar and install
// are used from it, but "latest" alone changes underneath every run. Renovate
// moves the digest.
// renovate: datasource=docker depName=cgr.dev/chainguard/wolfi-base
const baseImageTag = "latest@sha256:1c451d46a0d5c4e9f2b38e0e8d3e299564a1aa95c21973efcc1980a9d1d2e73e"

const defaultBaseImage = "cgr.dev/chainguard/wolfi-base:" + baseImageTag

func (m *Sops) container(ctx context.Context) (*dagger.Container, error) {
	if m.BaseImage == "" {
		m.BaseImage = defaultBaseImage
	}

	arch, err := engineArch(ctx)
	if err != nil {
		return nil, err
	}

	sopsBin := fmt.Sprintf("sops-v%s.linux.%s", sopsVersion, arch)
	sopsURL := fmt.Sprintf("https://github.com/getsops/sops/releases/download/v%s/", sopsVersion)
	checksums := fmt.Sprintf("sops-v%s.checksums.txt", sopsVersion)

	ageTarball := fmt.Sprintf("age-v%s-linux-%s.tar.gz", ageVersion, arch)
	ageSum, ok := ageSHA256[arch]
	if !ok {
		return nil, fmt.Errorf("no pinned age checksum for %s", arch)
	}
	ageURL := fmt.Sprintf("https://github.com/FiloSottile/age/releases/download/v%s/%s", ageVersion, ageTarball)

	ctr := dag.Container().
		From(m.BaseImage).
		WithMountedFile("/tmp/dl/"+sopsBin, dag.HTTP(sopsURL+sopsBin)).
		WithMountedFile("/tmp/dl/"+checksums, dag.HTTP(sopsURL+checksums)).
		WithMountedFile("/tmp/dl/"+ageTarball, dag.HTTP(ageURL)).
		WithWorkdir("/tmp/dl").
		WithExec([]string{"sh", "-c",
			// Verify sops against the release's checksum file, age against
			// the pinned sum.
			`grep " ` + sopsBin + `$" ` + checksums + ` | sha256sum -c - && ` +
				`echo "` + ageSum + `  ` + ageTarball + `" | sha256sum -c - && ` +
				`install -D -m 0755 ` + sopsBin + ` /usr/local/bin/sops && ` +
				`tar -xzf ` + ageTarball + ` -C /tmp && ` +
				`install -m 0755 /tmp/age/age /tmp/age/age-keygen /usr/local/bin/`,
		}).
		WithWorkdir("/").
		WithEntrypoint([]string{"sops"})

	return ctr, nil
}

// engineArch maps the engine's platform (e.g. linux/arm64) to the
// architecture name used in the sops and age release assets.
func engineArch(ctx context.Context) (string, error) {
	platform, err := dag.DefaultPlatform(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read engine platform: %w", err)
	}

	arch := strings.Split(strings.TrimPrefix(string(platform), "linux/"), "/")[0]
	switch arch {
	case "amd64", "arm64":
		return arch, nil
	default:
		return "", fmt.Errorf("unsupported platform %q: sops and age are downloaded for linux/amd64 or linux/arm64", platform)
	}
}
