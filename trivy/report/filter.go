package report

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type TrivySecretReport struct {
	Results []struct {
		Target  string `json:"Target"`
		Class   string `json:"Class"`
		Secrets []struct {
			RuleID    string `json:"RuleID"`
			Category  string `json:"Category"`
			Severity  string `json:"Severity"`
			Title     string `json:"Title"`
			StartLine int    `json:"StartLine"`
			EndLine   int    `json:"EndLine"`
			Match     string `json:"Match"`
		} `json:"Secrets"`
	} `json:"Results"`
}

// SearchSecrets returns one line per secret Trivy found, filtered by
// severity. Despite what it was called until now, it reads only Results
// entries of class "secret" and never looks at Vulnerabilities -- see
// SearchVulnerabilities for those.
//
// An empty severityFilter matches nothing here, which is why passing "" and
// expecting "all severities" produced an empty summary.
func SearchSecrets(ctx context.Context, scanOutput string, severityFilter string) ([]string, error) {
	var report TrivySecretReport
	if err := json.Unmarshal([]byte(scanOutput), &report); err != nil {
		return nil, fmt.Errorf("invalid Trivy JSON: %w", err)
	}

	filter := map[string]bool{}
	for _, sev := range strings.Split(severityFilter, ",") {
		filter[strings.TrimSpace(sev)] = true
	}

	var results []string
	for _, result := range report.Results {
		if result.Class != "secret" {
			continue // You could handle "os-pkgs", etc., separately
		}
		for _, secret := range result.Secrets { // pragma: allowlist secret
			if filter[secret.Severity] {
				results = append(results, fmt.Sprintf(
					"[SECRET] %s: %s (in %s at line %d)",
					secret.Severity,
					secret.Title,
					result.Target,
					secret.StartLine,
				))
			}
		}
	}
	return results, nil
}

// pragma: allowlist secret
// ...existing code...

// TrivyVulnerabilityReport is the part of a Trivy JSON report that carries
// package vulnerabilities. Separate from TrivySecretReport because a scan
// emits both kinds under the same Results key and they have nothing in common
// beyond the target.
type TrivyVulnerabilityReport struct {
	Results []struct {
		Target          string `json:"Target"`
		Class           string `json:"Class"`
		Type            string `json:"Type"`
		Vulnerabilities []struct {
			VulnerabilityID  string `json:"VulnerabilityID"`
			PkgName          string `json:"PkgName"`
			InstalledVersion string `json:"InstalledVersion"`
			FixedVersion     string `json:"FixedVersion"`
			Severity         string `json:"Severity"`
			Title            string `json:"Title"`
		} `json:"Vulnerabilities"`
	} `json:"Results"`
}

// SearchVulnerabilities returns one line per package vulnerability in a Trivy
// JSON report, optionally filtered by severity.
//
// It exists because nothing here read the Vulnerabilities key at all. The
// scan functions asked SearchSecrets for a summary, which only ever looks at
// Results entries of class "secret", so ParsedSummary on an image report was
// null however many CVEs the scan had found -- and a caller that wanted to
// gate on findings had nothing to gate on.
//
// An empty severityFilter means every severity, which is the reading a caller
// expects from "no filter". SearchSecrets takes the opposite one and matches
// nothing; that is left as it is rather than changed underneath its caller.
func SearchVulnerabilities(ctx context.Context, scanOutput string, severityFilter string) ([]string, error) {
	var report TrivyVulnerabilityReport
	if err := json.Unmarshal([]byte(scanOutput), &report); err != nil {
		return nil, fmt.Errorf("invalid Trivy JSON: %w", err)
	}

	filter := map[string]bool{}
	for _, sev := range strings.Split(severityFilter, ",") {
		if sev = strings.TrimSpace(sev); sev != "" {
			filter[sev] = true
		}
	}

	var results []string
	for _, result := range report.Results {
		for _, vuln := range result.Vulnerabilities {
			if len(filter) > 0 && !filter[vuln.Severity] {
				continue
			}

			fixed := vuln.FixedVersion
			if fixed == "" {
				fixed = "no fix available"
			}

			results = append(results, fmt.Sprintf(
				"[%s] %s in %s %s (fixed in %s) -- %s",
				vuln.Severity,
				vuln.VulnerabilityID,
				vuln.PkgName,
				vuln.InstalledVersion,
				fixed,
				result.Target,
			))
		}
	}

	return results, nil
}
