package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"

	"gopkg.in/yaml.v3"
)

// RenderInline renders a Go template from an inline string
func (m *Templating) RenderInline(
	ctx context.Context,
	templateData string,
	// +optional
	variables string,
	// +optional
	strictMode bool,
) (string, error) {

	data := map[string]interface{}{}

	if variables != "" {
		if err := decodeVars(variables, data); err != nil {
			return "", fmt.Errorf("decode variables: %w", err)
		}
	}

	opts := []string{}
	if strictMode {
		opts = append(opts, "missingkey=error")
	}

	t, err := template.
		New("dagger-xplane").
		Option(opts...).
		Parse(templateData)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var output bytes.Buffer
	if err := t.Execute(&output, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return output.String(), nil
}

func decodeVars(raw string, out map[string]interface{}) error {
	// Try JSON first
	if err := json.Unmarshal([]byte(raw), &out); err == nil {
		return nil
	}

	// Fallback to YAML
	return yaml.Unmarshal([]byte(raw), &out)
}
