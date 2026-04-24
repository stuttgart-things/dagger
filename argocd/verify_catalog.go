package main

import (
	"context"
	"dagger/argocd/internal/dagger"
	"fmt"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// VerifyCatalog is the catalog-wide gate for an argocd app-of-apps tree.
//
// It runs three checks in order and fails fast:
//  1. linting.ValidateJsonSchema over the whole tree — every *.schema.json parses.
//  2. Chart.yaml name uniqueness — duplicate names break ArgoCD resolution and
//     are the regression guard for stuttgart-things/argocd#41.
//  3. helm.Kubeconform per chart — rendered manifests match Kubernetes + CRD schemas.
//
// A chart is any directory containing a Chart.yaml. If a chart dir has a sibling
// file at <chart>/ci/default-values.yaml, it is passed as --values-file so charts
// with required values still render.
func (m *Argocd) VerifyCatalog(
	ctx context.Context,
	// Root of the argocd catalog (expected to contain Chart.yaml files in subdirs).
	src *dagger.Directory,
	// Schema locations passed through to kubeconform. Empty uses helm module defaults.
	// +optional
	schemaLocations []string,
	// Registry secret forwarded to helm.Kubeconform for private OCI dependencies.
	// +optional
	registrySecret *dagger.Secret,
	// Basename glob for JSON-schema validation, forwarded to linting.ValidateJsonSchema.
	// +optional
	// +default="**/*.schema.json"
	schemaGlob string,
) (string, error) {
	var out strings.Builder

	// 1. JSON schema parse gate.
	out.WriteString("=== [1/3] linting.ValidateJsonSchema ===\n")
	schemaOut, err := dag.Linting().ValidateJSONSchema(ctx, src, dagger.LintingValidateJSONSchemaOpts{Glob: schemaGlob})
	out.WriteString(schemaOut)
	if err != nil {
		return out.String(), fmt.Errorf("validate-json-schema failed: %w", err)
	}

	// 2. Chart discovery + name-uniqueness check.
	out.WriteString("\n=== [2/3] chart discovery + name uniqueness ===\n")
	charts, err := discoverCharts(ctx, src)
	if err != nil {
		return out.String(), err
	}
	if len(charts) == 0 {
		return out.String(), fmt.Errorf("no Chart.yaml files found under src")
	}
	fmt.Fprintf(&out, "found %d chart(s)\n", len(charts))
	seen := map[string][]string{}
	for _, c := range charts {
		seen[c.name] = append(seen[c.name], c.dir)
		fmt.Fprintf(&out, "  %s  (name=%s)\n", c.dir, c.name)
	}
	var dupes []string
	for name, dirs := range seen {
		if len(dirs) > 1 {
			dupes = append(dupes, fmt.Sprintf("%q used by: %s", name, strings.Join(dirs, ", ")))
		}
	}
	if len(dupes) > 0 {
		sort.Strings(dupes)
		return out.String(), fmt.Errorf("chart name uniqueness violated:\n  %s", strings.Join(dupes, "\n  "))
	}
	out.WriteString("all chart names are unique\n")

	// 3. Per-chart kubeconform.
	out.WriteString("\n=== [3/3] helm.Kubeconform per chart ===\n")
	for _, c := range charts {
		fmt.Fprintf(&out, "\n--- %s ---\n", c.dir)
		chartDir := src.Directory(c.dir)
		opts := dagger.HelmKubeconformOpts{
			SchemaLocations: schemaLocations,
			RegistrySecret:  registrySecret,
		}
		// Use ci/default-values.yaml when present — charts with required values need it.
		ciValues := path.Join(c.dir, "ci", "default-values.yaml")
		if has, _ := pathExists(ctx, src, ciValues); has {
			fmt.Fprintf(&out, "using values-file %s\n", ciValues)
			opts.ValuesFile = src.File(ciValues)
		}
		kcOut, err := dag.Helm().Kubeconform(ctx, chartDir, opts)
		out.WriteString(kcOut)
		if err != nil {
			return out.String(), fmt.Errorf("kubeconform failed for %s: %w", c.dir, err)
		}
	}

	out.WriteString("\n=== VerifyCatalog: OK ===\n")
	return out.String(), nil
}

type chartEntry struct {
	dir  string // directory containing Chart.yaml, relative to src
	name string // name: field from Chart.yaml
}

func discoverCharts(ctx context.Context, src *dagger.Directory) ([]chartEntry, error) {
	matches, err := src.Glob(ctx, "**/Chart.yaml")
	if err != nil {
		return nil, fmt.Errorf("glob Chart.yaml: %w", err)
	}
	sort.Strings(matches)

	charts := make([]chartEntry, 0, len(matches))
	for _, m := range matches {
		body, err := src.File(m).Contents(ctx)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", m, err)
		}
		var parsed struct {
			Name string `yaml:"name"`
		}
		if err := yaml.Unmarshal([]byte(body), &parsed); err != nil {
			return nil, fmt.Errorf("parse %s: %w", m, err)
		}
		if parsed.Name == "" {
			return nil, fmt.Errorf("%s: name field is empty", m)
		}
		charts = append(charts, chartEntry{
			dir:  path.Dir(m),
			name: parsed.Name,
		})
	}
	return charts, nil
}

func pathExists(ctx context.Context, src *dagger.Directory, p string) (bool, error) {
	matches, err := src.Glob(ctx, p)
	if err != nil {
		return false, err
	}
	return len(matches) > 0, nil
}
