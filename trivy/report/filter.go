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

func SearchVulnerabilities(ctx context.Context, scanOutput string, severityFilter string) ([]string, error) {
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
		for _, secret := range result.Secrets {
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
