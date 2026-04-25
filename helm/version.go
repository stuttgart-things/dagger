package main

import (
	"context"
)

// Version returns the versions of every binary baked into the helm
// container (helm, helmfile, polaris, kubeconform, conftest, vals,
// kubectl). Useful for debugging behavior drift between module
// releases.
func (m *Helm) Version(ctx context.Context) (string, error) {
	script := `
set -e
echo "=== helm ==="        && helm version --short
echo "=== helmfile ==="    && helmfile --version
echo "=== polaris ==="     && polaris version
echo "=== kubeconform ===" && kubeconform -v
echo "=== conftest ==="    && conftest --version | head -1
echo "=== vals ==="        && vals version
echo "=== kubectl ==="     && kubectl version --client 2>&1 | head -2
`
	return m.container().
		WithExec([]string{"sh", "-c", script}).
		Stdout(ctx)
}
