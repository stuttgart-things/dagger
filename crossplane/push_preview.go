package main

import (
	"context"
	"dagger/crossplane/internal/dagger"
	"fmt"
	"strings"
)

// PushPreview builds a single Crossplane Configuration and pushes it to an
// ephemeral, anonymous registry (ttl.sh by default) for try-it-on-a-cluster
// preview iteration — e.g. a per-PR preview package in CI.
//
// Unlike Push, it needs no credentials: ttl.sh accepts anonymous pushes and
// serves the image publicly for the duration encoded in the tag (`:24h`).
// The package name is read from crossplane.yaml, so the caller only supplies
// a path prefix and a TTL. It returns the ready-to-apply install manifest
// (a kind: Configuration whose spec.package points at the pushed ref).
func (m *Crossplane) PushPreview(
	ctx context.Context,
	// a single Configuration directory (containing crossplane.yaml, apis/, examples/)
	src *dagger.Directory,
	// +optional
	// +default="ttl.sh"
	// registry to push the preview package to (must accept anonymous pushes)
	registry string,
	// path prefix under the registry, e.g. "stuttgart-things/crossplane-configurations-pr42-abc1234"
	prefix string,
	// +optional
	// +default="24h"
	// image lifetime; for ttl.sh this is the tag (max 24h)
	ttl string,
) (string, error) {

	if m.XplaneContainer == nil {
		m.XplaneContainer = m.GetXplaneContainer(ctx)
	}

	if prefix == "" {
		return "", fmt.Errorf("prefix is required")
	}

	// Read the package name from crossplane.yaml so the caller doesn't have to.
	pkg, err := m.XplaneContainer.
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"yq", "-r", ".metadata.name", "crossplane.yaml"}).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("reading package name from crossplane.yaml: %w", err)
	}
	pkg = strings.TrimSpace(pkg)
	if pkg == "" {
		return "", fmt.Errorf("crossplane.yaml has no metadata.name")
	}

	destination := fmt.Sprintf("%s/%s/%s:%s", registry, prefix, pkg, ttl)

	// Build the .xpkg, then push anonymously — no docker config is written, so
	// no credentials are required (correct for ttl.sh).
	dirWithPackage := m.Package(ctx, src)

	if _, err := m.XplaneContainer.
		WithDirectory("/src", dirWithPackage).
		WithWorkdir("/src").
		WithExec([]string{"crossplane", "xpkg", "push", destination}).
		Stdout(ctx); err != nil {
		return "", fmt.Errorf("pushing %s: %w", destination, err)
	}

	// The canonical, reusable artifact: a ready-to-apply install manifest.
	// Presentation (e.g. wrapping in a PR-comment <details> block) is the
	// caller's concern. The "-preview" suffix avoids clashing with a canonical
	// install of the same Configuration on a cluster.
	manifest := fmt.Sprintf(`apiVersion: pkg.crossplane.io/v1
kind: Configuration
metadata:
  name: %s-preview
spec:
  package: %s
`, pkg, destination)

	return manifest, nil
}
