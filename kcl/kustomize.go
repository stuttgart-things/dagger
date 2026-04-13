package main

import (
	"context"
	"dagger/kcl/internal/dagger"
	"fmt"
)

// RenderKustomizeBase renders KCL output into a kustomize base directory.
// It calls Run() internally to produce multi-document YAML, then splits it
// into individual kind-name.yaml files and generates a kustomization.yaml.
//
// Example usage:
//
//	dagger call render-kustomize-base \
//	  --source ./deployment \
//	  --parameters-file ./tests/kcl-deploy-profile.yaml \
//	  export --path=/tmp/kustomize-base
func (m *Kcl) RenderKustomizeBase(
	ctx context.Context,
	// Local source directory (optional if using OCI source)
	// +optional
	source *dagger.Directory,
	// OCI source path (e.g., oci://ghcr.io/stuttgart-things/kcl-flux-instance)
	// +optional
	ociSource string,
	// KCL parameters as comma-separated key=value pairs
	// +optional
	parameters string,
	// YAML/JSON file containing KCL parameters as key-value pairs
	// +optional
	parametersFile *dagger.File,
	// Entry point file name
	// +optional
	// +default="main.k"
	entrypoint string,
	// Sub-path inside source to cd into before running kcl. Enables KCL
	// packages with relative path deps pointing outside their own directory.
	// +optional
	subpath string,
) (*dagger.Directory, error) {

	// Run KCL with formatOutput=true and outputFormat="yaml" to get clean multi-doc YAML
	renderedFile, err := m.Run(ctx, source, ociSource, parameters, parametersFile, true, "yaml", entrypoint, subpath)
	if err != nil {
		return nil, err
	}

	ctr := m.container().
		WithMountedFile("/rendered.yaml", renderedFile).
		WithWorkdir("/output")

	// Split multi-doc YAML into individual kind-name.yaml files and generate kustomization.yaml.
	// - Iterates over each YAML document using yq document index
	// - Skips null/empty documents
	// - Names files as lowercase kind-name.yaml with collision fallback (-N suffix)
	// - Writes kustomization.yaml listing all resource files
	splitScript := `#!/bin/sh
set -e

DOC_COUNT=$(yq eval-all '[.] | length' /rendered.yaml)

for i in $(seq 0 $((DOC_COUNT - 1))); do
  DOC=$(yq eval-all "select(documentIndex == $i)" /rendered.yaml)

  # Skip null or empty documents
  if [ "$DOC" = "null" ] || [ -z "$DOC" ]; then
    continue
  fi

  KIND=$(echo "$DOC" | yq eval '.kind // ""' -)
  NAME=$(echo "$DOC" | yq eval '.metadata.name // ""' -)

  if [ -z "$KIND" ] || [ -z "$NAME" ]; then
    continue
  fi

  # Lowercase kind
  KIND_LOWER=$(echo "$KIND" | tr '[:upper:]' '[:lower:]')

  # Sanitize name: lowercase, non-alphanumeric to dash, collapse dashes, trim
  NAME_SAFE=$(echo "$NAME" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9-]/-/g' | sed 's/--*/-/g' | sed 's/^-//;s/-$//')

  FILENAME="${KIND_LOWER}-${NAME_SAFE}.yaml"

  # Collision fallback: append document index if file exists
  if [ -f "/output/${FILENAME}" ]; then
    FILENAME="${KIND_LOWER}-${NAME_SAFE}-${i}.yaml"
  fi

  echo "$DOC" > "/output/${FILENAME}"
done

# Generate kustomization.yaml from all resource files
echo "apiVersion: kustomize.config.k8s.io/v1beta1" > /output/kustomization.yaml
echo "kind: Kustomization" >> /output/kustomization.yaml
echo "resources:" >> /output/kustomization.yaml

for f in /output/*.yaml; do
  BASENAME=$(basename "$f")
  if [ "$BASENAME" != "kustomization.yaml" ]; then
    echo "  - ${BASENAME}" >> /output/kustomization.yaml
  fi
done
`

	ctr = ctr.WithNewFile("/split.sh", splitScript, dagger.ContainerWithNewFileOpts{Permissions: 0o755}).
		WithExec([]string{"sh", "/split.sh"})

	return ctr.Directory("/output"), nil
}

// PushKustomizeBase renders a kustomize base from KCL and pushes it as an OCI artifact.
// It calls RenderKustomizeBase() internally, then uses oras to push the result.
//
// Example usage:
//
//	dagger call push-kustomize-base \
//	  --source ./deployment \
//	  --parameters-file ./tests/kcl-deploy-profile.yaml \
//	  --address ghcr.io/stuttgart-things/my-app-kustomize \
//	  --tag v1.0.0 \
//	  --user env:GITHUB_USER \
//	  --password env:GITHUB_TOKEN
func (m *Kcl) PushKustomizeBase(
	ctx context.Context,
	// Local source directory (optional if using OCI source)
	// +optional
	source *dagger.Directory,
	// OCI source path (e.g., oci://ghcr.io/stuttgart-things/kcl-flux-instance)
	// +optional
	ociSource string,
	// KCL parameters as comma-separated key=value pairs
	// +optional
	parameters string,
	// YAML/JSON file containing KCL parameters as key-value pairs
	// +optional
	parametersFile *dagger.File,
	// Entry point file name
	// +optional
	// +default="main.k"
	entrypoint string,
	// Sub-path inside source to cd into before running kcl. Enables KCL
	// packages with relative path deps pointing outside their own directory.
	// +optional
	subpath string,
	// OCI address (e.g., ghcr.io/stuttgart-things/my-app-kustomize)
	address string,
	// Version tag (e.g., v1.0.0)
	tag string,
	// Environment variable name for registry username
	// +optional
	// +default="GITHUB_USER"
	userName string,
	// Registry username as a secret
	// +optional
	user *dagger.Secret,
	// Environment variable name for registry password
	// +optional
	// +default="GITHUB_TOKEN"
	passwordName string,
	// Registry password as a secret
	// +optional
	password *dagger.Secret,
) (string, error) {

	// Render the kustomize base directory
	baseDir, err := m.RenderKustomizeBase(ctx, source, ociSource, parameters, parametersFile, entrypoint, subpath)
	if err != nil {
		return "", err
	}

	// Extract registry host from address (e.g., "ghcr.io/org/repo" -> "ghcr.io")
	registry := "ghcr.io"
	for i, c := range address {
		if c == '/' {
			registry = address[:i]
			break
		}
	}

	ref := fmt.Sprintf("%s:%s", address, tag)

	// Install oras, login, and push
	ctr := m.container().
		WithExec([]string{"sh", "-c", "curl -sL https://github.com/oras-project/oras/releases/download/v1.2.2/oras_1.2.2_linux_amd64.tar.gz | tar xz -C /usr/local/bin oras"}).
		WithMountedDirectory("/kustomize-base", baseDir).
		WithWorkdir("/kustomize-base").
		WithSecretVariable(userName, user).
		WithSecretVariable(passwordName, password).
		WithExec([]string{"sh", "-c", fmt.Sprintf("oras login %s -u $%s -p $%s", registry, userName, passwordName)}).
		WithExec([]string{"sh", "-c", fmt.Sprintf("oras push %s .", ref)})

	_, err = ctr.Stdout(ctx)
	if err != nil {
		return "", err
	}

	return ref, nil
}
