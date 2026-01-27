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

// Helper function to parse parameters string into a map
func parseParametersToMap(params string) map[string]string {
	result := make(map[string]string)
	if params == "" {
		return result
	}

	parameters := splitParameters(params)
	for _, param := range parameters {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result
}

// Helper function to convert parameter map back to comma-separated string
func mapToParameterString(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}

	var parts []string
	for k, v := range params {
		parts = append(parts, k+"="+v)
	}

	return strings.Join(parts, ",")
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
	entrypoint string) (*dagger.File, error) {

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

	// Merge parameters from file and CLI (CLI takes precedence)
	mergedParams := ""
	if parametersFile != nil {
		// Read parameters from file and convert to comma-separated format
		// Use tostring for values and gsub to escape newlines in multiline strings
		readParamsCmd := `yq eval -o=json /params.yaml | jq -r 'to_entries | map(.key + "=" + ((.value | tostring) | gsub("\n"; "\\n"))) | join(",")'`
		mergedParams, _ = ctr.WithExec([]string{"sh", "-c", readParamsCmd}).Stdout(ctx)
		mergedParams = strings.TrimSpace(mergedParams)
	}

	// Override with CLI parameters if provided
	if parameters != "" {
		if mergedParams != "" {
			// Merge: parse both, CLI params override file params
			fileParams := parseParametersToMap(mergedParams)
			cliParams := parseParametersToMap(parameters)

			// Merge maps (CLI overwrites file)
			for k, v := range cliParams {
				fileParams[k] = v
			}

			// Convert back to comma-separated string
			mergedParams = mapToParameterString(fileParams)
		} else {
			mergedParams = parameters
		}
	}

	// Add parameters if we have any
	if mergedParams != "" {
		// Split comma-separated parameters and add each as -D flag
		params := splitParameters(mergedParams)
		for _, param := range params {
			// Properly quote parameters to preserve special characters
			// Use single quotes to protect the value, but handle single quotes in the value
			quotedParam := "'" + strings.ReplaceAll(param, "'", "'\"'\"'") + "'"
			cmd += " -D " + quotedParam
		}
	}

	// Use -o option to write output to file
	cmd += " -o /output.yaml"

	// Execute and write /output.yaml
	ctr = ctr.WithExec([]string{"sh", "-c", cmd})

	// Post-process into clean YAML if formatOutput is enabled
	if formatOutput {
		postProcess := `
  cat /output.yaml \
    | grep -v "^items:" \
    | sed 's/^- /---\n/' \
    | sed '1d' \
    | sed 's/^  //' \
    | sed '/^[[:space:]]*$/d' \
    | awk 'NR==1{print "---"} NR>1' \
    > /output-processed.yaml
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
