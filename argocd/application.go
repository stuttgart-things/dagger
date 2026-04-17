package main

import (
	"context"
	"dagger/argocd/internal/dagger"
	"fmt"
	"strings"
)

// CreateApplication renders an ArgoCD Application manifest from the
// stuttgart-things/argocd-application KCL module (hosted as an OCI artifact)
// and optionally applies it to the ArgoCD-hosting cluster.
//
// See https://github.com/stuttgart-things/kcl/tree/main/kubernetes/argocd-application
// for the full parameter reference. Common scalar fields are exposed directly;
// complex nested fields (helm, kustomize, sources, syncPolicy, …) take JSON
// strings that pass straight through to `kcl run -D key=<json>`.
func (m *Argocd) CreateApplication(
	ctx context.Context,
	// metadata.name and output file basename.
	// Can also be supplied via parametersFile; the CLI value wins.
	// +optional
	name string,
	// YAML/JSON file with KCL parameters as key: value pairs. Every CLI flag
	// below takes precedence over values in this file.
	// +optional
	parametersFile *dagger.File,
	// Namespace where ArgoCD is installed (KCL default: "argocd")
	// +optional
	namespace string,
	// AppProject this Application belongs to (KCL default: "default")
	// +optional
	project string,
	// spec.source.repoURL (git URL or Helm repo URL)
	// +optional
	repoURL string,
	// spec.source.path (git dir; mutually exclusive with chart)
	// +optional
	path string,
	// spec.source.targetRevision
	// +optional
	targetRevision string,
	// spec.source.chart (Helm chart name; mutually exclusive with path)
	// +optional
	chart string,
	// spec.destination.server
	// +optional
	destServer string,
	// spec.destination.name (mutually exclusive with destServer)
	// +optional
	destName string,
	// spec.destination.namespace
	// +optional
	destNamespace string,
	// syncPolicy.syncOptions as JSON array (e.g. '["CreateNamespace=true"]')
	// +optional
	syncOptions string,
	// Full spec.source.helm dict as JSON
	// +optional
	helm string,
	// Full spec.source.kustomize dict as JSON
	// +optional
	kustomize string,
	// Entire spec.source dict as JSON (overrides repoURL/path/chart/helm/etc.)
	// +optional
	source string,
	// Multi-source spec.sources as JSON array (replaces source when set)
	// +optional
	sources string,
	// Entire spec.destination dict as JSON (overrides destServer/destName/destNamespace)
	// +optional
	destination string,
	// Entire spec.syncPolicy dict as JSON
	// +optional
	syncPolicy string,
	// spec.syncPolicy.automated as JSON
	// +optional
	automated string,
	// spec.syncPolicy.retry as JSON
	// +optional
	retry string,
	// spec.info as JSON array
	// +optional
	info string,
	// metadata.labels as JSON object
	// +optional
	labels string,
	// metadata.annotations as JSON object
	// +optional
	annotations string,
	// metadata.finalizers as JSON array
	// +optional
	finalizers string,
	// spec.revisionHistoryLimit
	// +optional
	revisionHistoryLimit string,
	// OCI source of the KCL module; append ?tag=<version> to pin
	// +optional
	// +default="oci://ghcr.io/stuttgart-things/argocd-application"
	ociSource string,
	// +optional
	// +default="yaml"
	fileExtension string,
	// Apply the rendered manifest via kubectl
	// +optional
	// +default=false
	applyToCluster bool,
	// Kubeconfig of the ArgoCD-hosting cluster (required when applyToCluster is true)
	// +optional
	kubeConfig *dagger.Secret,
) (*dagger.Directory, error) {

	if name == "" && parametersFile == nil {
		return nil, fmt.Errorf("either name or parametersFile must be provided")
	}
	if applyToCluster && kubeConfig == nil {
		return nil, fmt.Errorf("kubeConfig is required when applyToCluster is true")
	}

	var params []string
	add := func(key, value string) {
		if value != "" {
			params = append(params, key+"="+value)
		}
	}
	add("name", name)
	add("namespace", namespace)
	add("project", project)
	add("repoURL", repoURL)
	add("path", path)
	add("targetRevision", targetRevision)
	add("chart", chart)
	add("destServer", destServer)
	add("destName", destName)
	add("destNamespace", destNamespace)
	add("syncOptions", syncOptions)
	add("helm", helm)
	add("kustomize", kustomize)
	add("source", source)
	add("sources", sources)
	add("destination", destination)
	add("syncPolicy", syncPolicy)
	add("automated", automated)
	add("retry", retry)
	add("info", info)
	add("labels", labels)
	add("annotations", annotations)
	add("finalizers", finalizers)
	add("revisionHistoryLimit", revisionHistoryLimit)

	renderedFile := dag.Kcl().Run(dagger.KclRunOpts{
		OciSource:      ociSource,
		Parameters:     strings.Join(params, ","),
		ParametersFile: parametersFile,
		Entrypoint:     "main.k",
	})

	renderedContent, err := renderedFile.Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to render Application: %w", err)
	}

	outputBase := name
	if outputBase == "" {
		outputBase = "application"
	}
	outputPath := outputBase + "." + fileExtension
	outputDir := dag.Directory().WithNewFile(outputPath, renderedContent)

	if applyToCluster {
		applyNs := namespace
		if applyNs == "" {
			applyNs = "argocd"
		}
		if _, err := dag.Kubernetes().Kubectl(ctx, dagger.KubernetesKubectlOpts{
			Operation:  "apply",
			SourceFile: outputDir.File(outputPath),
			KubeConfig: kubeConfig,
			Namespace:  applyNs,
		}); err != nil {
			return nil, fmt.Errorf("kubectl apply failed: %w", err)
		}
	}

	return outputDir, nil
}
