/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

// https://github.com/dagger/dagger/pull/5833/files#diff-42807a87b4d8f4c8adb3861609de1a2a6a6158cf11b00b9b1b342c0a23f1bc03

package main

import (
	"context"
	"dagger/go/internal/dagger"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Go struct {
	Src             *dagger.Directory
	GoLangContainer *dagger.Container
	KoContainer     *dagger.Container
}

type WorkflowResult struct {
	LintOutput  string            // Output from the lint step
	BuildOutput *dagger.Directory // Output from the build step
}

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

// GetGoLangContainer return the default image for golang
func (m *Go) GetGoLangContainer(goVersion string) *dagger.Container {
	return dag.Container().
		From("golang:" + goVersion)
}

func (m *Go) GetKoContainer(
	// +optional
	// +default="v0.17.1"
	koVersion string,
) *dagger.Container {
	return dag.Container().
		From("ghcr.io/ko-build/ko:" + koVersion)
}

func New(
	// golang container
	// It need contain golang
	// +optional
	goLangContainer *dagger.Container,
	// +optional
	koContainer *dagger.Container,
	// +optional
	// +default="1.23.6"
	goLangVersion string,
	// +defaultPath="/"
	src *dagger.Directory,

) *Go {
	golang := &Go{}

	if goLangContainer != nil {
		golang.GoLangContainer = goLangContainer
	} else {
		golang.GoLangContainer = golang.GetGoLangContainer(goLangVersion)
	}

	if koContainer != nil {
		golang.KoContainer = koContainer
	} else {
		golang.KoContainer = golang.GetKoContainer("v0.17.1")
	}

	golang.Src = src

	return golang
}

// Lint runs the linter on the provided source code
func (m *Go) Lint(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	// +default="500s"
	timeout string,
) *dagger.Container {

	golangciLintRunOpts := dagger.GolangciLintRunOpts{
		Timeout: timeout,
	}

	return dag.GolangciLint().Run(src, golangciLintRunOpts)
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

func (m *Go) KoBuildAndScan(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	// +default="GITHUB_TOKEN"
	tokenName string,
	token *dagger.Secret,
	// +optional
	// +default="ko.local"
	repo string,
	// +optional
	// +default="."
	buildArg string,
	// +optional
	// +default="v0.17.1"
	koVersion string,
	// +optional
	// +default="true"
	push string,
	// +optional
	// +default="HIGH,CRITICAL"
	severityFilter string, // Comma-separated list of severities to filter (e.g., "HIGH,CRITICAL")
) (string, error) {
	// Step 1: Build the image using KoBuild
	buildOutput := m.KoBuild(ctx, src, tokenName, token, repo, buildArg, koVersion, push)

	// Step 2: Get the tarball from the build output
	tarball := buildOutput.File("test.tar") // Adjust the path if necessary

	// Step 3: Scan the tarball using Trivy
	scanResult, err := m.ScanTarBallImage(ctx, tarball)
	if err != nil {
		return "", fmt.Errorf("error scanning image: %w", err)
	}

	// Step 4: Parse the Trivy scan report and search for vulnerabilities
	vulnerabilities, err := m.SearchVulnerabilities(ctx, scanResult, severityFilter)
	if err != nil {
		return "", fmt.Errorf("error searching vulnerabilities: %w", err)
	}

	// Step 5: Return the vulnerabilities found
	if len(vulnerabilities) > 0 {
		return fmt.Sprintf("Found vulnerabilities: %v", vulnerabilities), nil
	}

	return "No vulnerabilities found.", nil
}

// SearchVulnerabilities parses the Trivy scan report and filters vulnerabilities by severity
func (m *Go) SearchVulnerabilities(
	ctx context.Context,
	scanResult *dagger.File,
	severityFilter string, // Comma-separated list of severities (e.g., "HIGH,CRITICAL")
) ([]string, error) {
	// Step 1: Read the scan result file
	scanJSON, err := scanResult.Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("error reading scan result: %w", err)
	}

	// Step 2: Parse the JSON report
	var report TrivyReport
	if err := json.Unmarshal([]byte(scanJSON), &report); err != nil {
		return nil, fmt.Errorf("error parsing scan report: %w", err)
	}

	// Step 3: Filter vulnerabilities by severity
	severities := strings.Split(severityFilter, ",")
	var vulnerabilities []string

	for _, result := range report.Results {
		for _, vulnerability := range result.Vulnerabilities {
			for _, severity := range severities {
				if strings.EqualFold(vulnerability.Severity, severity) {
					vulnerabilities = append(vulnerabilities, vulnerability.VulnerabilityID)
				}
			}
		}
	}

	return vulnerabilities, nil
}

// TrivyReport represents the structure of a Trivy JSON report
type TrivyReport struct {
	Results []struct {
		Vulnerabilities []struct {
			VulnerabilityID string `json:"VulnerabilityID"`
			Severity        string `json:"Severity"`
		} `json:"Vulnerabilities"`
	} `json:"Results"`
}

func (m *Go) RunWorkflowEntryStage(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	// +default="500s"
	lintTimeout string,
	// +optional
	// +default="1.23.6"
	goVersion string,
	// +optional
	// +default="linux"
	os string,
	// +optional
	// +default="amd64"
	arch string,
	// +optional
	// +default="main.go"
	goMainFile string,
	// +optional
	// +default="main"
	binName string,
	// +optional
	// +default="false"
	lintCanFail bool, // If true, linting can fail without stopping the workflow
	// +optional
	// +default="./..."
	testArg string, // Arguments for `go test`
	// +optional
	// +default="false"
	securityScanCanFail bool, // If true, security scan can fail without stopping the workflow
	// +optional
	// +default="false"
	trivyScanCanFail bool, // If true, Trivy scan can fail without stopping the workflow
) (*dagger.File, error) {
	// Create a struct to hold the statistics
	stats := WorkflowStats{}

	// Start timing the workflow
	startTime := time.Now()

	// Create a channel to collect errors from goroutines
	errChan := make(chan error, 5) // Buffer size of 5 for lint, build, test, security scan, and Trivy scan

	// Run Lint step in a goroutine
	go func() {
		lintStart := time.Now()
		lintOutput, err := m.Lint(ctx, src, lintTimeout).Stdout(ctx)
		if err != nil {
			if !lintCanFail {
				errChan <- fmt.Errorf("error running lint: %w", err)
				return
			}
			// If lintCanFail is true, log the error but continue
			stats.Lint.Findings = []string{fmt.Sprintf("Linting failed: %v", err)}
		} else {
			stats.Lint.Findings = strings.Split(lintOutput, "\n") // Split lint output into findings
		}
		stats.Lint.Duration = time.Since(lintStart).String()
		errChan <- nil
	}()

	// Run Build step in a goroutine
	go func() {
		buildStart := time.Now()
		buildOutput := m.Build(ctx, goVersion, os, arch, goMainFile, binName, src)
		stats.Build.Duration = time.Since(buildStart).String()

		// Calculate binary size
		binaryPath := binName
		binarySize, err := buildOutput.File(binaryPath).Size(ctx)
		if err != nil {
			errChan <- fmt.Errorf("error getting binary size: %w", err)
			return
		}
		stats.Build.BinarySize = fmt.Sprintf("%d bytes", binarySize)
		errChan <- nil
	}()

	// Run Test step in a goroutine
	go func() {
		testStart := time.Now()
		testOutput, err := m.Test(ctx, src, goVersion, testArg)
		if err != nil {
			errChan <- fmt.Errorf("error running tests: %w", err)
			return
		}
		stats.Test.Duration = time.Since(testStart).String()

		// Extract coverage from test output
		coverage := extractCoverage(testOutput)
		stats.Test.Coverage = coverage
		errChan <- nil
	}()

	// Run Security Scan step in a goroutine
	go func() {
		securityScanStart := time.Now()
		securityScanOutput, err := m.SecurityScan(ctx, src)
		if err != nil {
			if !securityScanCanFail {
				errChan <- fmt.Errorf("error running security scan: %w", err)
				return
			}
			// If securityScanCanFail is true, log the error but continue
			stats.SecurityScan.Findings = []string{fmt.Sprintf("Security scan failed: %v", err)}
		} else {
			stats.SecurityScan.Findings = strings.Split(securityScanOutput, "\n") // Split security scan output into findings
		}
		stats.SecurityScan.Duration = time.Since(securityScanStart).String()
		errChan <- nil
	}()

	// Run Trivy Scan step in a goroutine
	go func() {
		trivyScanStart := time.Now()
		trivyScanOutput, err := m.TrivyScan(ctx, src)
		if err != nil {
			if !trivyScanCanFail {
				errChan <- fmt.Errorf("error running Trivy scan: %w", err)
				return
			}
			// If trivyScanCanFail is true, log the error but continue
			stats.TrivyScan.Findings = []string{fmt.Sprintf("Trivy scan failed: %v", err)}
		} else {
			stats.TrivyScan.Findings = strings.Split(trivyScanOutput, "\n") // Split Trivy scan output into findings
		}
		stats.TrivyScan.Duration = time.Since(trivyScanStart).String()
		errChan <- nil
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		if err := <-errChan; err != nil {
			return nil, err
		}
	}

	// Track total workflow duration
	stats.TotalDuration = time.Since(startTime).String()

	// Generate JSON file with statistics
	statsJSON, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error generating stats JSON: %w", err)
	}

	// Write JSON to a file in the container
	statsFile := dag.Directory().
		WithNewFile("workflow-stats.json", string(statsJSON)).
		File("workflow-stats.json")

	// Return the stats file
	return statsFile, nil
}

// Helper function to extract coverage from test output
func extractCoverage(testOutput string) string {
	// Look for a line like "coverage: 75.0% of statements"
	lines := strings.Split(testOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, "coverage:") {
			return strings.TrimSpace(line)
		}
	}
	return "coverage: unknown"
}

// Returns lines that match a pattern in the files of the provided Directory
func (m *Go) Build(
	ctx context.Context,
	// +optional
	// +default="1.23.6"
	goVersion string,
	// +optional
	// +default="linux"
	os string,
	// +optional
	// +default="amd64"
	arch string,
	// +optional
	// +default="main.go"
	goMainFile string,
	// +optional
	// +default="main"
	binName string,
	src *dagger.Directory) *dagger.Directory {

	// MOUNT CLONED REPOSITORY INTO `GOLANG` IMAGE

	PrintDirectoryInfo(ctx, src)

	golang := m.
		GetGoLangContainer(goVersion).
		WithDirectory("/src", src).
		WithWorkdir("/src")

	fmt.Println("DIR", src)
	// DEFINE THE APPLICATION BUILD COMMAND
	path := "build/"
	golang = golang.WithExec([]string{
		"env",
		"GOOS=" + os,
		"GOARCH=" + arch,
		"go",
		"build",
		"-o",
		path + "/" + binName,
		goMainFile,
	})

	// GET REFERENCE TO BUILD OUTPUT DIRECTORY IN CONTAINER
	outputDir := golang.Directory(path)

	return outputDir
}

func PrintDirectoryInfo(ctx context.Context, src *dagger.Directory) {
	if src == nil {
		fmt.Println("Directory is nil")
		return
	}

	id, err := src.ID(ctx)
	if err != nil {
		fmt.Println("Error getting directory ID:", err)
		return
	}
	fmt.Println("Dagger Directory ID:", id)

	// List files inside the directory
	entries, err := src.Entries(ctx)
	if err != nil {
		fmt.Println("Error retrieving directory entries:", err)
		return
	}

	fmt.Println("Directory contains:", entries)
}

// Returns lines that match a pattern in the files of the provided Directory
func (m *Go) KoBuild(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	// +default="GITHUB_TOKEN"
	tokenName string,
	token *dagger.Secret,
	// +optional
	// +default="ko.local"
	repo string,
	// +optional
	// +default="."
	buildArg string,
	// +optional
	// +default="v0.17.1"
	koVersion string,
	// +optional
	// +default="true"
	push string,
) *dagger.Directory {

	srcDir := "/src"

	ko := m.
		GetKoContainer(koVersion).
		WithDirectory(srcDir, src).
		WithWorkdir(srcDir)

	// DEFINE THE APPLICATION BUILD COMMAND W/ KO
	ko = ko.
		WithEnvVariable("KO_DOCKER_REPO", repo).
		WithSecretVariable(tokenName, token).
		WithExec(
			[]string{"ko", "build", "--push=" + push, buildArg},
		)

	outputDir := ko.Directory(srcDir)

	return outputDir
}

func (m *Go) Test(
	ctx context.Context,
	src *dagger.Directory,
	goVersion string, // Go version to use for testing
	// +optional
	// +default="./..."
	testArg string, // Arguments for `go test`
) (string, error) {
	// Create a container with the specified Go version
	container := dag.Container().
		From(fmt.Sprintf("golang:%s", goVersion)). // Use the specified Go version
		WithDirectory("/src", src).
		WithWorkdir("/src")

	// Run Go tests with coverage
	output, err := container.
		WithExec([]string{"go", "test", "-cover", testArg}).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("error running tests: %w", err)
	}

	return output, nil
}

func (m *Go) TrivyScan(
	ctx context.Context,
	src *dagger.Directory,
) (string, error) {
	// Create a container with Trivy installed
	container := dag.Container().
		From("aquasec/trivy:latest"). // Use the official Trivy image
		WithDirectory("/src", src).
		WithWorkdir("/src")

	// Run Trivy to scan the source folder
	output, err := container.
		WithExec([]string{"trivy", "fs", "--severity", "HIGH,CRITICAL", "/src"}).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("error running Trivy: %w", err)
	}

	return output, nil
}

func (m *Go) SecurityScan(
	ctx context.Context,
	src *dagger.Directory,
) (string, error) {
	// Create a container with gosec installed
	container := dag.Container().
		From("securego/gosec:latest"). // Use the official gosec image
		WithDirectory("/src", src).
		WithWorkdir("/src")

	// Run gosec to scan the source code
	output, err := container.
		WithExec([]string{"gosec", "./..."}).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("error running gosec: %w", err)
	}

	return output, nil
}
