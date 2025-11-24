package main

import (
	"context"
	"dagger/dependencies/internal/dagger"
	"strings"
)

// DryRun runs Renovate in dry-run mode against a repository and returns all findings
func (m *Dependencies) RenovateDryRun(
	ctx context.Context,
	// Repository to scan (e.g., "owner/repo" or "stuttgart-things/dagger")
	repo string,
	// GitHub token for authentication (required)
	githubToken *dagger.Secret,
	// Optional: Renovate configuration file from the repository
	// +optional
	config *dagger.File,
) (string, error) {
	// Use the official Renovate Docker image

	renovateContainer := dag.Container().
		From("renovate/renovate:latest")

	renovateContainer = renovateContainer.
		WithEnvVariable("RENOVATE_DRY_RUN", "full").
		WithEnvVariable("LOG_LEVEL", "info"). // Reduced from debug
		WithEnvVariable("RENOVATE_PLATFORM", "github").
		WithEnvVariable("RENOVATE_AUTODISCOVER", "false").
		// Multiple ways to disable OpenTelemetry
		WithEnvVariable("RENOVATE_DISABLE_OPENTELEMETRY", "true").
		WithEnvVariable("OTEL_SDK_DISABLED", "true").
		WithEnvVariable("RENOVATE_OTEL_ENABLED", "false").
		WithEnvVariable("OTEL_EXPORTER_OTLP_ENDPOINT", "").
		WithEnvVariable("OTEL_TRACES_EXPORTER", "none").
		WithEnvVariable("OTEL_METRICS_EXPORTER", "none").
		WithEnvVariable("OTEL_LOGS_EXPORTER", "none").
		WithSecretVariable("GITHUB_COM_TOKEN", githubToken).
		WithSecretVariable("RENOVATE_TOKEN", githubToken)

	if config != nil {
		renovateContainer = renovateContainer.
			WithMountedFile("/usr/src/app/config.json", config).
			WithEnvVariable("RENOVATE_CONFIG_FILE", "/usr/src/app/config.json")
	}

	// Use Sync to wait for the command to complete, ignoring exit code
	// because Renovate may exit with non-zero due to OTLP errors even when successful
	result := renovateContainer.
		WithExec([]string{"renovate", repo}, dagger.ContainerWithExecOpts{
			// Don't fail on non-zero exit code - we'll check the output
			Expect: dagger.ReturnTypeAny,
		})

	// Get both stdout and stderr
	stdout, stdoutErr := result.Stdout(ctx)
	stderr, _ := result.Stderr(ctx)

	// Check if this is just an OTLP error (which we can ignore)
	// Look for "Repository finished" in stdout which indicates success
	if stdoutErr != nil && containsOTLPError(stderr) && strings.Contains(stdout, "Repository finished") {
		// The scan completed successfully, just had OTLP telemetry issues
		// Filter out the OTLP error from the output
		return filterOTLPErrors(stdout), nil
	}

	// Real error occurred
	if stdoutErr != nil {
		if stderr != "" {
			return stdout + "\n\nSTDERR:\n" + stderr, stdoutErr
		}
		return stdout, stdoutErr
	}

	return filterOTLPErrors(stdout), nil
}

// Helper function to check for OTLP errors
func containsOTLPError(text string) bool {
	return strings.Contains(text, "OTLPExporterError") ||
		strings.Contains(text, "unhandledRejection")
}

// Helper function to filter out OTLP errors from output
func filterOTLPErrors(output string) string {
	lines := strings.Split(output, "\n")
	var filtered []string
	skipBlock := false

	for _, line := range lines {
		// Start skipping when we hit the OTLP error
		if strings.Contains(line, "ERROR: unhandledRejection") {
			skipBlock = true
			continue
		}

		// Stop skipping after the error block
		if skipBlock && !strings.HasPrefix(line, " ") && line != "" {
			skipBlock = false
		}

		if !skipBlock {
			filtered = append(filtered, line)
		}
	}

	return strings.Join(filtered, "\n")
}

// Helper to get last N lines
func getLastLines(text string, n int) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= n {
		return text
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}
