package main

import (
	"dagger/helm/internal/dagger"
	"fmt"

	"dagger.io/dagger/dag"
)

func (m *Helm) container(
	// The Packer arch
	// +optional
	// +default="linux_amd64"
	arch string) *dagger.Container {

	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	// INSTALL HELMFILE BINARY
	var (
		helmfileVersion = "1.1.3"
		helmfileBin     = "helmfile"
		helmfileTar     = fmt.Sprintf("%s_%s_%s.tar.gz", helmfileBin, helmfileVersion, arch)
		helmfileURL     = fmt.Sprintf("https://github.com/helmfile/helmfile/releases/download/v%s/%s", helmfileVersion, helmfileTar)
		destBinPath     = "/usr/bin/" + helmfileBin
	)

	ctr := dag.
		Container().
		From(m.BaseImage)

	// INSTALL HELM
	ctr = ctr.WithExec([]string{
		"apk",
		"add",
		"--no-cache",
		"helm",
	})

	ctr = ctr.WithExec([]string{
		"wget", "-O", "/tmp/" + helmfileTar, helmfileURL,
	})

	ctr = ctr.WithExec([]string{
		"tar", "-xzf", "/tmp/" + helmfileTar, "-C", "/tmp/",
	})

	ctr = ctr.WithExec([]string{
		"mv", fmt.Sprintf("/tmp/%s", helmfileBin), destBinPath,
	})

	ctr = ctr.WithExec([]string{
		"chmod", "+x", destBinPath,
	})

	return ctr
}
