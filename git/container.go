package main

import (
	"context"
	"fmt"

	"dagger/git/internal/dagger"
)

func (m *Git) container(
	ctx context.Context) (*dagger.Container, error) {
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	ghVersion := "2.82.1"
	ghURL := "https://github.com/cli/cli/releases/download/v" + ghVersion + "/gh_" + ghVersion + "_linux_amd64.tar.gz"

	// Combine multiple shell commands into a single script for efficiency
	// This reduces container layer overhead and speeds up container creation
	installGhScript := fmt.Sprintf(`
wget -O /tmp/gh.tar.gz %s && \
tar -xzf /tmp/gh.tar.gz -C /tmp && \
mkdir -p /usr/local/bin && \
cp /tmp/gh_%s_linux_amd64/bin/gh /usr/local/bin/gh && \
chmod +x /usr/local/bin/gh && \
rm -rf /tmp/gh.tar.gz /tmp/gh_%s_linux_amd64
`, ghURL, ghVersion, ghVersion)

	ctr := dag.Container().
		From(m.BaseImage).
		WithExec([]string{"apk", "add", "--no-cache", "git", "wget"}).
		WithExec([]string{"sh", "-c", installGhScript}).
		WithEntrypoint([]string{"git"})

	return ctr, nil
}
