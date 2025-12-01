package main

import (
	"context"
	"dagger/kcl/internal/dagger"
	"strings"
)

// Run executes KCL code from a provided directory or OCI source with parameters
// Returns a Dagger file containing the rendered output
func (m *Kcl) Run(
	ctx context.Context,
	// Local source directory (optional if using OCI source)
	// +optional
	source *dagger.Directory,
	// OCI source path (e.g., oci://ghcr.io/stuttgart-things/kcl-flux-instance)
	// +optional
	ociSource string,
	// KCL parameters as comma-separated key=value pairs
	// Example: "name=my-flux,namespace=flux-system,version=2.4"
	// +optional
	parameters string,
	// +optional
	// +default="true"
	formatOutput bool,
	// Entry point file name
	// +optional
	// +default="main.k"
	entrypoint string) (*dagger.File, error) {

	ctr := m.container()

	// Handle OCI source or local source
	if ociSource != "" {
		// Use OCI source directly - kcl run will handle it
		ctr = ctr.WithWorkdir("/work")
	} else if source != nil {
		// Mount local directory
		ctr = ctr.
			WithMountedDirectory("/src", source).
			WithWorkdir("/src")
	} else {
		// Use current working directory
		ctr = ctr.WithWorkdir("/work")
	}

	// Build the kcl run command with --quiet and -o options
	cmd := "kcl run --quiet "

	// Add source (OCI or local entrypoint)
	if ociSource != "" {
		// Normalize OCI source - add oci:// prefix if missing
		if !strings.HasPrefix(ociSource, "oci://") {
			ociSource = "oci://" + ociSource
		}
		cmd += ociSource
	} else {
		cmd += entrypoint
	}

	// Add parameters if provided
	if parameters != "" {
		// Split comma-separated parameters and add each as -D flag
		params := splitParameters(parameters)
		for _, param := range params {
			cmd += " -D " + param
		}
	}

	// Use -o option to write output to file
	cmd += " -o /output.yaml"

	// Execute and write /output.yaml
	ctr = ctr.WithExec([]string{"sh", "-c", cmd})

	// Post-process into clean YAML
	postProcess := `
  cat /output.yaml \
    | grep -v "^items:" \
    | sed 's/^- /---\n/' \
    | sed '1d' \
    | sed 's/^  //' \
    | awk 'NR==1{print "---"} {print}' \
    > /output-processed.yaml
`

	ctr = ctr.WithExec([]string{"sh", "-c", postProcess})

	// Return processed output
	return ctr.File("/output-processed.yaml"), nil
}

// Helper function to split comma-separated parameters
func splitParameters(params string) []string {
	if params == "" {
		return []string{}
	}

	var result []string
	for _, p := range strings.Split(params, ",") {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
