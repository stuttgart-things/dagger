package main

import (
	"context"
	"dagger/linting/internal/dagger"
)

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
