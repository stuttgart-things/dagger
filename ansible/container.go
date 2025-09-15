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

	// Install base packages + Ansible dependencies with Wolfi-compatible names
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

	// Set Ansible env var to avoid tmpfile chown errors
	ctr = ctr.WithEnvVariable("ANSIBLE_ALLOW_WORLD_READABLE_TMPFILES", "true")

	return ctr
}
