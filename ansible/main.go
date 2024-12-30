// A generated module for Ansible functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/ansible/internal/dagger"
)

type Ansible struct {
	AnsibleContainer *dagger.Container
}

// Init Ansible Collection Structure
func (m *Ansible) InitCollection(ctx context.Context, src *dagger.Directory, namespace, name string) *dagger.Directory {

	ansible := m.AnsibleContainer.
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"ansible-galaxy", "collection", "init", namespace + "." + name})

	return ansible.Directory("/src")
}

func New(
	// ansible container
	// It need contain ansible
	// +optional
	ansibleContainer *dagger.Container,

) *Ansible {
	ansible := &Ansible{}

	if ansibleContainer != nil {
		ansible.AnsibleContainer = ansibleContainer
	} else {
		ansible.AnsibleContainer = ansible.GetAnsibleContainer()
	}
	return ansible
}

// GetXplaneContainer return the default image for helm
func (m *Ansible) GetAnsibleContainer() *dagger.Container {
	return dag.Container().
		From("ghcr.io/stuttgart-things/sthings-ansible:11.1.0")
}
