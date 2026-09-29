package main

import (
	"context"
	"dagger/kcl/internal/dagger"
)

// defaultDependencyRepo is the stuttgart-things mirror of the kcl-lang
// packages (stuttgart-things/kcl#323, filled by the kcl repo's kcl-mirror
// workflow). ghcr throttles ghcr.io/kcl-lang/* for everyone at times; on
// 2026-09-29 every render that pulled k8s failed with HTTP 429 for over an
// hour, retries (see withRegistryRetry) included.
//
// KPM_REPO replaces "kcl-lang" for every dependency kcl resolves from the
// default registry: short-form entries in kcl.mod (k8s = "1.36"), kcl.mod.lock
// entries that name repo = "kcl-lang/k8s", and -- the reason this is global
// rather than a kcl.mod rewrite -- the short-form dependencies inside packages
// we pull, such as k8s in stuttgart-things/crossplane-provider-helm. A package
// or tag missing from the mirror therefore fails the render: add it to
// mirror/kcl-lang.yaml in stuttgart-things/kcl.
const defaultDependencyRepo = "stuttgart-things/kcl-mirror"

func (m *Kcl) container() *dagger.Container {
	if m.BaseImage == "" {
		m.BaseImage = "kcllang/kcl:v0.12.3"
	}
	if m.DependencyRepo == "" {
		m.DependencyRepo = defaultDependencyRepo
	}

	return dag.Container().
		From(m.BaseImage).
		WithEnvVariable("KPM_REPO", m.DependencyRepo).
		WithExec([]string{"sh", "-c", "apt-get update && apt-get install -y --no-install-recommends jq curl && rm -rf /var/lib/apt/lists/*"}).
		WithExec([]string{"sh", "-c", "curl -sL https://github.com/mikefarah/yq/releases/download/v4.44.1/yq_linux_amd64 -o /usr/local/bin/yq && chmod +x /usr/local/bin/yq"})
}

// ValidateKcl validates KCL configuration files by compiling them
func (m *Kcl) ValidateKcl(ctx context.Context, source *dagger.Directory) (string, error) {
	// Validation by compilation - if files compile without errors, they are valid
	return m.container().
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"sh", "-c", "kcl run main.k > /dev/null && echo 'Validation successful: KCL files are syntactically correct'"}).
		Stdout(ctx)
}

// KclVersion returns the installed KCL version
func (m *Kcl) KclVersion(ctx context.Context) (string, error) {
	return m.container().
		WithExec([]string{"kcl", "version"}).
		Stdout(ctx)
}
