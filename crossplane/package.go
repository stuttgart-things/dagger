package main

import (
	"context"
	"dagger/crossplane/internal/dagger"
	"encoding/json"
	"fmt"
)

// Package Crossplane Package
func (m *Crossplane) Package(ctx context.Context, src *dagger.Directory) *dagger.Directory {

	xplane := m.XplaneContainer.
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"crossplane", "xpkg", "build"})

	buildArtifact, err := xplane.WithExec(
		[]string{"find", "-maxdepth", "1", "-name", "*.xpkg", "-exec", "basename", "{}", ";"}).
		Stdout(ctx)

	if err != nil {
		fmt.Println("ERROR GETTING BUILD ARTIFACT: ", err)
	}

	fmt.Println("BUILD PACKAGE: ", buildArtifact)

	return xplane.Directory("/src")
}

// Helper function to parse YAML string
func parseYAMLString(yamlStr string, result interface{}) error {
	// Simple JSON-like YAML parsing (handles basic key:value pairs)
	// For more complex YAML, consider using gopkg.in/yaml.v2
	return json.Unmarshal([]byte(yamlStr), result)
}

// Helper function to get map keys
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (m *Crossplane) InitPackage(ctx context.Context, name string) *dagger.Directory {
	container := m.GetXplaneContainer(ctx)

	output := container.
		WithExec([]string{"crossplane", "xpkg", "init", name, "configuration-template", "-d", name}).
		WithExec([]string{"ls", "-lta", name}).
		WithExec([]string{"rm", "-rf", name + "/NOTES.txt"})

	return output.Directory(name)
}
