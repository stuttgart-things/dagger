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

type Kcl struct {
	BaseImage string
}

// TestKcl runs a basic KCL test to verify the container and CLI are working
func (m *Kcl) TestKcl(ctx context.Context) (string, error) {
	// Create a very simple KCL test file to avoid complex parsing
	testKcl := `name = "hello-kcl"
version = "1.0.0"
`

	return m.container().
		WithNewFile("/tmp/simple.k", testKcl).
		WithWorkdir("/tmp").
		WithExec([]string{"kcl", "run", "simple.k"}).
		Stdout(ctx)
}

// RunKcl executes KCL code from a provided directory
func (m *Kcl) RunKcl(ctx context.Context, source *dagger.Directory, entrypoint string) (string, error) {
	if entrypoint == "" {
		entrypoint = "main.k"
	}

	return m.container().
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"kcl", "run", entrypoint}).
		Stdout(ctx)
}

// ValidateKcl validates KCL configuration files by compiling them
func (m *Kcl) ValidateKcl(ctx context.Context, source *dagger.Directory) (string, error) {
	// Validation by compilation - if files compile without errors, they are valid
	return m.container().
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"sh", "-c", "kcl run main.k > /dev/null && echo 'Validation successful: KCL files are syntactically correct'"}).
		Stdout(ctx)
}

// KclVersion returns the installed KCL version
func (m *Kcl) KclVersion(ctx context.Context) (string, error) {
	return m.container().
		WithExec([]string{"kcl", "version"}).
		Stdout(ctx)
}

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
