package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"dagger/kcl/internal/dagger"
)

// registryAuth carries the credentials a caller already hands a push, so the
// render in front of it can use them too.
//
// kcl pulls its dependencies (ghcr.io/kcl-lang/k8s and friends) while it
// renders. Anonymously, ghcr throttles those pulls: on 2026-09-29 every
// push-kustomize-base failed with
//
//	failed to resolve 1.31: GET "https://ghcr.io/v2/kcl-lang/k8s/manifests/1.31":
//	response status code 429: toomanyrequests
//
// for twenty minutes straight, in CI and locally, although the push that
// followed had credentials all along (stuttgart-things/dagger#391).
type registryAuth struct {
	registry     string
	userName     string
	user         *dagger.Secret
	passwordName string
	password     *dagger.Secret
}

// login runs `kcl registry login` in ctr. A nil auth, or one without both
// secrets, leaves ctr unchanged and the render anonymous, as before.
func (a *registryAuth) login(ctr *dagger.Container) *dagger.Container {
	if a == nil || a.user == nil || a.password == nil {
		return ctr
	}
	return ctr.
		WithSecretVariable(a.userName, a.user).
		WithSecretVariable(a.passwordName, a.password).
		WithExec([]string{"sh", "-c", fmt.Sprintf(`kcl registry login %s -u "$%s" -p "$%s"`, a.registry, a.userName, a.passwordName)})
}

// registryHost returns the host part of an OCI address ("ghcr.io/org/repo" ->
// "ghcr.io"), defaulting to ghcr.io.
func registryHost(address string) string {
	if i := strings.Index(address, "/"); i > 0 {
		return address[:i]
	}
	return "ghcr.io"
}

// renderRetryDelaySeconds is the backoff unit between render attempts; tests
// set it to zero.
var renderRetryDelaySeconds = 5

// renderAttempts is how often a rate-limited render is tried in total. Eight
// with the linear backoff wait up to ~2 min: the 429s of 2026-09-29 were
// intermittent per request (the same manifest answered 200 and 429 within a
// minute, with or without a login), so persistence is what gets through.
const renderAttempts = 8

// withRegistryRetry wraps a kcl command so that a registry rate limit (HTTP
// 429 / toomanyrequests) is retried with a linear backoff. Any other failure
// fails at once, and the command's stderr is always passed through -- kcl
// itself does not retry, and ghcr's retry-after is in the milliseconds.
func withRegistryRetry(cmd string) string {
	return fmt.Sprintf(`attempt=1
while :; do
  if %[1]s 2>/tmp/kcl-run.err; then cat /tmp/kcl-run.err >&2; exit 0; fi
  cat /tmp/kcl-run.err >&2
  if [ "$attempt" -ge %[2]d ] || ! grep -qE '429|toomanyrequests' /tmp/kcl-run.err; then exit 1; fi
  echo "registry rate limit, retrying in $((attempt * %[3]d))s (attempt $attempt of %[2]d)" >&2
  sleep $((attempt * %[3]d))
  attempt=$((attempt + 1))
done`, cmd, renderAttempts, renderRetryDelaySeconds)
}

// renderError turns a failed kcl render into an error that says what kcl
// said. Without it the failure surfaced lazily, as `! exit code: 1` on the
// `oras push` of PushKustomizeBase, with the actual cause (the 429) visible
// only with --progress plain.
func renderError(err error) error {
	var execErr *dagger.ExecError
	if errors.As(err, &execErr) {
		if msg := strings.TrimSpace(execErr.Stderr); msg != "" {
			return fmt.Errorf("kcl run failed: %s", msg)
		}
	}
	return fmt.Errorf("kcl run failed: %w", err)
}

// syncRender evaluates the render container now, so a failure is reported by
// the render and not by whatever consumes its output later.
func syncRender(ctx context.Context, ctr *dagger.Container) error {
	if _, err := ctr.Sync(ctx); err != nil {
		return renderError(err)
	}
	return nil
}
