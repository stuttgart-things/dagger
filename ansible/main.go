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
	"fmt"

	// "dagger/ansible/collections"
	"dagger/ansible/collections"
	"dagger/ansible/internal/dagger"
)

var (
	playbooks         = make(map[string]string)
	vars              = make(map[string]string)
	templates         = make(map[string]string)
	meta              = make(map[string]string)
	requirements      = make(map[string]string)
	src2              *dagger.Directory
	collectionWorkDir = "/collection"
)

type Ansible struct {
	AnsibleContainer *dagger.Container
}

// Init Ansible Collection Structure
func (m *Ansible) InitCollection(ctx context.Context, src *dagger.Directory, namespace, name string) *dagger.Directory {

	allCollectionFiles, err := src.Entries(ctx)
	if err != nil {
		fmt.Println("ERROR GETTING ENTRIES: ", err)
	}

	// GET COLLECTION ENTRIES FROM THE (GIVEN) SOURCE DIRECTORY
	for _, entry := range allCollectionFiles {
		content, err := src.File(entry).Contents(ctx)
		if err != nil {
			fmt.Println("ERROR GETTING CONTENTS: ", err)
		}
		playbooks, vars, templates, meta, requirements = collections.ProcessCollectionFile([]byte(content), playbooks, vars, templates, meta, requirements)
	}

	collectionNamespace := meta["namespace"]
	collectionName := meta["name"]
	collectionContentDir := collectionWorkDir + "/" + collectionNamespace + "/" + collectionName + "/"

	// INIT COLLECTION
	ansible := m.AnsibleContainer.
		WithDirectory(collectionWorkDir, src).
		WithWorkdir(collectionWorkDir).
		WithExec([]string{"ansible-galaxy", "collection", "init", collectionNamespace + "." + collectionName})

	// CREATE PLAYBOOKS ON COLLECTION DIRECTORY
	for key, value := range playbooks {
		ansible = ansible.WithNewFile(collectionContentDir+"plays/"+key+".yaml", value)
	}

	// CREATE VARS ON COLLECTION DIRECTORY
	for key, value := range vars {
		ansible = ansible.WithNewFile(collectionContentDir+"plays/vars/"+key+".yaml", value)
	}

	// CREATE TEMPLATES ON COLLECTION DIRECTORY
	for key, value := range templates {
		ansible = ansible.WithNewFile(collectionContentDir+"plays/templates/"+key+".yaml", value)
	}

	// CREATE REQUIREMENTS FILE ON CONTAINER + INSTALL ROLES
	if requirements["requirements"] != "" {
		ansible = ansible.WithNewFile(collectionContentDir+"meta/collection-requirements.yaml", requirements["requirements"])
		ansible = ansible.WithExec([]string{"ansible-galaxy", "install", "-r", collectionContentDir + "meta/collection-requirements.yaml", "-p", collectionContentDir + "/roles"})
	}

	return ansible.Directory(collectionWorkDir)
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

// GetAnsibleContainer return the default image for helm
func (m *Ansible) GetAnsibleContainer() *dagger.Container {
	return dag.Container().
		From("ghcr.io/stuttgart-things/sthings-ansible:11.1.0")
}
