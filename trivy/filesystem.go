package main

import (
	"context"
	"dagger/trivy/internal/dagger"
	"dagger/trivy/report"
	"encoding/json"
	"fmt"
)

func (m *Trivy) ScanFilesystem(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	// +default="HIGH,CRITICAL"
	severity string,
	// +optional
	// +default="0.64.1"
	trivyVersion string,
) (*dagger.File, error) {

	reportPath := "/src/trivy-fs-report.json"

	// Create container with Trivy image and mount source
	container := dag.Container().
		From("aquasec/trivy:"+trivyVersion).
		WithDirectory("/src", src).
		WithWorkdir("/src")

	// Run the scan (ignore exit code using `|| true`)
	container = container.WithExec([]string{
		"sh", "-c", fmt.Sprintf("trivy fs --format json --severity %s /src > %s || true", severity, reportPath),
	})

	// Read the original report
	reportFile := container.File(reportPath)
	reportStr, err := reportFile.Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read filesystem scan report: %w", err)
	}

	// Extract parsed vulnerabilities
	vulns, err := report.SearchVulnerabilities(ctx, reportStr, severity)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vulnerabilities: %w", err)
	}

	// Unmarshal original report JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(reportStr), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse original report JSON: %w", err)
	}

	// Inject parsed summary into the report
	parsed["ParsedSummary"] = vulns

	// Marshal back to JSON
	modifiedJSON, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal modified JSON: %w", err)
	}

	// Overwrite report file inside container with enriched content
	container = container.WithNewFile(reportPath, string(modifiedJSON))

	// Return the modified file
	return container.File(reportPath), nil
}
