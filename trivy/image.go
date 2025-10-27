package main

import (
	"context"
	"encoding/json"
	"fmt"

	"dagger/trivy/internal/dagger"
	"dagger/trivy/report"
)

// TrivyScan performs a security scan on a Docker image using its reference
func (m *Trivy) ScanImage(
	ctx context.Context,
	imageRef string, // Fully qualified image reference (e.g., "ttl.sh/my-repo:1.0.0")
	// +optional
	registryUser *dagger.Secret,
	// +optional
	registryPassword *dagger.Secret,
	// +optional
	// +default="HIGH,CRITICAL"
	severity string,
	// +optional
	// +default="0.64.1"
	trivyVersion string,
) (*dagger.File, error) {
	reportPath := "/tmp/trivy-image-report.json"

	// Configure Trivy container
	trivyContainer := dag.Container().
		From("aquasec/trivy:" + trivyVersion)

	if registryUser != nil && registryPassword != nil { // pragma: allowlist secret
		trivyContainer = trivyContainer.
			WithSecretVariable("TRIVY_USERNAME", registryUser).
			WithSecretVariable("TRIVY_PASSWORD", registryPassword)

		fmt.Println("✅ Trivy credentials configured")
	}

	trivyContainer = trivyContainer.WithExec([]string{
		"sh", "-c", fmt.Sprintf("trivy image --format json --severity %s %s > %s || true", severity, imageRef, reportPath),
	})

	reportFile := trivyContainer.File(reportPath)

	// Read original report
	reportStr, err := reportFile.Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read image scan report: %w", err)
	}

	// Parse vulnerabilities from report
	vulns, err := report.SearchVulnerabilities(ctx, reportStr, "") // "" = all severities
	if err != nil {
		return nil, fmt.Errorf("failed to parse vulnerabilities: %w", err)
	}

	// Enrich original report JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(reportStr), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse original report JSON: %w", err)
	}

	parsed["ParsedSummary"] = vulns

	modifiedJSON, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal modified report: %w", err)
	}

	// Overwrite report in container
	trivyContainer = trivyContainer.WithNewFile(reportPath, string(modifiedJSON))

	return trivyContainer.File(reportPath), nil
}
