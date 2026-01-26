package main

import (
	"context"
	"dagger/sops/internal/dagger"
	"fmt"
	"strings"
)

// GenerateSopsConfig generates a .sops.yaml configuration file with creation rules for the given AGE key.
// The fileExtensions parameter accepts a comma-separated list of extensions (e.g., "yaml,json,env").
// If not provided, defaults to "yaml,json".
func (m *Sops) GenerateSopsConfig(
	ctx context.Context,
	agePublicKey string,
	// +optional
	fileExtensions string,
) (*dagger.File, error) {
	if agePublicKey == "" {
		return nil, fmt.Errorf("agePublicKey is required")
	}

	// Default file extensions
	if fileExtensions == "" {
		fileExtensions = "yaml,json"
	}

	// Build creation rules
	var rules []string
	extensions := strings.Split(fileExtensions, ",")
	for _, ext := range extensions {
		ext = strings.TrimSpace(ext)
		ext = strings.TrimPrefix(ext, ".")
		if ext != "" {
			rule := fmt.Sprintf(`  - path_regex: .*\.%s
    age: "%s"`, ext, agePublicKey)
			rules = append(rules, rule)
		}
	}

	configContent := fmt.Sprintf(`---
creation_rules:
%s
`, strings.Join(rules, "\n"))

	ctr := dag.Container().
		From("alpine:latest").
		WithNewFile("/tmp/.sops.yaml", configContent)

	return ctr.File("/tmp/.sops.yaml"), nil
}
