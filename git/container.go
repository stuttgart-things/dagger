package main

import (
	"context"

	"dagger/git/internal/dagger"
)

func (m *Git) container(
	ctx context.Context) (*dagger.Container, error) {
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	ghVersion := "2.82.1"
	ghURL := "https://github.com/cli/cli/releases/download/v" + ghVersion + "/gh_" + ghVersion + "_linux_amd64.tar.gz"

	ctr := dag.Container().
		From(m.BaseImage).
		WithExec([]string{"apk", "add", "--no-cache", "git", "wget"}).
		WithExec([]string{"sh", "-c", "wget -O /tmp/gh.tar.gz " + ghURL}).
		WithExec([]string{"sh", "-c", "tar -xzf /tmp/gh.tar.gz -C /tmp"}).
		WithExec([]string{"sh", "-c", "mkdir -p /usr/local/bin"}).
		WithExec([]string{"sh", "-c", "cp /tmp/gh_" + ghVersion + "_linux_amd64/bin/gh /usr/local/bin/gh"}).
		WithExec([]string{"sh", "-c", "chmod +x /usr/local/bin/gh"}).
		WithExec([]string{"sh", "-c", "rm -rf /tmp/gh.tar.gz /tmp/gh_" + ghVersion + "_linux_amd64"}).
		WithEntrypoint([]string{"git"})

	return ctr, nil
}
