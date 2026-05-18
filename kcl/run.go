package main

import (
	"context"
	"dagger/kcl/internal/dagger"
	"strings"
)

// Helper function to split comma-separated parameters
// Handles array literals like accessModes=["ReadWriteMany"]
// and object literals like extraEnvVars={"KEY":"value"}
func splitParameters(params string) []string {
	if params == "" {
		return []string{}
	}

	var result []string
	var current strings.Builder
	depth := 0 // Track nested brackets and braces

	for i, ch := range params {
		switch ch {
		case '[', '{':
			depth++
			current.WriteRune(ch)
		case ']', '}':
			depth--
			current.WriteRune(ch)
		case ',':
			if depth == 0 {
				// Split here
				trimmed := strings.TrimSpace(current.String())
				if trimmed != "" {
					result = append(result, trimmed)
				}
				current.Reset()
			} else {
				// Keep comma inside brackets/braces
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}

		// Add last parameter
		if i == len(params)-1 {
			trimmed := strings.TrimSpace(current.String())
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
	}

	return result
}

// Run executes KCL code from a provided directory or OCI source with parameters
// Supports three methods of passing parameters (in order of precedence):
// 1. CLI parameters (--parameters flag) - highest priority
// 2. Parameters file (--parametersFile) - middle priority
// 3. Default values in KCL code - lowest priority
//
// Example usage with inline parameters:
//
//	dagger call -m kcl run --ociSource ghcr.io/stuttgart-things/kcl-ansible \
//	  --parameters 'pipelineRunName=run-test,namespace=tekton-ci'
//
// Example usage with parameters file:
//
//	dagger call -m kcl run --ociSource ghcr.io/stuttgart-things/kcl-ansible \
//	  --parametersFile ./params.yaml
//
// Example usage with both (CLI parameters override file values):
//
//	dagger call -m kcl run --ociSource ghcr.io/stuttgart-things/kcl-ansible \
//	  --parametersFile ./params.yaml \
//	  --parameters 'namespace=custom-namespace'
//
// Returns a Dagger file containing the rendered output (YAML by default)
func (m *Kcl) Run(
	ctx context.Context,
	// Local source directory (optional if using OCI source)
	// +optional
	source *dagger.Directory,
	// OCI source path (e.g., oci://ghcr.io/stuttgart-things/kcl-flux-instance)
	// +optional
	ociSource string,
	// KCL parameters as comma-separated key=value pairs
	// For complex JSON structures, you can use JSON syntax
	// Example: "name=my-flux,namespace=flux-system,storage={size:20Mi,mode:ReadWriteOnce}"
	// Takes precedence over parametersFile
	// +optional
	parameters string,
	// YAML/JSON file containing KCL parameters as key-value pairs
	// File format example:
	//   pipelineRunName: run-ansible-test-6
	//   namespace: tekton-ci
	//   ansiblePlaybooks:
	//     - sthings.baseos.setup
	// Parameters from --parameters flag override values from this file
	// +optional
	parametersFile *dagger.File,
	// +optional
	// +default="true"
	formatOutput bool,
	// Output format: yaml or json
	// +optional
	// +default="yaml"
	outputFormat string,
	// Entry point file name
	// +optional
	// +default="main.k"
	entrypoint string,
	// Sub-path inside source to cd into before running kcl. Enables KCL
	// packages with relative path deps pointing outside their own directory
	// (e.g. shared modules in a monorepo). Pass the repo root as source and
	// the sub-package path as subpath.
	// +optional
	subpath string) (*dagger.File, error) {

	ctr := m.container()

	// Mount parameters file if provided
	if parametersFile != nil {
		ctr = ctr.WithMountedFile("/params.yaml", parametersFile)
	}

	// Handle OCI source or local source
	if ociSource != "" {
		// Use OCI source directly - kcl run will handle it
		ctr = ctr.WithWorkdir("/work")
	} else if source != nil {
		// Mount local directory
		ctr = ctr.WithMountedDirectory("/src", source)
		if subpath != "" {
			ctr = ctr.WithWorkdir("/src/" + strings.TrimPrefix(subpath, "/"))
		} else {
			ctr = ctr.WithWorkdir("/src")
		}
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

	// If a parameters file is provided, transform it into a KCL settings file
	// (kcl_options:) and pass it via -Y. This preserves multi-line YAML values
	// as literal-block scalars end-to-end. The previous path flattened values
	// through `jq gsub("\n"; "\\n")` into a `-D` comma list, which turned real
	// newlines into the literal two-character sequence \n; KCL then emitted
	// single-quoted multi-line scalars, and YAML folds those to spaces on
	// read, corrupting any multi-line parameter value (issue #271).
	if parametersFile != nil {
		settingsCmd := `yq eval -o=yaml '{"kcl_options": [to_entries | .[] | {"key": .key, "value": .value}]}' /params.yaml > /settings.yaml`
		ctr = ctr.WithExec([]string{"sh", "-c", settingsCmd})
		cmd += " -Y /settings.yaml"
	}

	// CLI parameters use -D. KCL applies -D after -Y, so any key passed via
	// --parameters overrides the same key from the settings file.
	if parameters != "" {
		params := splitParameters(parameters)
		for _, param := range params {
			// Single-quote the value and escape embedded single quotes.
			quotedParam := "'" + strings.ReplaceAll(param, "'", "'\"'\"'") + "'"
			cmd += " -D " + quotedParam
		}
	}

	// Use -o option to write output to file
	cmd += " -o /output.yaml"

	// Execute and write /output.yaml
	ctr = ctr.WithExec([]string{"sh", "-c", cmd})

	// Post-process into clean YAML if formatOutput is enabled.
	//
	// The post-processor below is designed for KCL code whose top-level value
	// is a list of resources (kcl emits `items:` + indented list). Modern KCL
	// code that uses `manifests.yaml_stream(...)` already emits proper
	// multi-document YAML, and running the post-processor on it corrupts the
	// output (the `sed 's/^  //'` step strips two spaces from every nested
	// line, flattening nested lists like a Dapr Component's `spec.metadata`
	// into top-level keys, which then breaks downstream yq-based tooling).
	//
	// Detect multi-doc output and pass it through unchanged.
	if formatOutput {
		postProcess := `
  if head -c 4 /output.yaml | grep -q '^---' || grep -q '^---[[:space:]]*$' /output.yaml; then
    # kcl already produced multi-document YAML (e.g. via manifests.yaml_stream).
    # Pass through unchanged — the sed pipeline below would corrupt it.
    cp /output.yaml /output-processed.yaml
  else
    cat /output.yaml \
      | grep -v "^items:" \
      | sed 's/^- /---\n/' \
      | sed '1d' \
      | sed 's/^  //' \
      | sed '/^[[:space:]]*$/d' \
      | awk 'NR==1{print "---"} 1' \
      > /output-processed.yaml
  fi
`
		ctr = ctr.WithExec([]string{"sh", "-c", postProcess})
		// Return processed output
		return ctr.File("/output-processed.yaml"), nil
	}

	// Return raw output if formatOutput is disabled
	if outputFormat == "json" {
		// Convert YAML to JSON
		convertCmd := "yq eval -o=json /output.yaml > /output.json"
		ctr = ctr.WithExec([]string{"sh", "-c", convertCmd})
		return ctr.File("/output.json"), nil
	}

	// Return raw YAML
	return ctr.File("/output.yaml"), nil
}
