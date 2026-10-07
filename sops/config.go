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
	// Encrypt only the values whose keys match, e.g. '^(data|stringData)$'
	// for Kubernetes Secrets that Flux applies. Added to every rule.
	// +optional
	encryptedRegex string,
	// One rule for files matching this regex, e.g. '.*\.enc\.yaml$',
	// instead of one rule per file extension
	// +optional
	pathRegex string,
) (*dagger.File, error) {
	if agePublicKey == "" {
		return nil, fmt.Errorf("agePublicKey is required")
	}

	// Default file extensions
	if fileExtensions == "" {
		fileExtensions = "yaml,json"
	}

	// Build creation rules
	var pathRegexes []string
	if pathRegex != "" {
		pathRegexes = []string{pathRegex}
	} else {
		for _, ext := range strings.Split(fileExtensions, ",") {
			ext = strings.TrimPrefix(strings.TrimSpace(ext), ".")
			if ext != "" {
				pathRegexes = append(pathRegexes, `.*\.`+ext)
			}
		}
	}

	var rules []string
	for _, re := range pathRegexes {
		rule := fmt.Sprintf("  - path_regex: %s\n", yamlQuote(re))
		if encryptedRegex != "" {
			rule += fmt.Sprintf("    encrypted_regex: %s\n", yamlQuote(encryptedRegex))
		}
		rule += fmt.Sprintf("    age: %q\n", agePublicKey)
		rules = append(rules, rule)
	}

	configContent := "---\ncreation_rules:\n" + strings.Join(rules, "")

	// A plain file; no container (and no image pull) is needed to write it.
	return dag.Directory().
		WithNewFile(".sops.yaml", configContent).
		File(".sops.yaml"), nil
}

// yamlQuote returns s as a single-quoted YAML scalar, so regex characters
// such as backslashes and colons are taken literally.
func yamlQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
