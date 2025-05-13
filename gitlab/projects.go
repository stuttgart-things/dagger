package main

import (
	"bytes"
	"context"
	"dagger/gitlab/internal/dagger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Project struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
}

// GetProjectID looks up a project ID by its name and group path
func (g *Gitlab) GetProjectID(
	ctx context.Context,
	server string,
	token dagger.Secret,
	projectName string, // e.g., "resource-engines"
	groupPath string, // e.g., "Lab/stuttgart-things/idp"
) (string, error) {
	escapedGroup := url.PathEscape(groupPath)

	projectsJSON, err := g.ListProjects(ctx, server, token, escapedGroup)
	if err != nil {
		return "", fmt.Errorf("failed to list projects: %w", err)
	}

	var projects []Project
	if err := json.Unmarshal([]byte(projectsJSON), &projects); err != nil {
		return "", fmt.Errorf("failed to parse projects JSON: %w", err)
	}

	var matches []Project
	for _, project := range projects {
		if project.Name == projectName {
			matches = append(matches, project)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("project %q not found in group %q", projectName, groupPath)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("multiple projects named %q found in group %q", projectName, groupPath)
	}

	return fmt.Sprintf("%d", matches[0].ID), nil
}

// ListProjects lists all projects in a given group (with pagination)
func (g *Gitlab) ListProjects(
	ctx context.Context,
	server string,
	token dagger.Secret,
	groupPath string, // already escaped: "Lab%2Fstuttgart-things%2Fidp"
) (string, error) {
	tok, err := token.Plaintext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read secret: %w", err)
	}

	var allProjects []byte
	page := 1

	for {
		url := fmt.Sprintf(
			"https://%s/api/v4/groups/%s/projects?per_page=100&page=%d",
			server, groupPath, page,
		)

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

		body = bytes.TrimPrefix(body, []byte("["))
		body = bytes.TrimSuffix(body, []byte("]"))

		if len(body) > 0 {
			if len(allProjects) > 0 {
				allProjects = append(allProjects, ',')
			}
			allProjects = append(allProjects, body...)
		}

		if resp.Header.Get("X-Next-Page") == "" {
			break
		}
		page++
	}

	allProjects = append([]byte("["), allProjects...)
	allProjects = append(allProjects, ']')

	return string(allProjects), nil
}
