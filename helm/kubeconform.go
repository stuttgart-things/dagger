package main

import (
	"context"
	"dagger/helm/internal/dagger"
)

// defaultSchemaLocations is the Datree CRDs-catalog fallback set covering the
// CRDs argocd-catalog charts lean on (ArgoCD, Gateway-API, cert-manager, cilium).
// The `default` token keeps the upstream Kubernetes schemas in play.
var defaultSchemaLocations = []string{
	"default",
	"https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json",
}

func (m *Helm) Kubeconform(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	valuesFile *dagger.File,
	// +optional
	schemaLocations []string,
	// +optional
	registrySecret *dagger.Secret,
) (string, error) {

	renderedManifests, err := m.Render(
		ctx,
		src,
		valuesFile,
		registrySecret,
	)
	if err != nil {
		return "", err
	}

	if len(schemaLocations) == 0 {
		schemaLocations = defaultSchemaLocations
	}

	args := []string{"kubeconform", "-strict", "-summary", "-ignore-missing-schemas", "-output", "json"}
	for _, loc := range schemaLocations {
		args = append(args, "-schema-location", loc)
	}
	args = append(args, "/manifests/rendered.yaml")

	return m.container().
		WithWorkdir("/manifests").
		WithNewFile("rendered.yaml", renderedManifests).
		WithExec(args).
		Stdout(ctx)
}
