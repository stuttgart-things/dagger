package main

import (
	"context"
	"dagger/techdocs/internal/dagger"
	"fmt"
	"strings"
)

// Pinned tool versions.
// renovate: datasource=pypi depName=mkdocs-techdocs-core
const techdocsCoreVersion = "1.7.1"

// renovate: datasource=npm depName=@techdocs/cli
const techdocsCliVersion = "1.13.0"

// Node.js is downloaded from nodejs.org and checked against the release's
// SHASUMS256.txt instead of installed with apt, so the version does not
// depend on the day of the run. The tarball is .tar.gz: the slim python
// image has no xz.
// renovate: datasource=node-version depName=node
const nodeVersion = "22.23.3"

// The base image, pinned by digest. It brings python3 + pip for mkdocs;
// node comes from nodejs.org above. Renovate moves the digest.
// renovate: datasource=docker depName=python
const baseImageTag = "3.13-slim@sha256:70729b46c69b4f1e97c4822c1af3df53a1476cf5ddc6c087c0c10bc3a5678c2f"

const defaultBaseImage = "python:" + baseImageTag

func (m *Techdocs) container(ctx context.Context) (*dagger.Container, error) {
	if m.BaseImage == "" {
		m.BaseImage = defaultBaseImage
	}

	arch, err := nodeArch(ctx)
	if err != nil {
		return nil, err
	}

	nodeURL := fmt.Sprintf("https://nodejs.org/dist/v%s/", nodeVersion)
	tarball := fmt.Sprintf("node-v%s-linux-%s.tar.gz", nodeVersion, arch)

	return dag.Container().
		From(m.BaseImage).
		WithMountedFile("/tmp/dl/"+tarball, dag.HTTP(nodeURL+tarball)).
		WithMountedFile("/tmp/dl/SHASUMS256.txt", dag.HTTP(nodeURL+"SHASUMS256.txt")).
		WithWorkdir("/tmp/dl").
		WithExec([]string{"sh", "-c",
			`grep " ` + tarball + `$" SHASUMS256.txt | sha256sum -c - && ` +
				`tar -xzf ` + tarball + ` -C /usr/local --strip-components=1 ` +
				`--exclude='*/CHANGELOG.md' --exclude='*/README.md' --exclude='*/LICENSE'`,
		}).
		WithWorkdir("/").
		WithExec([]string{"pip", "install", "--no-cache-dir", "--root-user-action=ignore",
			"mkdocs-techdocs-core==" + techdocsCoreVersion}).
		WithExec([]string{"npm", "install", "--global", "--no-fund", "--no-audit",
			"@techdocs/cli@" + techdocsCliVersion}), nil
}

// nodeArch maps the engine's platform (e.g. linux/arm64) to the architecture
// name used in the Node.js release assets.
func nodeArch(ctx context.Context) (string, error) {
	platform, err := dag.DefaultPlatform(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read engine platform: %w", err)
	}

	switch arch := strings.Split(strings.TrimPrefix(string(platform), "linux/"), "/")[0]; arch {
	case "amd64":
		return "x64", nil
	case "arm64":
		return "arm64", nil
	default:
		return "", fmt.Errorf("unsupported platform %q: node is downloaded for linux/amd64 or linux/arm64", platform)
	}
}
