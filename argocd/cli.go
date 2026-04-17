package main

import (
	"context"
	"dagger/argocd/internal/dagger"
	"fmt"
)

// AddClusterCli registers a Kubernetes cluster in ArgoCD via `argocd cluster add`.
//
// It builds a Wolfi-based container with the argocd CLI and kubectl, logs in to the
// ArgoCD server, and runs `argocd cluster add <context> --name <clusterName>`.
// The kubeconfig context is taken from its current-context (or --source-context when
// supplied); no context rename is performed.
func (m *Argocd) AddClusterCli(
	ctx context.Context,
	// Kubeconfig of the target cluster
	kubeConfig *dagger.Secret,
	// ArgoCD server address (host[:port], no scheme)
	argocdServer string,
	// ArgoCD username (not a secret)
	username string,
	// ArgoCD password
	password *dagger.Secret,
	// Display name for the cluster in ArgoCD (--name on `argocd cluster add`)
	clusterName string,
	// +optional
	// +default="cgr.dev/chainguard/wolfi-base:latest"
	baseImage string,
	// +optional
	// +default=true
	insecure bool,
	// CA certificate (PEM) to verify the ArgoCD server. When provided, insecure is ignored.
	// +optional
	serverCert *dagger.File,
	// Directory of CA certificates (*.crt / *.pem) concatenated into the trust bundle
	// passed to `argocd login --server-crt`. When provided, insecure is ignored.
	// Takes precedence over serverCert.
	// +optional
	serverCertsDir *dagger.Directory,
	// Kubeconfig context to register. If empty, the kubeconfig's current-context is used.
	// +optional
	sourceContext string,
	// Use plain HTTP (no TLS) to talk to the ArgoCD server.
	// +optional
	// +default=false
	plaintext bool,
	// Wolfi apk package providing the argocd CLI. Pin to a major.minor that matches
	// your ArgoCD server (e.g. argo-cd-2.14, argo-cd-3.3). The "argo-cd" meta-package
	// tracks the latest major and can break against older servers.
	// +optional
	// +default="argo-cd-2.14"
	cliPackage string,
) (string, error) {

	if clusterName == "" {
		return "", fmt.Errorf("clusterName must not be empty")
	}
	if argocdServer == "" {
		return "", fmt.Errorf("argocdServer must not be empty")
	}

	const kubeconfigPath = "/tmp/kubeconfig"

	ctr := dag.Container().
		From(baseImage).
		WithExec([]string{"apk", "add", "--no-cache", cliPackage, "kubectl"}).
		WithMountedSecret(kubeconfigPath, kubeConfig, dagger.ContainerWithMountedSecretOpts{
			Mode: 0444,
		}).
		WithSecretVariable("ARGOCD_PASSWORD", password).
		WithEnvVariable("ARGOCD_USERNAME", username).
		WithEnvVariable("ARGOCD_SERVER", argocdServer).
		WithEnvVariable("CLUSTER_NAME", clusterName).
		WithEnvVariable("SOURCE_CONTEXT", sourceContext).
		WithEnvVariable("KUBECONFIG_PATH", kubeconfigPath)

	tlsFlag := ""
	bundleScript := ""
	switch {
	case plaintext:
		tlsFlag = "--plaintext"
	case serverCertsDir != nil:
		ctr = ctr.WithMountedDirectory("/tmp/argocd-cas", serverCertsDir)
		bundleScript = `mkdir -p /tmp/argocd-bundle
: > /tmp/argocd-bundle/ca.pem
for f in /tmp/argocd-cas/*.crt /tmp/argocd-cas/*.pem; do
  [ -f "$f" ] || continue
  cat "$f" >> /tmp/argocd-bundle/ca.pem
  printf '\n' >> /tmp/argocd-bundle/ca.pem
done
if [ ! -s /tmp/argocd-bundle/ca.pem ]; then
  echo "serverCertsDir contained no .crt/.pem files" >&2
  exit 1
fi
`
		tlsFlag = "--server-crt /tmp/argocd-bundle/ca.pem"
	case serverCert != nil:
		ctr = ctr.WithMountedFile("/tmp/argocd-server.crt", serverCert)
		tlsFlag = "--server-crt /tmp/argocd-server.crt"
	case insecure:
		tlsFlag = "--insecure"
	}

	// `yes |` auto-answers argocd's "server is not configured with TLS" y/n prompt
	// so we don't die on EOF when the server presents an unexpected TLS config.
	script := fmt.Sprintf(`set -eu
cp "$KUBECONFIG_PATH" /tmp/kubeconfig.rw
chmod 600 /tmp/kubeconfig.rw
export KUBECONFIG=/tmp/kubeconfig.rw

%s
src="${SOURCE_CONTEXT:-}"
if [ -z "$src" ]; then
  src=$(kubectl config current-context 2>/dev/null || true)
fi
if [ -z "$src" ]; then
  echo "could not determine source context; set --source-context explicitly. Available contexts:" >&2
  kubectl config get-contexts -o name >&2 || true
  exit 1
fi

yes | argocd login "$ARGOCD_SERVER" --username "$ARGOCD_USERNAME" --password "$ARGOCD_PASSWORD" %s --grpc-web
argocd cluster add "$src" --name "$CLUSTER_NAME" --kubeconfig /tmp/kubeconfig.rw --grpc-web --yes
`, bundleScript, tlsFlag)

	return ctr.WithExec([]string{"sh", "-c", script}).Stdout(ctx)
}
