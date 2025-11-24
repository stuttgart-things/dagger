package main

import (
	"context"
	"dagger/templating/internal/dagger"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"
)

// RenderTemplates renders Go templates with provided variables
// templates: comma-separated list of template paths (local in src or HTTPS URLs)
// variables: comma-separated list of key=value pairs (e.g., "name=John,age=30")
// variablesFile: optional YAML file with variables (variables parameter has higher priority)
// strictMode: if true, fail on missing variables; if false, render as "<no value>" (default: false)
func (m *Templating) Render(
	ctx context.Context,
	src *dagger.Directory,
	templates string,
	// +optional
	variables string,
	// +optional
	variablesFile string,
	// +optional
	strictMode bool,
) (*dagger.Directory, error) {
	// Initialize vars map
	vars := make(map[string]interface{})

	// Load variables from YAML file first (lower priority)
	if variablesFile != "" {
		yamlContent, err := src.File(variablesFile).Contents(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to read variables file %s: %w", variablesFile, err)
		}

		var yamlVars map[string]interface{}
		if err := yaml.Unmarshal([]byte(yamlContent), &yamlVars); err != nil {
			return nil, fmt.Errorf("failed to parse YAML variables file %s: %w", variablesFile, err)
		}

		// Copy YAML variables to vars map
		for k, v := range yamlVars {
			vars[k] = v
		}
	}

	// Parse and override with command-line variables (higher priority)
	if variables != "" {
		for _, pair := range strings.Split(variables, ",") {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid variable format: %s (expected key=value)", pair)
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Try to convert string value to appropriate type
			var typedValue interface{}
			switch value {
			case "true":
				typedValue = true
			case "false":
				typedValue = false
			default:
				// Keep as string
				typedValue = value
			}

			// Override YAML variables if they exist
			vars[key] = typedValue
		}
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
		if err := tmpl.Execute(&rendered, vars); err != nil {
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
