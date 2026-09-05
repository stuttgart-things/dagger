package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"dagger/trivy/internal/dagger"
	"dagger/trivy/report"
)

// ScanImage scans a published container image and returns Trivy's JSON
// report, enriched with a ParsedSummary listing the findings at the requested
// severities.
//
// Trivy's own exit code is deliberately discarded (`|| true`): it exits 0 on
// findings anyway unless asked otherwise, and swallowing it keeps a scan that
// found something from being confused with a scan that could not run. The
// verdict is taken from the parsed report instead, which is also what makes
// the summary and the gate agree by construction.
//
// A scan that cannot run at all is still an error -- an unreachable registry
// or an image that does not exist leaves no parseable JSON behind, and
// reporting that as "nothing found" would be the worst answer available.
//
// With failOnFinding set this returns an error, so a `dagger call scan-image
// ... export --path report.json` never reaches the export. To both keep the
// report and gate the build, call twice -- once to export, once to gate. The
// second call hits Dagger's cache and costs nothing. Govulncheck in the go
// module has the same shape for the same reason.
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
	// Fail when the scan finds anything at the requested severities. Off by
	// default so that adopting a newer module version cannot turn a green
	// pipeline red on its own; callers that want a gate ask for one.
	// +optional
	// +default=false
	failOnFinding bool,
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

	// Scanned at the requested severities, so everything in the report is a
	// finding; passing severity again would only re-filter what Trivy already
	// filtered, and disagree with it if the two spellings ever drifted.
	vulns, err := report.SearchVulnerabilities(ctx, reportStr, "")
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

	if failOnFinding && len(vulns) > 0 {
		return nil, fmt.Errorf(
			"trivy found %d vulnerabilities at severity %s in %s:\n%s",
			len(vulns), severity, imageRef, strings.Join(vulns, "\n"))
	}

	return trivyContainer.File(reportPath), nil
}
