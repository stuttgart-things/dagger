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

// Probe destinations used by the argocd#41 regression check. Hostnames are
// .invalid (RFC 2606) so they can never be confused with a real cluster.
const (
	probeServerA = "https://cluster-a.verify-catalog.invalid"
	probeServerB = "https://cluster-b.verify-catalog.invalid"
)

// VerifyCatalog is the catalog-wide gate for an argocd app-of-apps tree.
//
// Pipeline (fail-fast):
//  1. linting.ValidateJsonSchema — every *.schema.json under src parses.
//  2. Chart discovery + Chart.yaml name uniqueness pre-flight.
//  3. Per chart: helm.Lint + helm.Kubeconform (with ci/default-values.yaml
//     when present so charts with required values still render).
//  4. argocd#41 regression guard: render each chart twice with different
//     destination.server values and assert the rendered Application
//     metadata.name set differs. Charts that emit no Application resources
//     are skipped.
//
// A chart is any directory containing a Chart.yaml.
func (m *Argocd) VerifyCatalog(
	ctx context.Context,
	// Root of the argocd catalog (expected to contain Chart.yaml files in subdirs).
	src *dagger.Directory,
	// Schema locations passed through to kubeconform. Empty uses helm module defaults.
	// +optional
	schemaLocations []string,
	// Registry secret forwarded to helm functions for private OCI dependencies.
	// +optional
	registrySecret *dagger.Secret,
	// Basename glob for JSON-schema validation, forwarded to linting.ValidateJsonSchema.
	// +optional
	// +default="**/*.schema.json"
	schemaGlob string,
) (string, error) {
	var out strings.Builder

	// 1. JSON schema parse gate.
	out.WriteString("=== [1/4] linting.ValidateJsonSchema ===\n")
	schemaOut, err := dag.Linting().ValidateJSONSchema(ctx, src, dagger.LintingValidateJSONSchemaOpts{Glob: schemaGlob})
	out.WriteString(schemaOut)
	if err != nil {
		return out.String(), fmt.Errorf("validate-json-schema failed: %w", err)
	}

	// 2. Chart discovery + Chart.yaml name uniqueness pre-flight.
	out.WriteString("\n=== [2/4] chart discovery + Chart.yaml name uniqueness ===\n")
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
		return out.String(), fmt.Errorf("Chart.yaml name uniqueness violated:\n  %s", strings.Join(dupes, "\n  "))
	}
	out.WriteString("all Chart.yaml names are unique\n")

	// 3. Per-chart helm.Lint + helm.Kubeconform.
	out.WriteString("\n=== [3/4] helm.Lint + helm.Kubeconform per chart ===\n")
	for _, c := range charts {
		fmt.Fprintf(&out, "\n--- %s ---\n", c.dir)
		chartDir := src.Directory(c.dir)

		lintOut, err := dag.Helm().Lint(ctx, chartDir)
		out.WriteString(lintOut)
		if err != nil {
			return out.String(), fmt.Errorf("helm.Lint failed for %s: %w", c.dir, err)
		}

		kcOpts := dagger.HelmKubeconformOpts{
			SchemaLocations: schemaLocations,
			RegistrySecret:  registrySecret,
		}
		ciValues := path.Join(c.dir, "ci", "default-values.yaml")
		if has, _ := pathExists(ctx, src, ciValues); has {
			fmt.Fprintf(&out, "using values-file %s\n", ciValues)
			kcOpts.ValuesFile = src.File(ciValues)
		}
		kcOut, err := dag.Helm().Kubeconform(ctx, chartDir, kcOpts)
		out.WriteString(kcOut)
		if err != nil {
			return out.String(), fmt.Errorf("kubeconform failed for %s: %w", c.dir, err)
		}
	}

	// 4. argocd#41 regression: rendered Application metadata.name must differ
	//    across destination.server values. Skip charts that emit no Applications.
	out.WriteString("\n=== [4/4] argocd#41 regression: rendered Application name uniqueness across cluster destinations ===\n")
	for _, c := range charts {
		fmt.Fprintf(&out, "\n--- %s ---\n", c.dir)
		chartDir := src.Directory(c.dir)

		var baseBytes []byte
		ciValues := path.Join(c.dir, "ci", "default-values.yaml")
		if has, _ := pathExists(ctx, src, ciValues); has {
			s, err := src.File(ciValues).Contents(ctx)
			if err != nil {
				return out.String(), fmt.Errorf("read %s: %w", ciValues, err)
			}
			baseBytes = []byte(s)
		}

		fileA, err := overlayDestinationServer(baseBytes, probeServerA)
		if err != nil {
			return out.String(), fmt.Errorf("build probe values A for %s: %w", c.dir, err)
		}
		fileB, err := overlayDestinationServer(baseBytes, probeServerB)
		if err != nil {
			return out.String(), fmt.Errorf("build probe values B for %s: %w", c.dir, err)
		}

		manifestsA, err := dag.Helm().Render(ctx, chartDir, dagger.HelmRenderOpts{ValuesFile: fileA, RegistrySecret: registrySecret})
		if err != nil {
			return out.String(), fmt.Errorf("render probe A for %s: %w", c.dir, err)
		}
		manifestsB, err := dag.Helm().Render(ctx, chartDir, dagger.HelmRenderOpts{ValuesFile: fileB, RegistrySecret: registrySecret})
		if err != nil {
			return out.String(), fmt.Errorf("render probe B for %s: %w", c.dir, err)
		}

		appsA := extractApplicationNames(manifestsA)
		appsB := extractApplicationNames(manifestsB)
		if len(appsA) == 0 && len(appsB) == 0 {
			out.WriteString("no Application resources emitted — skipping\n")
			continue
		}
		// Identical Application name sets across two distinct destination.server
		// values means the chart does not differentiate Application names per
		// cluster — that is the argocd#41 regression.
		if equalStringSlices(appsA, appsB) {
			fmt.Fprintf(&out,
				"FAIL: rendered Application names did not change between cluster destinations\n  cluster-a (%s): %v\n  cluster-b (%s): %v\n",
				probeServerA, appsA, probeServerB, appsB)
			return out.String(), fmt.Errorf("argocd#41 regression: %s emits identical Application metadata.name set for different destination.server values", c.dir)
		}
		fmt.Fprintf(&out, "OK: Application names differ across destination.server\n  cluster-a: %v\n  cluster-b: %v\n", appsA, appsB)
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

// overlayDestinationServer returns a values file containing the optional base
// values plus destination.server forced to `server`. Backs the argocd#41
// double-render probe.
func overlayDestinationServer(base []byte, server string) (*dagger.File, error) {
	root := map[string]any{}
	if len(base) > 0 {
		if err := yaml.Unmarshal(base, &root); err != nil {
			return nil, fmt.Errorf("unmarshal base values: %w", err)
		}
		if root == nil {
			root = map[string]any{}
		}
	}
	dest, _ := root["destination"].(map[string]any)
	if dest == nil {
		dest = map[string]any{}
	}
	dest["server"] = server
	root["destination"] = dest

	body, err := yaml.Marshal(root)
	if err != nil {
		return nil, fmt.Errorf("marshal probe values: %w", err)
	}
	return dag.Directory().WithNewFile("probe-values.yaml", string(body)).File("probe-values.yaml"), nil
}

// extractApplicationNames returns the sorted metadata.name list of every
// argoproj.io Application resource in the rendered manifest stream.
func extractApplicationNames(manifests string) []string {
	names := []string{}
	dec := yaml.NewDecoder(strings.NewReader(manifests))
	for {
		var doc struct {
			APIVersion string `yaml:"apiVersion"`
			Kind       string `yaml:"kind"`
			Metadata   struct {
				Name string `yaml:"name"`
			} `yaml:"metadata"`
		}
		if err := dec.Decode(&doc); err != nil {
			break
		}
		if doc.Kind != "Application" {
			continue
		}
		if !strings.HasPrefix(doc.APIVersion, "argoproj.io/") {
			continue
		}
		if doc.Metadata.Name == "" {
			continue
		}
		names = append(names, doc.Metadata.Name)
	}
	sort.Strings(names)
	return names
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
