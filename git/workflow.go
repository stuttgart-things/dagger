package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"dagger/git/internal/dagger"
)

// ListGithubWorkflowRuns lists GitHub Actions workflow runs for a repository using the gh CLI.
// Returns the workflow runs as a formatted table.
func (m *Git) ListGithubWorkflowRuns(
	ctx context.Context,
	// Repository in format "owner/repo"
	repository string,
	// GitHub token for authentication
	token *dagger.Secret,
	// Filter by branch
	// +optional
	branch string,
	// Filter by workflow run status (queued, in_progress, completed, requested, waiting)
	// +optional
	status string,
	// Filter by workflow name or filename
	// +optional
	workflow string,
	// Maximum number of runs to return
	// +optional
	// +default=20
	limit int,
) (string, error) {

	ctr, err := m.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	ctr = ctr.WithSecretVariable("GH_TOKEN", token)

	args := []string{
		"gh", "run", "list",
		"--repo", repository,
		"--json", "databaseId,displayTitle,status,conclusion,headBranch,event,workflowName,createdAt,url",
		"--limit", fmt.Sprintf("%d", limit),
	}

	if branch != "" {
		args = append(args, "--branch", branch)
	}

	if status != "" {
		args = append(args, "--status", status)
	}

	if workflow != "" {
		args = append(args, "--workflow", workflow)
	}

	output, err := ctr.
		WithEntrypoint([]string{}).
		WithExec(args).
		Stdout(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to list workflow runs: %w", err)
	}

	var runs []struct {
		DatabaseID   int    `json:"databaseId"`
		DisplayTitle string `json:"displayTitle"`
		Status       string `json:"status"`
		Conclusion   string `json:"conclusion"`
		HeadBranch   string `json:"headBranch"`
		Event        string `json:"event"`
		WorkflowName string `json:"workflowName"`
		CreatedAt    string `json:"createdAt"`
		URL          string `json:"url"`
	}

	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &runs); err != nil {
		return strings.TrimSpace(output), nil
	}

	if len(runs) == 0 {
		return "No workflow runs found.", nil
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tWORKFLOW\tSTATUS\tCONCLUSION\tBRANCH\tEVENT\tTITLE\tCREATED")

	for _, r := range runs {
		conclusion := r.Conclusion
		if conclusion == "" {
			conclusion = "-"
		}
		createdAt := r.CreatedAt
		if len(createdAt) >= 19 {
			createdAt = createdAt[:19]
		}
		title := r.DisplayTitle
		if len(title) > 50 {
			title = title[:47] + "..."
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			r.DatabaseID, r.WorkflowName, r.Status, conclusion, r.HeadBranch, r.Event, title, createdAt)
	}

	w.Flush()
	return buf.String(), nil
}

// WaitForGithubWorkflowRun waits for a GitHub Actions workflow run to complete
// using gh run watch, then returns the final run state as JSON.
func (m *Git) WaitForGithubWorkflowRun(
	ctx context.Context,
	// Repository in format "owner/repo"
	repository string,
	// Workflow run ID to wait for
	runID string,
	// GitHub token for authentication
	token *dagger.Secret,
) (string, error) {

	ctr, err := m.container(ctx)
	if err != nil {
		return "", fmt.Errorf("container init failed: %w", err)
	}

	ctr = ctr.WithSecretVariable("GH_TOKEN", token)

	// Wait for completion, then fetch final state as JSON
	waitCmd := fmt.Sprintf(
		`gh run watch %s --repo %s 2>&1 || true; gh run view %s --repo %s --json databaseId,displayTitle,status,conclusion,headBranch,event,workflowName,url`,
		runID, repository, runID, repository,
	)

	output, err := ctr.
		WithEntrypoint([]string{}).
		WithExec([]string{"sh", "-c", waitCmd}).
		Stdout(ctx)

	if err != nil {
		return "", fmt.Errorf("failed waiting for workflow run: %w", err)
	}

	return strings.TrimSpace(output), nil
}
