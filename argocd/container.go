package main

import (
	"fmt"

	"dagger/argocd/internal/dagger"
)

const (
	defaultBaseImage         = "cgr.dev/chainguard/wolfi-base:latest"
	defaultArgocdDownloadURL = "https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64"
)

// BaseContainer returns a Wolfi-based container with curl, git, kubectl and the
// argocd CLI installed. The argocd binary is fetched from the given download URL
// and placed at /usr/local/bin/argocd.
func (m *Argocd) BaseContainer(
	// +optional
	// +default="cgr.dev/chainguard/wolfi-base:latest"
	baseImage string,
	// Download URL for the argocd CLI binary.
	// +optional
	// +default="https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64"
	argocdDownloadURL string,
) *dagger.Container {
	if baseImage == "" {
		baseImage = defaultBaseImage
	}
	if argocdDownloadURL == "" {
		argocdDownloadURL = defaultArgocdDownloadURL
	}

	return dag.Container().
		From(baseImage).
		WithExec([]string{"apk", "add", "--no-cache", "curl", "git", "kubectl"}).
		WithExec([]string{"sh", "-c", fmt.Sprintf(
			`curl -sSL -o /usr/local/bin/argocd %q && chmod +x /usr/local/bin/argocd`,
			argocdDownloadURL,
		)})
}
