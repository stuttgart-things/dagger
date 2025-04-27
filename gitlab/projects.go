package main

import (
	"context"
	"dagger/gitlab/internal/dagger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Project struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
}

// GetProjectID looks up a project ID by its repo path
func (g *Gitlab) GetProjectID(
	ctx context.Context,
	server string,
	token dagger.Secret,
	projectName string,
) (string, error) {
	projectsJSON, err := g.ListProjects(ctx, server, token)
	if err != nil {
		return "", fmt.Errorf("failed to list projects: %w", err)
	}

	fmt.Println("Projects JSON:", projectsJSON)

	var projects []Project
	if err := json.Unmarshal([]byte(projectsJSON), &projects); err != nil {
		return "", fmt.Errorf("failed to parse projects JSON: %w", err)
	}

	for _, project := range projects {
		if project.PathWithNamespace == projectName {
			return fmt.Sprintf("%d", project.ID), nil
		}
	}

	return "", fmt.Errorf("project %q not found", projectName)
}

// ListProjects now takes a Secret
func (g *Gitlab) ListProjects(
	ctx context.Context,
	server string,
	token dagger.Secret,
) (string, error) {
	// Read the secret value
	tok, err := token.Plaintext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read secret: %w", err)
	}

	url := "https://" + server + "/api/v4/projects"

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
