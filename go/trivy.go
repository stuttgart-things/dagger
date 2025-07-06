package main

import (
	"context"
	"dagger/go/internal/dagger"
	"fmt"
	"strings"
)

func (m *Go) TrivyScan(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	// +default="HIGH,CRITICAL"
	severity string, // Severity levels to include (e.g., "HIGH,CRITICAL")
	// +optional
	// +default="0.59.1"
	trivyVersion string,
) (*dagger.File, error) {
	// Create a container with Trivy installed
	container := dag.Container().
		From("aquasec/trivy:"+trivyVersion). // Use the official Trivy image
		WithDirectory("/src", src).
		WithWorkdir("/src")

	// Run Trivy to scan the source folder and write the output to a file
	reportPath := "/src/trivy-report.txt"
	container = container.
		WithExec([]string{"sh", "-c", fmt.Sprintf("trivy fs --severity %s /src > %s || true", severity, reportPath)}) // Ignore exit code

	// Get the Trivy report file
	reportFile := container.File(reportPath)

	// Return the Trivy report file
	return reportFile, nil
}

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

func (m *Go) ScanRemoteImage(
	ctx context.Context,
	imageAddress string, // Remote image address (e.g., "ko.local/my-image:latest")
	severityFilter string, // Comma-separated list of severities to filter (e.g., "HIGH,CRITICAL")
) (string, error) {
	// Create a container with Trivy installed
	container := dag.Container().
		From("aquasec/trivy:latest"). // Use the official Trivy image
		WithExec([]string{"trivy", "image", "--severity", severityFilter, imageAddress})

	// Capture the scan output
	output, err := container.Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("error running Trivy scan: %w", err)
	}

	return output, nil
}

// Lint runs the linter on the provided source code
func (m *Go) ScanTarBallImage(
	ctx context.Context,
	file *dagger.File,
) (*dagger.File, error) {
	scans := []*dagger.TrivyScan{
		dag.Trivy().ImageTarball(file),
	}

	// Grab the report as a file
	reportFile, err := scans[0].Report("json").Sync(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting report: %w", err)
	}

	return reportFile, nil
}
