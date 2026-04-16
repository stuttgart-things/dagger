package main

import (
	"dagger/oci/internal/dagger"
	"fmt"
)

// container returns a Wolfi-based container with the Flux CLI installed
func (m *Oci) container() *dagger.Container {
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	fluxVersion := "2.8.5"
	fluxURL := fmt.Sprintf(
		"https://github.com/fluxcd/flux2/releases/download/v%s/flux_%s_linux_amd64.tar.gz",
		fluxVersion, fluxVersion,
	)

	ctr := dag.Container().
		From(m.BaseImage).
		WithExec([]string{"apk", "add", "--no-cache", "wget", "curl", "git", "ca-certificates"}).
		WithExec([]string{"sh", "-c", "wget " + fluxURL + " -O /tmp/flux.tar.gz"}).
		WithExec([]string{"tar", "-xzf", "/tmp/flux.tar.gz", "-C", "/tmp/"}).
		WithExec([]string{"mv", "/tmp/flux", "/usr/bin/flux"}).
		WithExec([]string{"chmod", "+x", "/usr/bin/flux"}).
		WithExec([]string{"rm", "-f", "/tmp/flux.tar.gz"})

	return ctr
}
