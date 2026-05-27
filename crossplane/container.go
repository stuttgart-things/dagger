package main

import (
	"context"
	"dagger/crossplane/internal/dagger"
)

const (
	kubeconformVersion = "v0.6.7"
	// openapi2jsonschema.py is pinned to the same kubeconform release so the
	// script and binary evolve together.
	openapi2jsonschemaURL = "https://raw.githubusercontent.com/yannh/kubeconform/" + kubeconformVersion + "/scripts/openapi2jsonschema.py"
)

// GetXplaneContainer returns the default image for Crossplane with crossplane and kcl2xrd installed
func (m *Crossplane) GetXplaneContainer(ctx context.Context) *dagger.Container {
	return dag.Container().
		From("cgr.dev/chainguard/wolfi-base:latest").
		// Install dependencies. python + py3-pyyaml + crane back the Verify pipeline
		// (CRD schema extraction + provider image export); curl + yq are also used elsewhere.
		WithExec([]string{"apk", "add", "curl", "yq", "crane", "python-3.13", "py3-pyyaml"}).
		// Install crossplane
		WithExec([]string{"curl", "https://releases.crossplane.io/stable/current/bin/linux_amd64/crank", "--output", "crossplane"}).
		WithExec([]string{"mv", "crossplane", "/usr/bin/crossplane"}).
		WithExec([]string{"chmod", "+x", "/usr/bin/crossplane"}).
		// Install kcl2xrd
		WithExec([]string{"curl", "-L", "https://github.com/ggkhrmv/kcl2xrd/releases/download/v0.8.0/kcl2xrd-linux-amd64", "--output", "kcl2xrd"}).
		WithExec([]string{"mv", "kcl2xrd", "/usr/bin/kcl2xrd"}).
		WithExec([]string{"chmod", "+x", "/usr/bin/kcl2xrd"}).
		// Install kubeconform (release binary; not in wolfi apk)
		WithExec([]string{"sh", "-c",
			"curl -sL https://github.com/yannh/kubeconform/releases/download/" + kubeconformVersion +
				"/kubeconform-linux-amd64.tar.gz | tar -xz -C /usr/bin kubeconform"}).
		WithExec([]string{"chmod", "+x", "/usr/bin/kubeconform"}).
		// Install openapi2jsonschema (CRD -> JSON Schema converter used by Verify).
		// Filenames are lowercase, which matches kubeconform's {{.ResourceKind}}
		// template (it lowercases the kind from the resource).
		WithExec([]string{"curl", "-sL", openapi2jsonschemaURL, "-o", "/usr/bin/openapi2jsonschema"}).
		WithExec([]string{"chmod", "+x", "/usr/bin/openapi2jsonschema"})
}
