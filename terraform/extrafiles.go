package main

import (
	"context"
	"fmt"

	"dagger/terraform/internal/dagger"
)

// withExtraFiles places extraFiles next to the Terraform code in workDir, each
// under its base name, e.g. a CA bundle a provider reads as
// "${path.module}/ca.crt". A name already taken in terraformDir, by another
// extra file or by a reserved name (terraform.tfvars.json while
// secretJsonVariables is mounted) is an error: nothing is shadowed silently.
// It returns the names placed, so the caller can drop them from the
// directory it hands back.
func withExtraFiles(
	ctx context.Context,
	ctr *dagger.Container,
	workDir string,
	terraformDir *dagger.Directory,
	extraFiles []*dagger.File,
	reserved ...string,
) (*dagger.Container, []string, error) {
	taken := map[string]string{}
	for _, r := range reserved {
		taken[r] = "a file the module mounts itself"
	}

	names := make([]string, 0, len(extraFiles))
	for _, f := range extraFiles {
		name, err := f.Name(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("extraFiles: reading a file name: %w", err)
		}
		if by, ok := taken[name]; ok {
			return nil, nil, fmt.Errorf("extraFiles: %q clashes with %s", name, by)
		}
		exists, err := terraformDir.Exists(ctx, name)
		if err != nil {
			return nil, nil, fmt.Errorf("extraFiles: checking %q in terraformDir: %w", name, err)
		}
		if exists {
			return nil, nil, fmt.Errorf("extraFiles: %q already exists in terraformDir, refusing to overwrite it", name)
		}
		taken[name] = "another extra file of the same name"
		names = append(names, name)
		ctr = ctr.WithFile(workDir+"/"+name, f)
	}

	return ctr, names, nil
}
