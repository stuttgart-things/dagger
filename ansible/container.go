package main

import (
	"dagger/ansible/internal/dagger"
)

func (m *Ansible) container(
	// The Packer arch
	// +optional
	// +default="linux_amd64"
	arch string) *dagger.Container {

	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	ctr := dag.
		Container().
		From(m.BaseImage)

	// INSTALL BASE PACKAGES + ANSIBLE DEPENDENCIES WITH WOLFI-COMPATIBLE NAMES
	ctr = ctr.WithExec([]string{
		"apk",
		"add",
		"--no-cache",
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
		"gzip",
		"acl",
	})

	// Install Ansible via pip
	ctr = ctr.WithExec([]string{
		"pip3",
		"install",
		"--no-cache-dir",
		"ansible",
		"hvac",
		"passlib"})

	// SET ANSIBLE ENV VAR TO AVOID TMPFILE CHOWN ERRORS
	ctr = ctr.WithEnvVariable("ANSIBLE_ALLOW_WORLD_READABLE_TMPFILES", "true")

	return ctr
}
