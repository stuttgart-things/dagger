package main

import (
	"context"
	"dagger/argocd/internal/dagger"
	"fmt"
	"strings"
)

// CreateAppProject renders an ArgoCD AppProject manifest from the
// stuttgart-things/argocd-app-project KCL module (hosted as an OCI artifact) and
// optionally applies it to the ArgoCD-hosting cluster.
//
// Every field of the AppProject can be overridden via the individual parameters
// below. Complex fields (destinations, whitelists, labels) take JSON strings
// that are passed straight to `kcl run -D key=<json>`.
func (m *Argocd) CreateAppProject(
	ctx context.Context,
	// AppProject name (metadata.name and the output file basename)
	name string,
	// Namespace where ArgoCD is installed
	// +optional
	// +default="argocd"
	namespace string,
	// Free-form description written to spec.description
	// +optional
	description string,
	// Allowed source repo URLs, JSON array (e.g. '["https://github.com/org/repo"]')
	// +optional
	sourceRepos string,
	// Deployment destinations, JSON array of {server?,name?,namespace}
	// (e.g. '[{"server":"https://10.0.0.1:6443","namespace":"*"}]')
	// +optional
	destinations string,
	// Cluster-scoped resource kinds allowed, JSON array of {group,kind}
	// +optional
	clusterResourceWhitelist string,
	// Namespace-scoped resource kinds allowed, JSON array of {group,kind}
	// +optional
	namespaceResourceWhitelist string,
	// metadata.labels as JSON object (e.g. '{"team":"platform"}')
	// +optional
	labels string,
	// metadata.annotations as JSON object
	// +optional
	annotations string,
	// OCI source of the KCL module; append ?tag=<version> to pin.
	// +optional
	// +default="oci://ghcr.io/stuttgart-things/argocd-app-project"
	ociSource string,
	// File extension for the rendered manifest
	// +optional
	// +default="yaml"
	fileExtension string,
	// When true, apply the rendered manifest to the cluster via kubectl.
	// +optional
	// +default=false
	applyToCluster bool,
	// Kubeconfig of the ArgoCD-hosting cluster. Required when applyToCluster is true.
	// +optional
	kubeConfig *dagger.Secret,
) (*dagger.Directory, error) {

	if name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}
	if applyToCluster && kubeConfig == nil {
		return nil, fmt.Errorf("kubeConfig is required when applyToCluster is true")
	}

	params := []string{
		"name=" + name,
		"namespace=" + namespace,
	}
	addIfSet := func(key, value string) {
		if value != "" {
			params = append(params, key+"="+value)
		}
	}
	addIfSet("description", description)
	addIfSet("sourceRepos", sourceRepos)
	addIfSet("destinations", destinations)
	addIfSet("clusterResourceWhitelist", clusterResourceWhitelist)
	addIfSet("namespaceResourceWhitelist", namespaceResourceWhitelist)
	addIfSet("labels", labels)
	addIfSet("annotations", annotations)

	renderedFile := dag.Kcl().Run(dagger.KclRunOpts{
		OciSource:  ociSource,
		Parameters: strings.Join(params, ","),
		Entrypoint: "main.k",
	})

	renderedContent, err := renderedFile.Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to render AppProject: %w", err)
	}

	outputPath := name + "." + fileExtension
	outputDir := dag.Directory().WithNewFile(outputPath, renderedContent)

	if applyToCluster {
		if _, err := dag.Kubernetes().Kubectl(ctx, dagger.KubernetesKubectlOpts{
			Operation:  "apply",
			SourceFile: outputDir.File(outputPath),
			KubeConfig: kubeConfig,
			Namespace:  namespace,
		}); err != nil {
			return nil, fmt.Errorf("kubectl apply failed: %w", err)
		}
	}

	return outputDir, nil
}
