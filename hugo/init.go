package main

import (
	"context"
	"dagger/hugo/internal/dagger"
)

func (m *Hugo) InitSite(
	ctx context.Context,
	name string,
	config *dagger.File,
	content *dagger.Directory,
	// The Theme to use
	// +optional
	// +default="github.com/joshed-io/reveal-hugo"
	theme string,

) (*dagger.Directory, error) {
	// CREATE BASE CONTAINER WITH HUGO INSTALLED
	baseCtr := m.container()

	// CREATE NEW SITE AND INITIALIZE MODULES
	ctr := baseCtr.
		WithExec([]string{"hugo", "new", "site", name}).
		WithWorkdir(name).
		WithFile("hugo.toml", config).
		WithExec([]string{"hugo", "mod", "init", name}).
		WithExec([]string{"hugo", "mod", "get", theme}).
		WithExec([]string{"hugo", "mod", "tidy"}).
		WithExec([]string{"hugo", "mod", "vendor"}).
		WithExec([]string{"tree"})

	// GET THE INITIALIZED SITE DIRECTORY
	siteDir := ctr.Directory(".").
		WithFile("hugo.toml", config).
		WithDirectory("content", content)

	// 	// LIST ALL FILE ENTRIES
	// 	// entries, err := siteDir.Entries(ctx)
	// 	// if err != nil {
	// 	// 	panic(err)
	// 	// }

	// 	// for _, entry := range entries {
	// 	// 	println(entry)
	// 	// }

	// if err := listAllFiles(ctx, siteDir, "."); err != nil {
	// 	panic(err)
	// }

	return siteDir, nil
}

// func listAllFiles(ctx context.Context, dir *dagger.Directory, path string) error {
// 	entries, err := dir.Entries(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to list entries at %s: %w", path, err)
// 	}

// 	for _, entry := range entries {
// 		fullPath := path + "/" + entry

// 		// Try to read as a subdirectory
// 		subdir := dir.Directory(entry)
// 		subEntries, err := subdir.Entries(ctx)
// 		if err == nil && len(subEntries) > 0 {
// 			// It's a directory
// 			fmt.Println(fullPath + "/")
// 			if err := listAllFiles(ctx, subdir, fullPath); err != nil {
// 				return err
// 			}
// 		} else {
// 			// It's a file or an empty dir
// 			fmt.Println(fullPath)
// 		}
// 	}
// 	return nil
// }
