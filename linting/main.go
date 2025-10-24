// A generated module for Linting functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/linting/internal/dagger"
)

type Linting struct{
	BaseImage string
}

func (m *Linting) LintYAML(
	ctx context.Context,
	src *dagger.Directory,
	// LintYAML lints YAML files in the provided directory
	// +optional
	// +default=".yamllint"
	configPath string,
	// +optional
	// +default="yamllint-findings.txt"
	outputFile string,
) (*dagger.File, error) {
	if outputFile == "" {
		outputFile = "yamllint-findings.txt"
	}

	ctr := m.container().
		WithMountedDirectory("/mnt", src).
		WithWorkdir("/mnt")

	// Build command based on whether config is provided
	var cmd []string
	if configPath != "" {
		cmd = []string{"sh", "-c", "yamllint -c " + configPath + " . > " + outputFile + " 2>&1 || true"}
	} else {
		cmd = []string{"sh", "-c", "yamllint . > " + outputFile + " 2>&1 || true"}
	}

	// Run yamllint
	ctr = ctr.WithExec(cmd)

	// Return the linting report file
	report := ctr.File("/mnt/" + outputFile)
	return report, nil
}

func (m *Linting) LintMarkdown(
	ctx context.Context,
	src *dagger.Directory,
	// LintMarkdown lints Markdown files in the provided directory
	// +optional
	// +default=".mdlrc"
	configPath string,
	// +optional
	// +default="markdown-findings.txt"
	outputFile string,
) (*dagger.File, error) {
	if outputFile == "" {
		outputFile = "markdown-findings.txt"
	}

	ctr := m.container().
		WithMountedDirectory("/mnt", src).
		WithWorkdir("/mnt")

	// Build command based on whether config is provided
	var cmd []string
	if configPath != "" {
		cmd = []string{"sh", "-c", "mdl -c " + configPath + " . > " + outputFile + " 2>&1 || true"}
	} else {
		cmd = []string{"sh", "-c", "mdl . > " + outputFile + " 2>&1 || true"}
	}

	// Run mdl
	ctr = ctr.WithExec(cmd)

	// Return the linting report file
	report := ctr.File("/mnt/" + outputFile)
	return report, nil
}
