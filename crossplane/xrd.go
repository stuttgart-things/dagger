package main

import (
	"context"
	"dagger/crossplane/internal/dagger"
	"fmt"
	"strings"
)

// GenerateXRD runs kcl2xrd with the given flags on the provided KCL schema file
func (m *Crossplane) GenerateDefinition(
	ctx context.Context,
	kclSchema *dagger.File,
	// +optional
	// comma-separated flags, e.g., "-i cloudinit.k,-v v1alpha1,-o cloudinit-xrd.yaml"
	cliFlags string,
) *dagger.File {
	containerPath := "/work/schema.k"
	outputPath := "/work/output.yaml" // always use this path

	container := m.GetXplaneContainer(ctx).
		WithMountedFile(containerPath, kclSchema).
		WithWorkdir("/work")

	// Convert comma-separated flags
	flags := strings.Split(cliFlags, ",")

	// Ensure -i points to mounted file
	hasInput := false
	hasOutput := false
	for i, f := range flags {
		if f == "-i" || f == "--input" {
			flags[i+1] = containerPath
			hasInput = true
		}
		if f == "-o" || f == "--output" {
			flags[i+1] = outputPath
			hasOutput = true
		}
	}

	if !hasInput {
		flags = append([]string{"-i", containerPath}, flags...)
	}
	if !hasOutput {
		flags = append(flags, "-o", outputPath)
	}

	// Run kcl2xrd
	container = container.WithExec([]string{
		"sh", "-c",
		"kcl2xrd " + strings.Join(flags, " ") + " > /work/output.yaml",
	})
	// Return the generated file
	return container.File(outputPath)
}

func (m *Crossplane) ModifyDefinition(
	ctx context.Context,
	xrd *dagger.File,
	// +optional
	// +default="apiextensions.crossplane.io/v2"
	apiVersion string,
	// +optional
	// +default="Namespaced"
	scope string,
	// +optional
	// +default="Foreground"
	deletePolicy string,
	// +optional
	singularName string,
) *dagger.File {

	inputPath := "/work/input.yaml"
	outputPath := "/work/xrd-modified.yaml"

	container := m.GetXplaneContainer(ctx).
		WithMountedFile(inputPath, xrd).
		WithWorkdir("/work").
		WithExec([]string{
			"sh", "-c",
			// 🔑 make it writable first
			"cp input.yaml xrd.yaml",
		})

	var cmds []string

	if apiVersion != "" {
		cmds = append(cmds,
			fmt.Sprintf(`yq -i '.apiVersion = "%s"' xrd.yaml`, apiVersion),
		)
	}

	if deletePolicy != "" {
		cmds = append(cmds,
			fmt.Sprintf(`yq -i '.spec.defaultCompositeDeletePolicy = "%s"' xrd.yaml`, deletePolicy),
		)
	}

	if scope != "" {
		cmds = append(cmds,
			fmt.Sprintf(`yq -i '.spec.scope = "%s"' xrd.yaml`, scope),
		)
	}

	if singularName != "" {
		cmds = append(cmds,
			fmt.Sprintf(`yq -i '.spec.names.singular = "%s"' xrd.yaml`, singularName),
		)
	}

	// Ensure output always exists
	cmds = append(cmds, "cp xrd.yaml xrd-modified.yaml")

	container = container.WithExec([]string{
		"sh", "-c",
		strings.Join(cmds, " && "),
	})

	return container.File(outputPath)
}
