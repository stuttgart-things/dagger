package main

import (
	"dagger/helm/internal/dagger"
	"fmt"
)

func (m *Helm) container() *dagger.Container {
	arch := "linux_amd64"
	helmfileVersion := "1.1.3"
	helmfileBin := "helmfile"
	helmfileTar := fmt.Sprintf("%s_%s_%s.tar.gz", helmfileBin, helmfileVersion, arch)
	helmfileURL := fmt.Sprintf("https://github.com/helmfile/helmfile/releases/download/v%s/%s", helmfileVersion, helmfileTar)
	destBinPath := "/usr/bin/" + helmfileBin

	polarisVersion := "9.6.4"
	polarisTar := fmt.Sprintf("polaris_%s.tar.gz", arch)
	polarisURL := fmt.Sprintf("https://github.com/FairwindsOps/polaris/releases/download/%s/polaris_linux_amd64.tar.gz", polarisVersion)
	polarisBinPath := "/usr/bin/polaris"

	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	ctr := dag.Container().
		From(m.BaseImage)

	ctr = ctr.WithExec([]string{"apk", "add", "--no-cache", "wget", "helm", "git"})

	// Install helmfile
	ctr = ctr.WithExec([]string{"wget", "-O", "/tmp/" + helmfileTar, helmfileURL})
	ctr = ctr.WithExec([]string{"tar", "-xzf", "/tmp/" + helmfileTar, "-C", "/tmp/"})
	ctr = ctr.WithExec([]string{"mv", fmt.Sprintf("/tmp/%s", helmfileBin), destBinPath})
	ctr = ctr.WithExec([]string{"chmod", "+x", destBinPath})
	ctr = ctr.WithExec([]string{"helmfile", "init", "--force"})

	// Install Polaris
	ctr = ctr.WithExec([]string{"wget", "-O", "/tmp/" + polarisTar, polarisURL})
	ctr = ctr.WithExec([]string{"tar", "-xzf", "/tmp/" + polarisTar, "-C", "/tmp/"})
	ctr = ctr.WithExec([]string{"mv", "/tmp/polaris", polarisBinPath})
	ctr = ctr.WithExec([]string{"chmod", "+x", polarisBinPath})

	return ctr
}
