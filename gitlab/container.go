package main

import (
	"context"

	"dagger/gitlab/internal/dagger"
)

func (g *Gitlab) container(
	ctx context.Context) (*dagger.Container, error) {

	baseImage := "cgr.dev/chainguard/wolfi-base:latest"

	glabVersion := "1.61.0"
	glabURL := "https://gitlab.com/api/v4/projects/gitlab-org%2Fcli/packages/generic/glab/" + glabVersion + "/glab_" + glabVersion + "_linux_amd64.tar.gz"

	ctr := dag.Container().
		From(baseImage).
		WithExec([]string{"apk", "add", "--no-cache", "git", "wget", "jq", "bash"}).
		WithExec([]string{"sh", "-c", "wget -O /tmp/glab.tar.gz " + glabURL}).
		WithExec([]string{"sh", "-c", "tar -xzf /tmp/glab.tar.gz -C /tmp"}).
		WithExec([]string{"sh", "-c", "mkdir -p /usr/local/bin"}).
		WithExec([]string{"sh", "-c", "cp /tmp/bin/glab /usr/local/bin/glab"}).
		WithExec([]string{"sh", "-c", "chmod +x /usr/local/bin/glab"}).
		WithExec([]string{"sh", "-c", "rm -rf /tmp/glab.tar.gz /tmp/bin"}).
		WithEntrypoint([]string{"glab"})

	return ctr, nil
}
