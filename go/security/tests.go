/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

// Package security provides utilities for handling security-related functionality.
package security

import "strings"

// ExtractCoverage extracts coverage information from test output.
func ExtractCoverage(testOutput string) string {
	// Look for a line like "coverage: 75.0% of statements"
	lines := strings.Split(testOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, "coverage:") {
			return strings.TrimSpace(line)
		}
	}
	return "coverage: unknown"
}
