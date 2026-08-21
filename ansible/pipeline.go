package main

import (
	"context"
	"dagger/ansible/internal/dagger"
	"fmt"
	"strings"
)

// RunCollectionBuildPipeline orchestrates init, modify and build of an ansible collection
func (m *Ansible) RunCollectionBuildPipeline(
	ctx context.Context,
	src *dagger.Directory,
	// The Ansible version, defaults to defaultAnsibleVersion
	// +optional
	ansibleVersion string,
	// Collection version, passed through to InitCollection. Falls back to a
	// generated calendar version, which is NOT monotonic.
	// +optional
	version string,
) (*dagger.Directory, error) {

	// INIT COLLECTION
	collection, err := m.InitCollection(ctx, src, version)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize collection: %w", err)
	}
	fmt.Println("Collection initialized with namespace:", collection.Namespace, "and name:", collection.Name)

	// MODIFY COLLECTION
	modifiedCollectionDir := m.ModifyRoleIncludes(ctx, collection.Directory.Directory(collection.Namespace+"/"+collection.Name), ansibleVersion)

	// BUILD COLLECTION
	buildCollection := m.Build(ctx, modifiedCollectionDir)

	// SEARCH FOR BUILD ARTIFACT
	entries, err := buildCollection.Entries(ctx)
	if err != nil {
		fmt.Println("ERROR GETTING ENTRIES: ", err)
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry, ".tar.gz") {
			fmt.Println("COLLECTION ARTIFACT", entry)
		}
	}

	return buildCollection, nil
}
