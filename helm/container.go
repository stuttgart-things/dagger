package main

import (
	"dagger/helm/internal/dagger"
	"fmt"
)

func (m *Helm) container() *dagger.Container {
	arch := "linux_amd64"

	helmfileVersion := "1.1.7"
	helmfileBin := "helmfile"
	helmfileTar := fmt.Sprintf("%s_%s_%s.tar.gz", helmfileBin, helmfileVersion, arch)
	helmfileURL := fmt.Sprintf("https://github.com/helmfile/helmfile/releases/download/v%s/%s", helmfileVersion, helmfileTar)
	helmfileBinPath := "/usr/bin/" + helmfileBin

	polarisVersion := "10.1.1"
	polarisTar := fmt.Sprintf("polaris_%s.tar.gz", arch)
	polarisURL := fmt.Sprintf("https://github.com/FairwindsOps/polaris/releases/download/%s/polaris_linux_amd64.tar.gz", polarisVersion)
	polarisBinPath := "/usr/bin/polaris"

	valsVersion := "0.42.1"
	valsBin := "vals"
	valsTar := fmt.Sprintf("%s_%s_%s.tar.gz", valsBin, valsVersion, arch)
	valsURL := fmt.Sprintf("https://github.com/helmfile/vals/releases/download/v%s/%s", valsVersion, valsTar)
	valsBinPath := "/usr/bin/" + valsBin

	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	ctr := dag.Container().
		From(m.BaseImage)

	ctr = ctr.WithExec([]string{"apk", "add", "--no-cache", "wget", "helm", "git"})

	// INSTALL HELMFILE
	ctr = ctr.WithExec([]string{"wget", "-O", "/tmp/" + helmfileTar, helmfileURL})
	ctr = ctr.WithExec([]string{"tar", "-xzf", "/tmp/" + helmfileTar, "-C", "/tmp/"})
	ctr = ctr.WithExec([]string{"mv", fmt.Sprintf("/tmp/%s", helmfileBin), helmfileBinPath})
	ctr = ctr.WithExec([]string{"chmod", "+x", helmfileBinPath})
	ctr = ctr.WithExec([]string{"helmfile", "init", "--force"})

	// INSTALL POLARIS
	ctr = ctr.WithExec([]string{"wget", "-O", "/tmp/" + polarisTar, polarisURL})
	ctr = ctr.WithExec([]string{"tar", "-xzf", "/tmp/" + polarisTar, "-C", "/tmp/"})
	ctr = ctr.WithExec([]string{"mv", "/tmp/polaris", polarisBinPath})
	ctr = ctr.WithExec([]string{"chmod", "+x", polarisBinPath})

	// INSTALL VALS
	ctr = ctr.WithExec([]string{"wget", "-O", "/tmp/" + valsTar, valsURL})
	ctr = ctr.WithExec([]string{"tar", "-xzf", "/tmp/" + valsTar, "-C", "/tmp/"})
	ctr = ctr.WithExec([]string{"mv", fmt.Sprintf("/tmp/%s", valsBin), valsBinPath})
	ctr = ctr.WithExec([]string{"chmod", "+x", valsBinPath})

	return ctr
}
