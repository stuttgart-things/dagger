package main

import (
	"dagger/ansible/internal/dagger"
)

// defaultAnsibleVersion is the single pin for the Ansible release installed into
// the module container. Bump it here; every function picks it up.
const defaultAnsibleVersion = "14.3.1"

func (m *Ansible) container(
	// The base image
	// +optional
	// +default="cgr.dev/chainguard/wolfi-base:latest"
	baseImage string,
	// The Ansible version, defaults to defaultAnsibleVersion
	// +optional
	version string,
) *dagger.Container {

	if baseImage == "" {
		baseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	if version == "" {
		version = defaultAnsibleVersion
	}

	ctr := dag.
		Container().
		From(baseImage)

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
		"sudo",
	})

	// Install Ansible via pip
	ctr = ctr.WithExec([]string{
		"pip3",
		"install",
		"--no-cache-dir",
		"ansible==" + version,
		"hvac",
		"passlib",
		"jmespath",
		// Runtime dependency of the kubernetes.core collection, which ships with
		// ansible but is unusable without it: every module fails with "Failed to
		// import the required Python library (kubernetes)". Needed by playbooks
		// like sthings.container.kind_secrets that talk to a cluster through
		// kubernetes.core rather than kubectl.
		"kubernetes"})

	// Create ansible.cfg with proper settings
	ctr = ctr.WithNewFile("/etc/ansible/ansible.cfg", `[defaults]
allow_world_readable_tmpfiles = true
host_key_checking = false
pipelining = true
timeout = 30
gathering = smart
fact_caching = memory

[ssh_connection]
pipelining = true
ssh_args = -o ControlMaster=auto -o ControlPersist=60s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no

[privilege_escalation]
become = true
become_method = sudo
become_user = root
become_ask_pass = false
`)

	// SET ANSIBLE ENV VAR TO AVOID TMPFILE CHOWN ERRORS
	ctr = ctr.WithEnvVariable("ANSIBLE_ALLOW_WORLD_READABLE_TMPFILES", "true")

	return ctr
}
