package main

import (
	"context"
	"strings"
)

// SearchVulnerabilities parses the Trivy scan report and filters vulnerabilities by severity
func (m *Go) SearchVulnerabilities(
	ctx context.Context,
	scanOutput string, // The scan output as a string
	severityFilter string, // Comma-separated list of severities to filter (e.g., "HIGH,CRITICAL")
) ([]string, error) {
	// Parse the scan output and filter vulnerabilities by severity
	var vulnerabilities []string

	// Example: Split the scan output into lines and filter by severity
	lines := strings.Split(scanOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, severityFilter) {
			vulnerabilities = append(vulnerabilities, line)
		}
	}

	return vulnerabilities, nil
}
