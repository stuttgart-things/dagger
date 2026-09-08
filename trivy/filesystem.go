package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"dagger/trivy/internal/dagger"
	"dagger/trivy/report"
)

// ScanFilesystem scans a directory and returns Trivy's JSON report, enriched
// with a ParsedSummary listing the findings at the requested severities.
//
// A filesystem scan turns up two unrelated kinds of finding: package
// vulnerabilities read out of lockfiles, and secrets matched in the files
// themselves. Until now only the secrets were parsed, so a directory whose
// go.sum pinned a vulnerable dependency produced ParsedSummary: null and a
// successful exit -- the report carried the CVE while the summary said
// nothing. Both kinds are collected now. Secrets keep their existing
// "[SECRET] ..." prefix, so the two remain distinguishable in one list and
// ParsedSummary stays the array it always was.
//
// Trivy's own exit code is deliberately discarded (`|| true`): it exits 0 on
// findings anyway unless asked otherwise, and swallowing it keeps a scan that
// found something from being confused with a scan that could not run. The
// verdict is taken from the parsed report instead, which is what makes the
// summary and the gate agree by construction.
//
// A scan that cannot run at all is still an error -- unparseable output leaves
// nothing to reason about, and reporting that as "nothing found" would be the
// worst answer available.
//
// With failOnFinding set this returns an error, so a `dagger call
// scan-filesystem ... export --path report.json` never reaches the export. To
// both keep the report and gate the build, call twice -- once to export, once
// to gate. The second call hits Dagger's cache and costs nothing. ScanImage
// and the go module's Govulncheck have the same shape for the same reason.
func (m *Trivy) ScanFilesystem(
	ctx context.Context,
	src *dagger.Directory,
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

	// Scanned at the requested severities, so everything Trivy kept is a
	// finding; passing severity again would only re-filter what Trivy already
	// filtered, and disagree with it if the two spellings ever drifted.
	vulns, err := report.SearchVulnerabilities(ctx, reportStr, "")
	if err != nil {
		return nil, fmt.Errorf("failed to parse vulnerabilities: %w", err)
	}

	// SearchSecrets reads an empty filter as "match nothing" rather than "match
	// everything", so it has to be given the severities explicitly. Left that
	// way rather than changed underneath its other caller.
	secrets, err := report.SearchSecrets(ctx, reportStr, severity)
	if err != nil {
		return nil, fmt.Errorf("failed to parse secrets: %w", err)
	}

	findings := append(append([]string{}, vulns...), secrets...)

	// Unmarshal original report JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(reportStr), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse original report JSON: %w", err)
	}

	// Inject parsed summary into the report
	parsed["ParsedSummary"] = findings

	// Marshal back to JSON
	modifiedJSON, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal modified JSON: %w", err)
	}

	// Overwrite report file inside container with enriched content
	container = container.WithNewFile(reportPath, string(modifiedJSON))

	if failOnFinding && len(findings) > 0 {
		return nil, fmt.Errorf(
			"trivy found %d findings at severity %s (%d vulnerabilities, %d secrets):\n%s",
			len(findings), severity, len(vulns), len(secrets),
			strings.Join(findings, "\n"))
	}

	// Return the modified file
	return container.File(reportPath), nil
}
