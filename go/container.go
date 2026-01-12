package main

import (
	"context"
	"dagger/go/internal/dagger"
	"fmt"
	"strings"
)

func (m *Go) KoBuild(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	// +default="GITHUB_TOKEN"
	tokenName string,
	// +optional
	token *dagger.Secret,
	// +optional
	// +default="ko.local"
	repo string,
	// +optional
	// +default="."
	buildArg string,
	// +optional
	// +default="v0.18.0"
	koVersion string,
	// +optional
	// +default="true"
	push string,
) (string, error) {
	srcDir := "/src"

	ctr := m.
		GetKoContainer(koVersion).
		WithDirectory(srcDir, src).
		WithWorkdir(srcDir).
		WithEnvVariable("GIT_COMMIT", "dev")

	if push == "true" {
		ctr = ctr.
			WithEnvVariable("KO_DOCKER_REPO", repo).
			WithSecretVariable(tokenName, token)
	}

	// Add OCI layout path when not pushing
	args := []string{"ko", "build", "--push=" + push, buildArg}

	if push == "false" {
		args = append(args, "--oci-layout-path=/oci-layout")
	}

	output, err := ctr.WithExec(args).Stdout(ctx)

	if err != nil {
		return "", fmt.Errorf("error running ko build: %w", err)
	}

	return strings.TrimSpace(output), nil
}

// GetGoLangContainer returns the golang build container
func (m *Go) GetGoLangContainer(
	// +optional
	// +default="1.25.5"
	goVersion string,

	// +optional
	// +default="alpine"
	variant string, // alpine | bookworm | bullseye | ""
) *dagger.Container {

	image := "golang:" + goVersion
	if variant != "" {
		image += "-" + variant
	}

	goModCache := dag.CacheVolume("go-mod-cache-" + goVersion)
	goBuildCache := dag.CacheVolume("go-build-cache-" + goVersion)

	return dag.Container().
		From(image).
		WithMountedCache("/go/pkg/mod", goModCache).
		WithMountedCache("/root/.cache/go-build", goBuildCache)
}

func (m *Go) GetKoContainer(
	// +optional
	// +default="v0.18.0"
	koVersion string,
) *dagger.Container {
	return dag.Container().
		From("ghcr.io/ko-build/ko:" + koVersion)
}

func (m *Go) GetReleaserContainer(
	// +optional
	// +default="v2.13.2"
	releaserVersion string,
) *dagger.Container {
	return dag.Container().
		From("goreleaser/goreleaser:" + releaserVersion)
}
