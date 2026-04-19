// A generated module for Argocd functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/argocd/internal/dagger"
	"fmt"
	"strings"
)

type Argocd struct{}

// AddCluster registers a Kubernetes cluster in ArgoCD via `argocd cluster add`.
//
// Builds a Wolfi-based container, installs kubectl and the argocd CLI (downloaded
// directly from the upstream GitHub release), logs in to the ArgoCD server, renames
// the kubeconfig context to clusterName when needed (handy for k3s/k3d kubeconfigs
// whose context is just "default"), and runs `argocd cluster add <clusterName> --yes`.
// When labels are provided, `argocd cluster set` is invoked afterwards to apply them.
func (m *Argocd) AddCluster(
	ctx context.Context,
	// Kubeconfig of the target cluster to register
	kubeConfig *dagger.Secret,
	// ArgoCD server address (host[:port], no scheme)
	argocdServer string,
	// ArgoCD username
	username string,
	// ArgoCD password
	password *dagger.Secret,
	// Display name for the cluster in ArgoCD (also the renamed kubeconfig context)
	clusterName string,
	// +optional
	// +default="cgr.dev/chainguard/wolfi-base:latest"
	baseImage string,
	// Use plain HTTP to talk to the ArgoCD server.
	// +optional
	// +default=true
	plaintext bool,
	// Skip TLS verification when talking to the ArgoCD server (ignored if plaintext).
	// +optional
	// +default=true
	insecure bool,
	// Existing context in the kubeconfig to rename to clusterName. Defaults to "default"
	// to match k3s/k3d kubeconfigs. Set empty to skip the rename.
	// +optional
	// +default="default"
	sourceContext string,
	// Labels to apply to the registered cluster via `argocd cluster set --label`.
	// Each entry is key=value, e.g. ["auto-project=true", "env=prod"].
	// +optional
	labels []string,
	// Download URL for the argocd CLI binary.
	// +optional
	// +default="https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64"
	argocdDownloadURL string,
) (string, error) {
	if clusterName == "" {
		return "", fmt.Errorf("clusterName must not be empty")
	}
	if argocdServer == "" {
		return "", fmt.Errorf("argocdServer must not be empty")
	}

	const kubeconfigPath = "/tmp/kubeconfig"

	ctr := m.BaseContainer(baseImage, argocdDownloadURL).
		WithMountedSecret(kubeconfigPath, kubeConfig, dagger.ContainerWithMountedSecretOpts{
			Mode: 0444,
		}).
		WithSecretVariable("ARGOCD_PASSWORD", password).
		WithEnvVariable("ARGOCD_USERNAME", username).
		WithEnvVariable("ARGOCD_SERVER", argocdServer).
		WithEnvVariable("CLUSTER_NAME", clusterName).
		WithEnvVariable("SOURCE_CONTEXT", sourceContext)

	tlsFlag := ""
	switch {
	case plaintext:
		tlsFlag = "--plaintext"
	case insecure:
		tlsFlag = "--insecure"
	}

	var labelArgs strings.Builder
	for _, l := range labels {
		if l == "" {
			continue
		}
		fmt.Fprintf(&labelArgs, " --label %q", l)
	}
	setCmd := ""
	if labelArgs.Len() > 0 {
		setCmd = fmt.Sprintf(`argocd cluster set "$CLUSTER_NAME"%s`, labelArgs.String())
	}

	// `yes |` auto-answers the "server is not configured with TLS" y/n prompt.
	script := fmt.Sprintf(`set -eu
cp %q /tmp/kubeconfig.rw
chmod 600 /tmp/kubeconfig.rw
export KUBECONFIG=/tmp/kubeconfig.rw

if [ -n "$SOURCE_CONTEXT" ] && [ "$SOURCE_CONTEXT" != "$CLUSTER_NAME" ]; then
  if kubectl config get-contexts -o name | grep -qx "$SOURCE_CONTEXT"; then
    kubectl config rename-context "$SOURCE_CONTEXT" "$CLUSTER_NAME"
  fi
fi

yes | argocd login "$ARGOCD_SERVER" --username "$ARGOCD_USERNAME" --password "$ARGOCD_PASSWORD" %s --grpc-web
argocd cluster add "$CLUSTER_NAME" --yes
%s
`, kubeconfigPath, tlsFlag, setCmd)

	return ctr.WithExec([]string{"sh", "-c", script}).Stdout(ctx)
}
