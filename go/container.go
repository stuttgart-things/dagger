package main

import (
	"context"
	"dagger/go/internal/dagger"
	"fmt"
	"strings"
)

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
) (string, error) {
	srcDir := "/src"

	ko := m.
		GetKoContainer(koVersion).
		WithDirectory(srcDir, src).
		WithWorkdir(srcDir)

	// DEFINE THE APPLICATION BUILD COMMAND W/ KO
	output, err := ko.
		WithEnvVariable("KO_DOCKER_REPO", repo).
		WithSecretVariable(tokenName, token).
		WithExec(
			[]string{"ko", "build", "--push=" + push, buildArg},
		).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("error running ko build: %w", err)
	}

	// Extract the image address from the output
	imageAddress := strings.TrimSpace(output)
	return imageAddress, nil
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
