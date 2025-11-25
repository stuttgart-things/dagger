package main

import (
	"context"
	"dagger/templating/internal/dagger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"
)

// RenderFromFile renders Go templates with data loaded from a YAML or JSON file
// templates: comma-separated list of template paths (local in src or HTTPS URLs)
// dataFile: path to YAML or JSON file containing template data (local in src or HTTPS URL, file extension determines format)
// strictMode: if true, fail on missing variables; if false, render as "<no value>" (default: false)
func (m *Templating) RenderFromFile(
	ctx context.Context,
	src *dagger.Directory,
	templates string,
	dataFile string,
	// +optional
	strictMode bool,
) (*dagger.Directory, error) {
	var fileContent string
	var err error

	// Check if dataFile is an HTTPS URL
	if strings.HasPrefix(dataFile, "https://") {
		// Download the data file from URL
		resp, err := http.Get(dataFile)
		if err != nil {
			return nil, fmt.Errorf("failed to download data file from %s: %w", dataFile, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to download data file from %s: status %d", dataFile, resp.StatusCode)
		}

		// Read the content
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read data file from %s: %w", dataFile, err)
		}
		fileContent = string(bodyBytes)
	} else {
		// Read the data file from local directory
		fileContent, err = src.File(dataFile).Contents(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to read data file %s: %w", dataFile, err)
		}
	}

	// Parse the file based on extension
	var data map[string]interface{}
	ext := strings.ToLower(filepath.Ext(dataFile))

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal([]byte(fileContent), &data); err != nil {
			return nil, fmt.Errorf("failed to parse YAML data file %s: %w", dataFile, err)
		}
	case ".json":
		if err := json.Unmarshal([]byte(fileContent), &data); err != nil {
			return nil, fmt.Errorf("failed to parse JSON data file %s: %w", dataFile, err)
		}
	default:
		return nil, fmt.Errorf("unsupported data file format: %s (supported: .yaml, .yml, .json)", ext)
	}

	// Parse template paths
	templatePaths := strings.Split(templates, ",")
	if len(templatePaths) == 0 {
		return nil, fmt.Errorf("no templates provided")
	}

	// Create a container to work with templates
	container := dag.Container().
		From("golang:1.22-alpine").
		WithMountedDirectory("/src", src).
		WithWorkdir("/output")

	// Process each template
	for _, tmplPath := range templatePaths {
		tmplPath = strings.TrimSpace(tmplPath)
		if tmplPath == "" {
			continue
		}

		var tmplContent string
		var err error

		// Check if it's an HTTPS URL
		if strings.HasPrefix(tmplPath, "https://") {
			// Download template from URL
			resp, err := http.Get(tmplPath)
			if err != nil {
				return nil, fmt.Errorf("failed to download template from %s: %w", tmplPath, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("failed to download template from %s: status %d", tmplPath, resp.StatusCode)
			}

			// Read the content
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read template from %s: %w", tmplPath, err)
			}
			tmplContent = string(bodyBytes)
		} else {
			// Read from local directory
			tmplContent, err = src.File(tmplPath).Contents(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to read template %s: %w", tmplPath, err)
			}
		}

		// Verify template syntax and configure missingkey behavior
		tmpl := template.New(tmplPath).Funcs(sprig.TxtFuncMap())
		if strictMode {
			// In strict mode, fail on missing keys
			tmpl = tmpl.Option("missingkey=error")
		} else {
			// Default: render missing keys as "<no value>"
			tmpl = tmpl.Option("missingkey=default")
		}

		tmpl, err = tmpl.Parse(tmplContent)
		if err != nil {
			return nil, fmt.Errorf("template verification failed for %s: %w", tmplPath, err)
		}

		// Render template
		var rendered strings.Builder
		if err := tmpl.Execute(&rendered, data); err != nil {
			return nil, fmt.Errorf("template rendering failed for %s: %w", tmplPath, err)
		}

		// Determine output filename
		outputName := tmplPath
		if strings.HasPrefix(tmplPath, "https://") {
			// Extract filename from URL
			parts := strings.Split(tmplPath, "/")
			outputName = parts[len(parts)-1]
		}
		// Remove .tmpl extension if present
		outputName = strings.TrimSuffix(outputName, ".tmpl")

		// Write rendered content to container
		container = container.WithNewFile(outputName, rendered.String())
	}

	// Return the directory with rendered templates
	return container.Directory("/output"), nil
}
