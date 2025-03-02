/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package stats

type WorkflowStats struct {
	Lint struct {
		Duration string   `json:"duration"`
		Findings []string `json:"findings"`
	} `json:"lint"`
	Build struct {
		Duration   string `json:"duration"`
		BinarySize string `json:"binarySize"`
	} `json:"build"`
	Test struct {
		Duration string `json:"duration"`
		Coverage string `json:"coverage"`
	} `json:"test"`
	SecurityScan struct {
		Duration string   `json:"duration"`
		Findings []string `json:"findings"`
	} `json:"securityScan"`
	TrivyScan struct {
		Duration string   `json:"duration"`
		Findings []string `json:"findings"`
	} `json:"trivyScan"`
	TotalDuration string `json:"totalDuration"` // Total duration of the workflow
}
