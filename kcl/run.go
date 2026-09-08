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
	// OCI source path (e.g., oci://ghcr.io/stuttgart-things/kcl-flux-instance).
	// A version may be given as ?tag=0.3.0 or as an inline :0.3.0 -- the
	// inline form is rewritten to the query form, because kcl reads the
	// version only from a tag query parameter and would otherwise resolve to
	// the newest published tag.
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

	// Add source (OCI or local entrypoint). The source is single-quoted
	// because normalizeOciSource emits a `?`, which is a shell glob character:
	// an unmatched glob happens to pass through unchanged in this container's
	// sh, but that is luck, not a guarantee.
	if ociSource != "" {
		cmd += shellQuote(normalizeOciSource(ociSource))
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
	// Detect output that is already a rendered manifest and pass it through
	// unchanged.
	//
	// The leading-`---` test alone is not enough: a module that renders a
	// single resource emits a document that starts directly with
	// `apiVersion:` and contains no `---` at all. That fell through to the
	// sed pipeline below, where `sed '1d'` removed the apiVersion line and
	// `sed 's/^  //'` flattened metadata and spec into top-level keys -- the
	// manifest still looked plausible but no longer applied. A top-level
	// `apiVersion:` at column 0 is only possible when the document is already
	// flat, because in the items-list shape every resource key is indented
	// underneath `- `.
	if formatOutput {
		postProcess := `
  if head -c 4 /output.yaml | grep -q '^---' || grep -q '^---[[:space:]]*$' /output.yaml || grep -q '^apiVersion:' /output.yaml; then
    # kcl already produced a rendered manifest (multi-document via
    # manifests.yaml_stream, or a single document starting with apiVersion).
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

// normalizeOciSource returns src as an oci:// reference, rewriting an inline
// `:tag` into the `?tag=` query form that kcl actually honours.
//
// kcl parses an OCI source with url.Parse and reads the version only from a
// `tag` query parameter (kpm, pkg/downloader/source.go, Oci.FromString). An
// inline `:0.2.0` therefore stays part of the repository path and the
// reference silently resolves to the newest published tag -- kcl even logs
// "the latest version '0.3.0' will be downloaded" while the caller's string
// still reads as pinned. Measured against ghcr.io/stuttgart-things/harvester-vm
// on kcl 0.12.4. See stuttgart-things/kcl#231 and #237.
//
// Deliberately left alone:
//   - a source that already carries a query, or a digest reference (@sha256:...)
//   - a reference with no repository path at all, which cannot be a module
//   - the colon in a registry port, which is why only the final path segment
//     is examined: localhost:5000/foo/bar keeps its port
func normalizeOciSource(src string) string {
	const scheme = "oci://"

	ref := strings.TrimPrefix(src, scheme)
	if strings.ContainsAny(ref, "?@") {
		return scheme + ref
	}

	slash := strings.LastIndex(ref, "/")
	if slash < 0 {
		return scheme + ref
	}

	name := ref[slash+1:]
	colon := strings.LastIndex(name, ":")
	if colon <= 0 || colon == len(name)-1 {
		return scheme + ref
	}

	return scheme + ref[:slash+1] + name[:colon] + "?tag=" + name[colon+1:]
}

// shellQuote wraps s in single quotes for `sh -c`, escaping any it contains.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}
