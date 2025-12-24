package main

import "testing"

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name         string
		templateData string
		data         map[string]interface{}
		expected     string
	}{
		{
			name:         "single variable",
			templateData: "Hello {{ .Name }}!",
			data: map[string]interface{}{
				"Name": "World",
			},
			expected: "Hello World!",
		},
		{
			name:         "multiple variables",
			templateData: "{{ .Greeting }}, {{ .Name }}!",
			data: map[string]interface{}{
				"Greeting": "Hi",
				"Name":     "Alice",
			},
			expected: "Hi, Alice!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderTemplate(tt.templateData, tt.data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
