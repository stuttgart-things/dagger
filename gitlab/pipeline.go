package main

import (
	"context"
	"fmt"
	"strings"

	"dagger/gitlab/internal/dagger"
)

// ListGitlabPipelines lists CI/CD pipelines for a GitLab project using the glab CLI.
// Returns the pipeline list as JSON.
func (g *Gitlab) ListGitlabPipelines(
	ctx context.Context,
	// GitLab project path in format "group/project" or "group/subgroup/project"
	repository string,
	// GitLab token for authentication
	token *dagger.Secret,
	// GitLab host (e.g., "gitlab.com" or custom instance)
	// +optional
	// +default="gitlab.com"
	gitlabHost string,
	// Filter by branch or ref
	// +optional
	ref string,
	// Filter by pipeline status (running, pending, success, failed, canceled, skipped, created, manual)
	// +optional
	status string,
) (string, error) {

	ctr, err := g.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	ctr = ctr.
		WithSecretVariable("GITLAB_TOKEN", token).
		WithEnvVariable("GITLAB_HOST", gitlabHost)

	args := []string{"glab", "ci", "list", "-R", repository, "--output", "json"}

	if ref != "" {
		args = append(args, "--ref", ref)
	}

	if status != "" {
		args = append(args, "--status", status)
	}

	output, err := ctr.
		WithEntrypoint([]string{}).
		WithExec(args).
		Stdout(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to list pipelines: %w", err)
	}

	return strings.TrimSpace(output), nil
}

// WaitForGitlabPipeline waits for a GitLab pipeline to reach a terminal state
// (success, failed, canceled, or skipped) by polling the GitLab API.
// Returns the full pipeline JSON response once the pipeline completes.
func (g *Gitlab) WaitForGitlabPipeline(
	ctx context.Context,
	// GitLab project path in format "group/project" or "group/subgroup/project"
	repository string,
	// Pipeline ID to wait for
	pipelineID string,
	// GitLab token for authentication
	token *dagger.Secret,
	// GitLab host (e.g., "gitlab.com" or custom instance)
	// +optional
	// +default="gitlab.com"
	gitlabHost string,
	// Polling interval in seconds
	// +optional
	// +default=10
	interval int,
	// Timeout in seconds
	// +optional
	// +default=600
	timeout int,
) (string, error) {

	ctr, err := g.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	ctr = ctr.
		WithSecretVariable("GITLAB_TOKEN", token).
		WithEnvVariable("GITLAB_HOST", gitlabHost)

	// URL-encode the repository path for the API
	encodedRepo := strings.ReplaceAll(repository, "/", "%%2F")

	waitScript := fmt.Sprintf(`
END_TIME=$(($(date +%%s) + %d))
while true; do
  RESPONSE=$(glab api "projects/%s/pipelines/%s" --hostname "%s" 2>&1)
  STATUS=$(echo "$RESPONSE" | jq -r '.status // empty')
  if [ -z "$STATUS" ]; then
    echo "error: failed to get pipeline status: $RESPONSE" >&2
    exit 1
  fi
  case "$STATUS" in
    success|failed|canceled|skipped|manual)
      echo "$RESPONSE"
      exit 0
      ;;
  esac
  if [ $(date +%%s) -ge $END_TIME ]; then
    echo "timeout: pipeline still in status '$STATUS' after %d seconds" >&2
    exit 1
  fi
  sleep %d
done`, timeout, encodedRepo, pipelineID, gitlabHost, timeout, interval)

	output, err := ctr.
		WithEntrypoint([]string{}).
		WithExec([]string{"sh", "-c", waitScript}).
		Stdout(ctx)

	if err != nil {
		return "", fmt.Errorf("failed waiting for pipeline: %w", err)
	}

	return strings.TrimSpace(output), nil
}
