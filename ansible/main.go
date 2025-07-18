/*
Package main provides a Dagger-based automation interface for building, modifying, and executing
Ansible collections and playbooks within containerized environments.

This module includes functionality to:
  - Initialize Ansible collections from source files with metadata extraction and templating.
  - Modify role names to conform to Ansible Galaxy standards (dashes to underscores).
  - Build Ansible collections into distributable `.tar.gz` archives.
  - Execute Ansible playbooks with optional support for Vault secrets and inventories.
  - Automate GitHub releases for built collections using GitHub tokens.

The Ansible pipeline leverages a customizable Ansible container and supports:
  - Injecting playbooks, roles, templates, and modules.
  - Enforcing semantic versioning for collections.
  - Supporting Vault authentication via AppRole for secrets at runtime.
  - Executing multiple playbooks with optional parameters and environment secrets.

Usage is built around the Dagger API and is meant for CI/CD pipelines, release automation,
or repeatable infrastructure-as-code workflows.
*/

package main

import (
	"dagger/ansible/internal/dagger"
)

var (
	playbooks         = make(map[string]string)
	vars              = make(map[string]string)
	templates         = make(map[string]string)
	modules           = make(map[string]string)
	meta              = make(map[string]string)
	requirements      = make(map[string]string)
	collectionWorkDir = "/collection"
	workDir           = "/src"
)

type Ansible struct {
	AnsibleContainer *dagger.Container
	// Base Wolfi image to use
	// +optional
	// +default="cgr.dev/chainguard/wolfi-base:latest"
	BaseImage string
}

func New(
	// ansible container
	// It need contain ansible
	// +optional
	ansibleContainer *dagger.Container,
	// +optional
	githubContainer *dagger.Container,

) *Ansible {
	ansible := &Ansible{}

	if ansibleContainer != nil {
		ansible.AnsibleContainer = ansibleContainer
	}

	return ansible
}
