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
	modules           = make(map[string]string)
	meta              = make(map[string]string)
	requirements      = make(map[string]string)
	collectionWorkDir = "/collection"
)

type CollectionResult struct {
	Directory *dagger.Directory
	Namespace string
	Name      string
}

type Ansible struct {
	AnsibleContainer *dagger.Container
}

// RunPipeline orchestrates running both Lint and Build steps
func (m *Ansible) RunCollectionBuildPipeline(ctx context.Context, src *dagger.Directory) (*dagger.Directory, error) {

	// INIT COLLECTION
	collection, err := m.InitCollection(ctx, src)
	if err != nil {
		fmt.Println("Failed to initialize collection: %v", err)
	}
	fmt.Println("Collection initialized with namespace:", collection.Namespace, "and name:", collection.Name)

	// initCollectionDir, nampespace, name := m.InitCollection(ctx, src)
	// fmt.Println("NAMESPACE", nampespace)
	// fmt.Println("NAME", name)

	// LOOP OVER ALL FILES IN THE COLLECTION DIRECTORY
	// files, err := initCollectionDir.Entries(ctx)
	// if err != nil {
	// 	fmt.Println("ERROR GETTING ENTRIES: ", err)
	// }
	// fmt.Println("ALL CREATED FILES: ", files)

	// // MODIFY ROLE INCLUDES

	modifiedCollectionDir := m.ModifyRoleIncludes(ctx, collection.Directory.Directory(collection.Namespace+"/"+collection.Name))

	buildCollection := m.Build(ctx, modifiedCollectionDir)

	entries, err := buildCollection.Entries(ctx)
	if err != nil {
		fmt.Println("ERROR GETTING ENTRIES: ", err)
	}

	// Search for a .tgz file
	for _, entry := range entries {
		if strings.HasSuffix(entry, ".tar.gz") {
			fmt.Println("Found .tgz file:", entry)
		}
	}

	fmt.Println("BUILD COLLECTION: ", buildCollection)

	// // MODIFY ROLE INCLUDES
	// src = m.ModifyRoleIncludes(ctx, src)

	// // BUILD
	// src = m.Build(ctx, src)

	return buildCollection, nil

}

// BUILDS A GIVEN COLLECTION DIR TO A ARCHIVE FILE (.TGZ)
func (m *Ansible) GithubRelease(
	ctx context.Context,
	tag string,
	title string,
	files []*dagger.File,
	notes string,
	token *dagger.Secret,
) error {

	releaseOptions := dagger.GhReleaseCreateOpts{
		Repo:      "stuttgart-things/dagger",
		VerifyTag: false,
		Files:     files,
		Token:     token,
	}

	// CREATE GITHUB RELEASE
	err := dag.
		Gh().
		Release().
		Create(
			ctx,
			tag,
			title,
			releaseOptions,
		)
	if err != nil {
		return fmt.Errorf("failed to create GitHub release: %w", err)
	}

	return nil
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
	ansible := m.AnsibleContainer.
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
