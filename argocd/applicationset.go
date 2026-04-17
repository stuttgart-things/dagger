package main

import (
	"context"
	"dagger/argocd/internal/dagger"
	"fmt"
	"strings"
)

// CreateApplicationSet renders an ArgoCD ApplicationSet manifest from the
// stuttgart-things/argocd-application-set KCL module (hosted as an OCI artifact)
// and optionally applies it to the ArgoCD-hosting cluster.
//
// See https://github.com/stuttgart-things/kcl/tree/main/kubernetes/argocd-application-set
// for the full parameter reference. The `generators` field is the main lever;
// every complex nested field takes a JSON string that passes straight through
// to `kcl run -D key=<json>`. Argo Go-template expressions like
// `{{ .cluster.name }}` survive the JSON parser untouched as long as they're
// inside string values.
func (m *Argocd) CreateApplicationSet(
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
	// spec.goTemplate (pass "false" to disable Go templating)
	// +optional
	goTemplate string,
	// spec.goTemplateOptions as JSON array
	// +optional
	goTemplateOptions string,
	// spec.generators as JSON array. The module's default renders the `kro` example.
	// +optional
	generators string,
	// template.spec.project
	// +optional
	project string,
	// template.metadata.name (usually contains Go-template expressions)
	// +optional
	templateName string,
	// template.metadata.namespace
	// +optional
	templateNamespace string,
	// template.metadata.labels as JSON object
	// +optional
	templateLabels string,
	// template.metadata.annotations as JSON object
	// +optional
	templateAnnotations string,
	// template.metadata.finalizers as JSON array
	// +optional
	templateFinalizers string,
	// Whole template.metadata dict as JSON (overrides templateName/templateLabels/…)
	// +optional
	templateMetadata string,
	// template.spec.source dict as JSON (single-source apps)
	// +optional
	source string,
	// template.spec.sources as JSON array (multi-source apps; replaces source)
	// +optional
	sources string,
	// template.spec.destination.server
	// +optional
	destServer string,
	// template.spec.destination.name (mutually exclusive with destServer)
	// +optional
	destName string,
	// template.spec.destination.namespace
	// +optional
	destNamespace string,
	// Whole template.spec.destination dict as JSON
	// +optional
	destination string,
	// template.spec.syncPolicy.syncOptions as JSON array
	// +optional
	syncOptions string,
	// template.spec.syncPolicy.automated as JSON
	// +optional
	automated string,
	// template.spec.syncPolicy.retry as JSON
	// +optional
	retry string,
	// Whole template.spec.syncPolicy dict as JSON
	// +optional
	templateSyncPolicy string,
	// Whole template.spec dict as JSON
	// +optional
	templateSpec string,
	// Whole spec.template dict as JSON (overrides everything under template.*)
	// +optional
	template string,
	// spec.syncPolicy (appset-level, e.g. preserveResourcesOnDeletion) as JSON
	// +optional
	syncPolicyTopLevel string,
	// spec.strategy (RollingSync) as JSON
	// +optional
	strategy string,
	// spec.preservedFields as JSON
	// +optional
	preservedFields string,
	// spec.templatePatch (raw string)
	// +optional
	templatePatch string,
	// metadata.labels as JSON object
	// +optional
	labels string,
	// metadata.annotations as JSON object
	// +optional
	annotations string,
	// metadata.finalizers as JSON array
	// +optional
	finalizers string,
	// OCI source of the KCL module; append ?tag=<version> to pin
	// +optional
	// +default="oci://ghcr.io/stuttgart-things/argocd-application-set"
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
	add("goTemplate", goTemplate)
	add("goTemplateOptions", goTemplateOptions)
	add("generators", generators)
	add("project", project)
	add("templateName", templateName)
	add("templateNamespace", templateNamespace)
	add("templateLabels", templateLabels)
	add("templateAnnotations", templateAnnotations)
	add("templateFinalizers", templateFinalizers)
	add("templateMetadata", templateMetadata)
	add("source", source)
	add("sources", sources)
	add("destServer", destServer)
	add("destName", destName)
	add("destNamespace", destNamespace)
	add("destination", destination)
	add("syncOptions", syncOptions)
	add("automated", automated)
	add("retry", retry)
	add("templateSyncPolicy", templateSyncPolicy)
	add("templateSpec", templateSpec)
	add("template", template)
	add("syncPolicyTopLevel", syncPolicyTopLevel)
	add("strategy", strategy)
	add("preservedFields", preservedFields)
	add("templatePatch", templatePatch)
	add("labels", labels)
	add("annotations", annotations)
	add("finalizers", finalizers)

	renderedFile := dag.Kcl().Run(dagger.KclRunOpts{
		OciSource:      ociSource,
		Parameters:     strings.Join(params, ","),
		ParametersFile: parametersFile,
		Entrypoint:     "main.k",
	})

	renderedContent, err := renderedFile.Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to render ApplicationSet: %w", err)
	}

	outputBase := name
	if outputBase == "" {
		outputBase = "applicationset"
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
