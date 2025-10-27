// A generated module for KCL functions
//
// This module provides KCL (KCL Configuration Language) functionality through Dagger.
// It includes functions to run KCL code, validate configurations, and test the KCL CLI.
//
// KCL is a constraint-based record and functional language hosted by CNCF that enhances
// the writing of complex configurations, including those for cloud-native scenarios.

package main

import (
	"context"
	"dagger/kcl/internal/dagger"
)

// ConvertCrd converts a single CRD file (local or web source) to KCL models
// Returns a directory containing the generated models/v1beta1/ structure
func (m *Kcl) ConvertCrd(ctx context.Context,
	// +optional
	crdSource string,
	// +optional
	crdFile *dagger.File,
) *dagger.Directory {
	ctr := m.container()

	// Handle local file or web source
	if crdFile != nil {
		// Local file provided
		ctr = ctr.WithMountedFile("/tmp/crd.yaml", crdFile)
	} else if crdSource != "" {
		// Web source - download the CRD
		ctr = ctr.WithExec([]string{"wget", "-O", "/tmp/crd.yaml", crdSource})
	} else {
		// Create empty directory if no source provided
		return dag.Directory()
	}

	// Import CRD and generate KCL models
	ctr = ctr.WithExec([]string{"kcl", "import", "-m", "crd", "/tmp/crd.yaml"})

	// Return the generated models directory
	return ctr.Directory("/models")
}

// ConvertCrdToDirectory converts a CRD and outputs the models to a specified working directory
// This version gives more control over the output structure and allows custom organization
func (m *Kcl) ConvertCrdToDirectory(ctx context.Context, workdir *dagger.Directory,
	// +optional
	crdSource string,
	// +optional
	crdFile *dagger.File,
	// +optional
	outputPath string,
) *dagger.Directory {
	if outputPath == "" {
		outputPath = "models"
	}

	ctr := m.container().WithMountedDirectory("/workspace", workdir).WithWorkdir("/workspace")

	// Handle local file or web source
	if crdFile != nil {
		// Local file provided
		ctr = ctr.WithMountedFile("/tmp/crd.yaml", crdFile)
	} else if crdSource != "" {
		// Web source - download the CRD
		ctr = ctr.WithExec([]string{"wget", "-O", "/tmp/crd.yaml", crdSource})
	} else {
		// Return original directory if no source provided
		return workdir
	}

	// Create output directory if it doesn't exist
	ctr = ctr.WithExec([]string{"mkdir", "-p", outputPath})

	// Import CRD and generate KCL models to specific output path
	ctr = ctr.WithExec([]string{"sh", "-c", "cd " + outputPath + " && kcl import -m crd /tmp/crd.yaml"})

	// Return the updated directory with models
	return ctr.Directory("/workspace")
}
