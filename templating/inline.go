package main

import (
	"bytes"
	"text/template"

	"gopkg.in/yaml.v3"
)

func RenderTemplate(templateData string, data map[string]interface{}) (string, error) {
	t, err := template.New("template").Parse(templateData)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	if err := t.Execute(&output, data); err != nil {
		return "", err
	}

	return output.String(), nil
}

func RenderVarsToVarsFile(data map[string]interface{}) (string, error) {
	out, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
