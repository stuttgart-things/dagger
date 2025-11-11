package main

import (
	"context"
	"dagger/linting/internal/dagger"
)

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
