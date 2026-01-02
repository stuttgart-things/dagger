package main

import (
	"context"
	"dagger/crossplane/internal/dagger"
	"dagger/crossplane/templates"
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

// InitCustomPackage scaffolds a custom Crossplane package with templates and configuration
// Parameters:
//   - packageName: Name of the package to create (required)
//   - kind: Composite Resource kind (optional, defaults to packageName)
//   - namespace: Kubernetes namespace (optional, defaults to "crossplane-system")
//   - dataYaml: YAML string with additional template data (optional)
// Usage: dagger call init-custom-package --package-name my-pvc --kind ClusterResourcePVC --namespace default --data-yaml '{"storage":"10Gi"}'
func (m *Crossplane) InitCustomPackage(
	ctx context.Context,
	packageName string,
	kind string,
	namespace string,
	dataYaml string,
) (*dagger.Directory, error) {

	if packageName == "" {
		return nil, fmt.Errorf("package name is required")
	}

	// Set defaults
	if namespace == "" {
		namespace = "crossplane-system"
	}

	if kind == "" {
		kind = packageName
	}

	xplane := m.XplaneContainer
	workingDir := "/" + packageName + "/"

	// Build data map with defaults
	data := map[string]interface{}{
		"name":      packageName,
		"kind":      kind,
		"namespace": namespace,
	}

	// Parse optional YAML data if provided
	if dataYaml != "" {
		var customData map[string]interface{}
		err := parseYAMLString(dataYaml, &customData)
		if err != nil {
			fmt.Printf("⚠️  Failed to parse data YAML: %v (continuing with defaults)\n", err)
		} else {
			// Merge custom data
			for k, v := range customData {
				data[k] = v
			}
		}
	}

	fmt.Printf("📦 Scaffolding Crossplane package: %s\n", packageName)
	fmt.Printf("   Kind: %s\n", kind)
	fmt.Printf("   Namespace: %s\n", namespace)
	fmt.Printf("   Data keys: %v\n\n", getMapKeys(data))

	// Create files from templates
	fileCount := 0
	for _, template := range templates.PackageFiles {
		rendered := templates.RenderTemplate(template.Template, data)
		if rendered == "" {
			fmt.Printf("⚠️  Template rendered empty for: %s\n", template.Destination)
			continue
		}

		filePath := workingDir + template.Destination
		fmt.Printf("✓ Creating: %s\n", filePath)

		xplane = xplane.WithNewFile(filePath, rendered)
		fileCount++
	}

	if fileCount == 0 {
		return nil, fmt.Errorf("no template files were created")
	}

	fmt.Printf("\n✅ Successfully created %d files for package: %s\n", fileCount, packageName)

	return xplane.Directory(workingDir), nil
}

// Helper function to parse YAML string
func parseYAMLString(yamlStr string, result interface{}) error {
	// Simple JSON-like YAML parsing (handles basic key:value pairs)
	// For more complex YAML, consider using gopkg.in/yaml.v2
	import "encoding/json"
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

// Init Crossplane Package
func (m *Crossplane) InitPackage(ctx context.Context, name string) *dagger.Directory {

	output := m.XplaneContainer.
		WithExec([]string{"crossplane", "xpkg", "init", name, "configuration-template", "-d", name}).
		WithExec([]string{"ls", "-lta", name}).
		WithExec([]string{"rm", "-rf", name + "/NOTES.txt"})

	return output.Directory(name)
}
