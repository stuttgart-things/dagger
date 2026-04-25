package main

import (
	"context"
	"dagger/kcl/internal/dagger"
)

// RenderClusterbookCluster renders the clusterbook-cluster-gen KCL module
// from its OCI artifact and returns the rendered YAML file. It is a thin
// wrapper around Run() with the OCI source pinned to
// ghcr.io/stuttgart-things/clusterbook-cluster-gen:0.1.0 by default.
//
// Example usage:
//
//	dagger call -m ./kcl render-clusterbook-cluster \
//	  --parameters 'clusterName=demo,networkKey=net-a,providerConfigRef={"name":"default"},kubeconfigSecretRef={"name":"kubeconfig","namespace":"default"}' \
//	  export --path=/tmp/clusterbook-cluster.yaml
func (m *Kcl) RenderClusterbookCluster(
	ctx context.Context,
	// OCI source path for the clusterbook-cluster-gen module.
	// +optional
	// +default="ghcr.io/stuttgart-things/clusterbook-cluster-gen:0.1.0"
	ociSource string,
	// KCL parameters as comma-separated key=value pairs.
	// +optional
	parameters string,
	// YAML/JSON file containing KCL parameters as key-value pairs.
	// +optional
	parametersFile *dagger.File,
	// Output format: yaml or json.
	// +optional
	// +default="yaml"
	outputFormat string,
) (*dagger.File, error) {
	return m.Run(ctx, nil, ociSource, parameters, parametersFile, true, outputFormat, "main.k", "")
}
