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

func (m *Sops) container(ctx context.Context) (*dagger.Container, error) {
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	arch, err := engineArch(ctx)
	if err != nil {
		return nil, err
	}

	sopsBin := fmt.Sprintf("sops-v%s.linux.%s", sopsVersion, arch)
	sopsURL := fmt.Sprintf("https://github.com/getsops/sops/releases/download/v%s/", sopsVersion)
	checksums := fmt.Sprintf("sops-v%s.checksums.txt", sopsVersion)

	ageTarball := fmt.Sprintf("age-v%s-linux-%s.tar.gz", ageVersion, arch)
	ageURL := fmt.Sprintf("https://github.com/FiloSottile/age/releases/download/v%s/%s", ageVersion, ageTarball)

	ctr := dag.Container().
		From(m.BaseImage).
		WithMountedFile("/tmp/dl/"+sopsBin, dag.HTTP(sopsURL+sopsBin)).
		WithMountedFile("/tmp/dl/"+checksums, dag.HTTP(sopsURL+checksums)).
		WithMountedFile("/tmp/dl/"+ageTarball, dag.HTTP(ageURL)).
		WithWorkdir("/tmp/dl").
		WithExec([]string{"sh", "-c",
			// Verify the sops binary against the release's checksum file.
			`grep " ` + sopsBin + `$" ` + checksums + ` | sha256sum -c - && ` +
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
