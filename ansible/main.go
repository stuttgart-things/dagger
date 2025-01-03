/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"context"
	"fmt"
	"strings"

	"dagger/ansible/collections"
	"dagger/ansible/internal/dagger"
)

var (
	playbooks         = make(map[string]string)
	vars              = make(map[string]string)
	templates         = make(map[string]string)
	meta              = make(map[string]string)
	requirements      = make(map[string]string)
	collectionWorkDir = "/collection"
)

type Ansible struct {
	AnsibleContainer *dagger.Container
}

// BUILDS A GIVEN COLLECTION DIR TO A ARCHIVE FILE (.TGZ)
func (m *Ansible) Build(ctx context.Context, src *dagger.Directory) *dagger.Directory {

	ansible := m.AnsibleContainer.
		WithDirectory(collectionWorkDir, src).
		WithWorkdir(collectionWorkDir).
		WithExec([]string{"ls", "-lta"}).
		WithExec([]string{"ansible-galaxy", "collection", "build"}).
		WithExec([]string{"ls", "-lta"})

	return ansible.Directory(collectionWorkDir)
}

// BUILDS A GIVEN COLLECTION DIR TO A ARCHIVE FILE (.TGZ)
func (m *Ansible) ModifyRoleIncludes(ctx context.Context, src *dagger.Directory) *dagger.Directory {

	ansible := m.AnsibleContainer.
		WithDirectory(collectionWorkDir, src).
		WithWorkdir(collectionWorkDir)

	roleDirs, err := ansible.Directory(collectionWorkDir + "/roles").Entries(ctx)
	if err != nil {
		fmt.Println("ERROR GETTING ENTRIES: ", err)
	}

	// RENAME ALL ROLENAMES WITH DASHES TO UNDERSCORES
	for _, roleDir := range roleDirs {

		if strings.Contains(roleDir, "-") {

			// SET NEW COLLECTION ROLE NAME
			collectionRoleName := strings.Replace(roleDir, "-", "_", -1)

			// RENAME ROLE DIR
			ansible = ansible.WithExec([]string{"mv", collectionWorkDir + "/roles/" + roleDir, collectionWorkDir + "/roles/" + collectionRoleName})

			// REPLACE ALL ROLE REFERENCES IN YAML FILES INSIDE THE ROLES DIRECTORY
			ansible = ansible.WithExec([]string{
				"sh", "-c",
				fmt.Sprintf(
					`find %s -name "*.yaml" -type f -exec sed -i "s/%s/%s/g" {} +`,
					collectionWorkDir+"/roles/",
					roleDir,
					collectionRoleName,
				),
			})
		}
	}

	return ansible.Directory(collectionWorkDir)

}

// INIT ANSIBLE COLLECTION STRUCTURE
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
		ansible = ansible.WithNewFile(collectionContentDir+"playbooks/"+key+".yaml", value)
	}

	// CREATE VARS ON COLLECTION DIRECTORY
	for key, value := range vars {
		ansible = ansible.WithNewFile(collectionContentDir+"playbooks/vars/"+key+".yaml", value)
	}

	// CREATE TEMPLATES ON COLLECTION DIRECTORY
	for key, value := range templates {
		ansible = ansible.WithNewFile(collectionContentDir+"playbooks/templates/"+key+".yaml", value)
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
