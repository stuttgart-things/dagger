package main

import (
	"context"
	"dagger/kyverno/internal/dagger"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Kyverno struct {
	// +optional
	// +default="cgr.dev/chainguard/wolfi-base:latest"
	BaseImage string
}

// defaultKyvernoVersion is the CLI release used when kyvernoVersion is empty.
// Keep it in step with the +default annotations below, which have to be
// literals.
const defaultKyvernoVersion = "1.19.1"

func (m *Kyverno) Validate(
	ctx context.Context,
	policy *dagger.Directory,
	resource *dagger.Directory,
	// Kyverno CLI release; match it to the Kyverno running on the cluster
	// +optional
	// +default="1.19.1"
	kyvernoVersion string,
) error {
	kyverno := m.container(kyvernoVersion).
		WithMountedDirectory("/policy", policy).
		WithMountedDirectory("/resource", resource).
		WithWorkdir("/")

	result, err := kyverno.
		WithExec([]string{
			"kubectl-kyverno",
			"apply",
			"/policy",
			"--resource",
			"/resource"}).
		Stdout(ctx)

	if err != nil {
		return fmt.Errorf("failed to validate: %w", err)
	}

	fmt.Println(result)
	return nil
}

// Test runs `kyverno test` over a directory holding kyverno-test.yaml files and
// the policies and resources they name, and returns the result table.
//
// Validate answers "do these resources pass these policies?". Test answers the
// question every policy change raises: does the policy still refuse what it
// must refuse, and still admit what it must admit? A policy that admits
// everything passes Validate against good resources forever; it fails a test
// that expects a refusal.
//
// Any result that does not match its expectation is an error, and so is a path
// without a single kyverno-test.yaml (--require-tests): a run over the wrong
// directory must not pass by testing nothing. The error carries the table.
//
// Policies that verify image signatures reach the registry and Rekor during
// the test, so their results depend on what is published at the time: a
// signed fixture image has to stay in its registry. For the same reason the
// test is never answered from cache.
// +cache="never"
func (m *Kyverno) Test(
	ctx context.Context,
	// Directory holding kyverno-test.yaml files and the files they name
	src *dagger.Directory,
	// Path inside src to search for kyverno-test.yaml files
	// +optional
	// +default="."
	path string,
	// Kyverno CLI release; match it to the Kyverno running on the cluster
	// +optional
	// +default="1.19.1"
	kyvernoVersion string,
	// Fail on deprecation warnings, such as the one every kyverno.io/v1
	// ClusterPolicy draws from 1.19 on. Needs a CLI of 1.19 or later.
	// +optional
	// +default=false
	warningsAsErrors bool,
) (string, error) {
	cmd := []string{"kubectl-kyverno", "test", path, "--require-tests", "--remove-color"}
	if warningsAsErrors {
		cmd = append(cmd, "--warnings-as-errors")
	}

	out, err := m.container(kyvernoVersion).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		// CACHE_BUSTER forces a fresh run: a verdict on an image signature
		// is only as current as the registry state it was checked against.
		WithEnvVariable("CACHE_BUSTER", strconv.FormatInt(time.Now().UnixNano(), 10)).
		WithExec(cmd).
		Stdout(ctx)
	if err != nil {
		var execErr *dagger.ExecError
		if errors.As(err, &execErr) {
			return "", fmt.Errorf("kyverno test failed: %s\n%s",
				strings.TrimSpace(execErr.Stderr), execErr.Stdout)
		}
		return "", fmt.Errorf("kyverno test failed: %w", err)
	}

	return out, nil
}

func (m *Kyverno) Version(
	ctx context.Context,
	// Kyverno CLI release
	// +optional
	// +default="1.19.1"
	kyvernoVersion string,
) (version string) {
	kyverno := m.container(kyvernoVersion)

	cmd := []string{"kubectl-kyverno", "version"}

	version, err := kyverno.WithExec(cmd).Stdout(ctx)
	if err != nil {
		fmt.Println("Error running kyverno version: ", err)
		return
	}
	fmt.Println("Kyverno version: ", version)

	return version
}

// container returns BaseImage with the Kyverno CLI of the given release.
//
// The binary is copied out of Kyverno's own release image rather than
// installed with apk. Wolfi ships one current CLI, so the version used to be
// whatever it carried on the day the container was built -- and 1.19 refuses
// policies that earlier releases accepted. Every release has an image; not
// every release is in Wolfi.
func (m *Kyverno) container(kyvernoVersion string) *dagger.Container {
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	cli := dag.Container().
		From(kyvernoImage(kyvernoVersion)).
		File("/ko-app/kubectl-kyverno")

	return dag.Container().From(m.BaseImage).
		WithFile("/usr/bin/kubectl-kyverno", cli, dagger.ContainerWithFileOpts{Permissions: 0o755}).
		WithEntrypoint([]string{"kubectl-kyverno"})
}

// kyvernoImage returns the image of a Kyverno CLI release. Its tags are
// v-prefixed; a bare "1.19.1" is accepted as well.
func kyvernoImage(version string) string {
	if version == "" {
		version = defaultKyvernoVersion
	}
	if version[0] >= '0' && version[0] <= '9' {
		version = "v" + version
	}
	return "ghcr.io/kyverno/kyverno-cli:" + version
}
