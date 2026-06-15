package main

import (
	"dagger/linting/internal/dagger"
)

func (m *Linting) container() *dagger.Container {
	if m.BaseImage == "" {
		// Pin the base image (tag, not :latest) so pre-commit runs are
		// reproducible and don't depend on docker.io resolving a mutable
		// alpine:latest on every run. Overridable via --base-image. See #288.
		m.BaseImage = "alpine:3.21"
	}

	ctr := dag.Container().From(m.BaseImage)

	ctr = ctr.WithExec([]string{
		"apk",
		"add",
		"--no-cache",
		"wget",
		"yamllint",
		"ruby",
		"python3",
		"py3-pip",
		"git",
		"bash",
		"shellcheck",
		"jq",
	})

	// Install the hadolint static binary so the `hadolint` pre-commit hook runs
	// natively in-container. hadolint ships as a static musl binary, so it runs
	// directly on Alpine — no Docker daemon needed (there is none inside a Dagger
	// container, which is why the old `hadolint-docker` hook never worked). wget
	// is already installed above. Pin the version for reproducibility. See #302.
	ctr = ctr.WithExec([]string{
		"sh", "-c",
		"wget -qO /usr/local/bin/hadolint " +
			"https://github.com/hadolint/hadolint/releases/download/v2.12.0/hadolint-Linux-x86_64 " +
			"&& chmod +x /usr/local/bin/hadolint",
	})

	ctr = ctr.WithExec([]string{
		"gem",
		"install",
		"mdl",
		"--no-document",
	})

	ctr = ctr.WithExec([]string{
		"pip3",
		"install",
		"--break-system-packages",
		"pre-commit",
		"detect-secrets",
	})

	return ctr
}
