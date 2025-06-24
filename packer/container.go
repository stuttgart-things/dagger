package main

import (
	"dagger/packer/internal/dagger"
	"fmt"
)

func (m *Packer) container(
	// The Packer version to use
	// +optional
	// +default="1.13.1"
	packerVersion,
	// The Packer arch
	// +optional
	// +default="linux_amd64"
	arch string) *dagger.Container {

	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	var (
		packerBin   = "packer"
		packerZip   = fmt.Sprintf("%s_%s_%s.zip", packerBin, packerVersion, arch)
		packerURL   = fmt.Sprintf("https://releases.hashicorp.com/%s/%s/%s", packerBin, packerVersion, packerZip)
		destBinPath = "/usr/bin/" + packerBin

		// Govc configuration
		govcVersion = "0.51.0"
		govcArch    = "Linux_x86_64"
		govcURL     = fmt.Sprintf("https://github.com/vmware/govmomi/releases/download/v%s/govc_%s.tar.gz", govcVersion, govcArch)
		govcTarball = fmt.Sprintf("govc_%s.tar.gz", govcArch)
	)

	ctr := dag.Container().From(m.BaseImage)

	// Install base packages + Ansible dependencies with Wolfi-compatible names
	ctr = ctr.WithExec([]string{"apk", "add", "--no-cache",
		"wget",
		"unzip",
		"bash",
		"coreutils",
		"python3",
		"py3-pip",
		"openssh-client",
		"ca-certificates-bundle",
		"cdrkit",
		"git",
		"sshpass",
		"gzip", // Already correct in Wolfi
	})

	// Install Ansible via pip
	ctr = ctr.WithExec([]string{"pip3", "install", "--no-cache-dir", "ansible", "hvac", "passlib"})

	// Install Packer
	ctr = ctr.WithExec([]string{"wget", "-q", packerURL})  // Added -q for quieter output
	ctr = ctr.WithExec([]string{"unzip", "-q", packerZip}) // Added -q for quieter output
	ctr = ctr.WithExec([]string{"mv", packerBin, destBinPath})
	ctr = ctr.WithExec([]string{"chmod", "+x", destBinPath})
	ctr = ctr.WithExec([]string{"rm", packerZip}) // Clean up packer zip

	// Install govc
	ctr = ctr.WithExec([]string{"wget", "-q", govcURL, "-O", govcTarball})
	ctr = ctr.WithExec([]string{"tar", "-xzf", govcTarball})
	ctr = ctr.WithExec([]string{"mv", "govc", "/usr/bin"})
	ctr = ctr.WithExec([]string{"chmod", "+x", "/usr/bin/govc"})
	ctr = ctr.WithExec([]string{"rm", govcTarball}) // Clean up tarball

	return ctr
}
