// A generated module for Trivy functions
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
	"strings"
)

type Trivy struct{}

// TrivyReport represents the structure of a Trivy JSON report
type TrivyReport struct {
	Results []struct {
		Vulnerabilities []struct {
			VulnerabilityID string `json:"VulnerabilityID"`
			Severity        string `json:"Severity"`
		} `json:"Vulnerabilities"`
	} `json:"Results"`
}

// func (m *Trivy) ScanRemoteImage(
// 	ctx context.Context,
// 	imageAddress string, // Remote image address (e.g., "ko.local/my-image:latest")
// 	severityFilter string, // Comma-separated list of severities to filter (e.g., "HIGH,CRITICAL")
// ) (string, error) {
// 	// Create a container with Trivy installed
// 	container := dag.Container().
// 		From("aquasec/trivy:latest"). // Use the official Trivy image
// 		WithExec([]string{"trivy", "image", "--severity", severityFilter, imageAddress})

// 	// Capture the scan output
// 	output, err := container.Stdout(ctx)
// 	if err != nil {
// 		return "", fmt.Errorf("error running Trivy scan: %w", err)
// 	}

// 	return output, nil
// }

// SearchVulnerabilities parses the Trivy scan report and filters vulnerabilities by severity
func (m *Trivy) SearchVulnerabilities(
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
