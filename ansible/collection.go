package main

import (
	"context"
	"dagger/ansible/collections"
	"dagger/ansible/internal/dagger"
	"fmt"
)

type CollectionResult struct {
	Directory *dagger.Directory
	Namespace string
	Name      string
}

// BUILDS A GIVEN COLLECTION DIR TO A ARCHIVE FILE (.TGZ)
func (m *Ansible) Build(
	ctx context.Context,
	src *dagger.Directory) *dagger.Directory {

	ansible := m.container(m.BaseImage).
		WithDirectory(collectionWorkDir, src).
		WithWorkdir(collectionWorkDir).
		WithExec([]string{"ls", "-lta"}).
		WithExec([]string{"ansible-galaxy", "collection", "build"}).
		WithExec([]string{"ls", "-lta"})

	return ansible.Directory(collectionWorkDir)
}

// INIT ANSIBLE COLLECTION STRUCTURE
func (m *Ansible) InitCollection(
	ctx context.Context,
	src *dagger.Directory) (*CollectionResult, error) {

	metaInformation := make(map[string]interface{})

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
		playbooks, vars, modules, templates, meta, requirements = collections.ProcessCollectionFile([]byte(content), playbooks, vars, modules, templates, meta, requirements)
	}

	metaInformation["namespace"] = meta["namespace"]
	metaInformation["name"] = meta["name"]
	metaInformation["authors"] = meta["authors"]
	metaInformation["version"] = collections.GenerateSemanticVersion()

	collectionContentDir := collectionWorkDir + "/" + metaInformation["namespace"].(string) + "/" + metaInformation["name"].(string)

	// INIT COLLECTION
	ansible := m.container(m.BaseImage).
		WithDirectory(collectionWorkDir, src).
		WithWorkdir(collectionWorkDir).
		WithExec([]string{"ansible-galaxy", "collection", "init", metaInformation["namespace"].(string) + "." + metaInformation["name"].(string)})

	// CHANGE META INFORMATION
	renderedGalaxyFile := collections.RenderTemplate(collections.GalaxyConfig, metaInformation)
	fmt.Println("RENDERED GALAXY FILE: ", renderedGalaxyFile)
	ansible = ansible.WithNewFile(collectionContentDir+"/galaxy.yml", renderedGalaxyFile)

	//ansible = ansible.WithExec([]string{"ansible-galaxy", "install", "-r", collectionContentDir + "meta/collection-requirements.yaml", "-p", collectionContentDir + "/roles"})

	// CREATE PLAYBOOKS ON COLLECTION DIRECTORY
	for key, value := range playbooks {
		ansible = ansible.WithNewFile(collectionContentDir+"/playbooks/"+key+".yaml", value)
	}

	// CREATE VARS ON COLLECTION DIRECTORY
	for key, value := range vars {
		ansible = ansible.WithNewFile(collectionContentDir+"/playbooks/vars/"+key+".yaml", value)
	}

	// CREATE TEMPLATES ON COLLECTION DIRECTORY
	for key, value := range templates {
		ansible = ansible.WithNewFile(collectionContentDir+"/playbooks/templates/"+key+".yaml", value)
	}

	// CREATE MODULES ON COLLECTION DIRECTORY
	for key, value := range modules {
		ansible = ansible.WithNewFile(collectionContentDir+"/plugins/module_utils/"+key+".py", value)
	}

	// CREATE REQUIREMENTS FILE ON CONTAINER + INSTALL ROLES
	if requirements["requirements"] != "" {
		ansible = ansible.WithNewFile(collectionContentDir+"/meta/collection-requirements.yaml", requirements["requirements"])
		ansible = ansible.WithExec([]string{"ansible-galaxy", "install", "-r", collectionContentDir + "/meta/collection-requirements.yaml", "-p", collectionContentDir + "/roles"})
	}

	return &CollectionResult{
		Directory: ansible.Directory(collectionWorkDir),
		Namespace: metaInformation["namespace"].(string),
		Name:      metaInformation["name"].(string),
	}, nil
}
