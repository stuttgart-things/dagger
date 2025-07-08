/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package security

import "strings"

// Helper function to extract coverage from test output
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
