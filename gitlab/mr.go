package main

import (
	"context"
	"dagger/gitlab/internal/dagger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type MergeRequestChanges struct {
	Changes []Change `json:"changes"`
}

type Change struct {
	NewPath string `json:"new_path"`
	OldPath string `json:"old_path"`
}

type MergeRequest struct {
	ID           int    `json:"id"`
	IID          int    `json:"iid"`
	Title        string `json:"title"`
	State        string `json:"state"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
}

// GetMergeRequestID finds the MR ID (IID) by title
func (g *Gitlab) GetMergeRequestID(
	ctx context.Context,
	token dagger.Secret,
	server string,
	projectID string,
	mergeRequestTitle string,
) (string, error) {
	mrsJSON, err := g.ListMergeRequests(ctx, server, token, projectID)
	if err != nil {
		return "", fmt.Errorf("failed to list merge requests: %w", err)
	}

	var mrs []MergeRequest
	if err := json.Unmarshal([]byte(mrsJSON), &mrs); err != nil {
		return "", fmt.Errorf("failed to parse merge requests JSON: %w", err)
	}

	for _, mr := range mrs {
		if mr.Title == mergeRequestTitle {
			return fmt.Sprintf("%d", mr.IID), nil
		}
	}

	return "", fmt.Errorf("merge request %q not found", mergeRequestTitle)
}

// ListMergeRequestChanges lists changed files between MR source and target branch
func (g *Gitlab) ListMergeRequestChanges(
	ctx context.Context,
	server string,
	token dagger.Secret,
	projectID string,
	mergeRequestID string,
) ([]string, error) {
	tok, err := token.Plaintext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret: %w", err)
	}

	url := fmt.Sprintf("https://"+server+"/api/v4/projects/%s/merge_requests/%s/changes", projectID, mergeRequestID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", tok)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status: %d - %s", resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var mrChanges MergeRequestChanges
	if err := json.Unmarshal(body, &mrChanges); err != nil {
		return nil, fmt.Errorf("failed to parse changes JSON: %w", err)
	}

	var files []string
	for _, change := range mrChanges.Changes {
		files = append(files, change.NewPath)
	}

	return files, nil
}

// ListMergeRequests fetches all MRs for a given project
func (g *Gitlab) ListMergeRequests(
	ctx context.Context,
	server string,
	token dagger.Secret,
	projectID string,
) (string, error) {
	tok, err := token.Plaintext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read secret: %w", err)
	}

	url := fmt.Sprintf("https://"+server+"/api/v4/projects/%s/merge_requests", projectID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", tok)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status: %d - %s", resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

func (g *Gitlab) PrintMergeRequestFileChanges(
	ctx context.Context,
	repoURL string,
	server string,
	token dagger.Secret,
	projectID string,
	mergeRequestID string,
	branch string,
) error {

	// 1. Clone the repo
	repoDir, err := g.Clone(ctx, repoURL, token, branch)
	if err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	// 2. Get list of changed files from MR
	changedFiles, err := g.ListMergeRequestChanges(ctx, server, token, projectID, mergeRequestID)
	if err != nil {
		return fmt.Errorf("failed to list changed files: %w", err)
	}

	// 3. For each changed file, read and print its content
	for _, filePath := range changedFiles {
		file := repoDir.File(filePath)

		content, err := file.Contents(ctx)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		fmt.Printf("=== File: %s ===\n%s\n\n", filePath, content)
	}

	return nil
}
