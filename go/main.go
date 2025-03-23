/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"context"
	"dagger/go/internal/dagger"
	"dagger/go/security"
	"dagger/go/stats"
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

type GoBuildOpts struct {
	GoVersion  string
	Os         string
	Arch       string
	GoMainFile string
	BinName    string
	Ldflags    string // Add this field for linker flags
}

// GetGoLangContainer returns the default image for golang
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
	ldflags string,
	// +optional
	// +default="2.22.1"
	secureGoVersion string,
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
	// +optional
	// +default="HIGH,CRITICAL"
	trivySeverity string, // Severity levels to include (e.g., "HIGH,CRITICAL")
	// +optional
	// +default="0.59.1"
	trivyVersion string,
) (*dagger.File, error) {
	// Create a struct to hold the statistics
	stats := stats.WorkflowStats{}

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

		buildOutput := m.Binary(ctx, src, goVersion, os, arch, goMainFile, binName, ldflags)
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
		coverage := security.ExtractCoverage(testOutput)
		stats.Test.Coverage = coverage
		errChan <- nil
	}()

	// RUN SECURITY SCAN STEP IN A GOROUTINE
	go func() {
		securityScanStart := time.Now()
		reportFile, err := m.SecurityScan(ctx, src, secureGoVersion)
		if err != nil {
			if !securityScanCanFail {
				errChan <- fmt.Errorf("error running security scan: %w", err)
				return
			}
			// If securityScanCanFail is true, log the error but continue
			stats.SecurityScan.Findings = []string{fmt.Sprintf("Security scan failed: %v", err)}
		} else {
			// Read the report file contents
			reportContent, err := reportFile.Contents(ctx)
			if err != nil {
				errChan <- fmt.Errorf("error reading security report: %w", err)
				return
			}
			stats.SecurityScan.Findings = strings.Split(reportContent, "\n") // Split report content into findings
		}
		stats.SecurityScan.Duration = time.Since(securityScanStart).String()
		errChan <- nil
	}()

	// RUN TRIVY SCAN STEP IN A GOROUTINE
	go func() {
		trivyScanStart := time.Now()
		reportFile, err := m.TrivyScan(ctx, src, trivySeverity, trivyVersion)
		if err != nil {
			if !trivyScanCanFail {
				errChan <- fmt.Errorf("error running Trivy scan: %w", err)
				return
			}
			// If trivyScanCanFail is true, log the error but continue
			stats.TrivyScan.Findings = []string{fmt.Sprintf("Trivy scan failed: %v", err)}
		} else {
			// Read the report file contents
			reportContent, err := reportFile.Contents(ctx)
			if err != nil {
				errChan <- fmt.Errorf("error reading Trivy report: %w", err)
				return
			}
			stats.TrivyScan.Findings = strings.Split(reportContent, "\n") // Split report content into findings
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
	// Step 1: Build the image using KoBuild and push it to the remote registry
	imageAddress, err := m.KoBuild(ctx, src, tokenName, token, repo, buildArg, koVersion, push)
	if err != nil {
		return "", fmt.Errorf("error building and pushing image: %w", err)
	}

	// Step 2: Scan the remote image using Trivy
	scanResult, err := m.ScanRemoteImage(ctx, imageAddress, severityFilter)
	if err != nil {
		return "", fmt.Errorf("error scanning remote image: %w", err)
	}

	// Step 3: Parse the Trivy scan report and search for vulnerabilities
	vulnerabilities, err := m.SearchVulnerabilities(ctx, scanResult, severityFilter)
	if err != nil {
		return "", fmt.Errorf("error searching vulnerabilities: %w", err)
	}

	// Step 4: Return the vulnerabilities found
	if len(vulnerabilities) > 0 {
		return fmt.Sprintf("Found vulnerabilities: %v", vulnerabilities), nil
	}

	return "No vulnerabilities found.", nil
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

func (m *Go) RunWorkflowContainerStage(
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
	// Step 1: Build the image with ko and scan it for vulnerabilities
	scanResult, err := m.KoBuildAndScan(ctx, src, tokenName, token, repo, buildArg, koVersion, push, severityFilter)
	if err != nil {
		return "", fmt.Errorf("error during KoBuildAndScan: %w", err)
	}

	// Step 2: Check if vulnerabilities were found
	if strings.Contains(scanResult, "Found vulnerabilities:") {
		return scanResult, nil // Return the vulnerabilities found
	}

	// Step 3: If no vulnerabilities were found, build and push the image to the remote registry
	imageAddress, err := m.KoBuild(ctx, src, tokenName, token, repo, buildArg, koVersion, push)
	if err != nil {
		return "", fmt.Errorf("error building and pushing image: %w", err)
	}

	// Step 4: Return the image address
	return fmt.Sprintf("Image built and pushed successfully: %s", imageAddress), nil
}
