package main

import (
	"context"
	"dagger/linting/internal/dagger"
)

func (m *Linting) RunPreCommit(
	ctx context.Context,
	src *dagger.Directory,
	// LintPreCommit runs pre-commit hooks on the provided directory
	// +optional
	// +default=".pre-commit-config.yaml"
	configPath string,
	// +optional
	// +default="pre-commit-findings.txt"
	outputFile string,
	// +optional
	skipHooks []string,
) (*dagger.File, error) {
	if outputFile == "" {
		outputFile = "pre-commit-findings.txt"
	}

	ctr := m.container().
		WithMountedDirectory("/mnt", src).
		WithWorkdir("/mnt")

	// If a custom config path is provided, copy it to the expected location
	if configPath != "" && configPath != ".pre-commit-config.yaml" {
		ctr = ctr.WithExec([]string{"sh", "-c", "cp " + configPath + " .pre-commit-config.yaml"})
	}

	// Build the pre-commit command with SKIP environment variable if needed
	var cmdStr string
	if len(skipHooks) > 0 {
		// Join hook IDs with commas for SKIP env var
		skipList := ""
		for i, hook := range skipHooks {
			if i > 0 {
				skipList += ","
			}
			skipList += hook
		}
		cmdStr = "SKIP=" + skipList + " pre-commit run -a > " + outputFile + " 2>&1 || true"
	} else {
		cmdStr = "pre-commit run -a > " + outputFile + " 2>&1 || true"
	}

	// Run pre-commit
	cmd := []string{"sh", "-c", cmdStr}
	ctr = ctr.WithExec(cmd)

	// Return the linting report file
	report := ctr.File("/mnt/" + outputFile)
	return report, nil
}
