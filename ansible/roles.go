package main

import (
	"context"
	"dagger/ansible/internal/dagger"
	"fmt"
	"strings"
)

// BUILDS A GIVEN COLLECTION DIR TO A ARCHIVE FILE (.TGZ)
func (m *Ansible) ModifyRoleIncludes(
	ctx context.Context,
	src *dagger.Directory) *dagger.Directory {

	ansible := m.container(m.BaseImage).
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
