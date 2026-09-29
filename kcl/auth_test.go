package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// fakeKcl writes a script that fails with msg for the first `failures` calls
// and succeeds after that, counting calls in a file.
func fakeKcl(t *testing.T, failures int, msg string) (cmd, counter string) {
	t.Helper()
	dir := t.TempDir()
	counter = filepath.Join(dir, "calls")
	script := filepath.Join(dir, "kcl")
	body := `#!/bin/sh
n=$(cat ` + counter + ` 2>/dev/null || echo 0); n=$((n+1)); echo $n > ` + counter + `
if [ "$n" -le ` + strconv.Itoa(failures) + ` ]; then echo "` + msg + `" >&2; exit 1; fi
echo rendered
`
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return script, counter
}

func runRetry(t *testing.T, cmd string) (exitCode int, stderr string) {
	t.Helper()
	old := renderRetryDelaySeconds
	renderRetryDelaySeconds = 0
	t.Cleanup(func() { renderRetryDelaySeconds = old })

	c := exec.Command("sh", "-c", withRegistryRetry(cmd))
	var errBuf strings.Builder
	c.Stderr = &errBuf
	err := c.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode(), errBuf.String()
	} else if err != nil {
		t.Fatal(err)
	}
	return 0, errBuf.String()
}

func calls(t *testing.T, counter string) int {
	t.Helper()
	b, err := os.ReadFile(counter)
	if err != nil {
		t.Fatal(err)
	}
	n, _ := strconv.Atoi(strings.TrimSpace(string(b)))
	return n
}

// The #391 case: ghcr answers 429 a few times, then lets the pull through.
func TestWithRegistryRetry_RetriesRateLimit(t *testing.T) {
	cmd, counter := fakeKcl(t, 2, `response status code 429: toomanyrequests`)
	code, stderr := runRetry(t, cmd)
	if code != 0 {
		t.Fatalf("want success after two rate limits, exit %d, stderr:\n%s", code, stderr)
	}
	if n := calls(t, counter); n != 3 {
		t.Fatalf("want 3 attempts, got %d", n)
	}
	if !strings.Contains(stderr, "toomanyrequests") || !strings.Contains(stderr, "retrying") {
		t.Fatalf("the rate limit and the retry must be visible in stderr:\n%s", stderr)
	}
}

func TestWithRegistryRetry_GivesUpAfterAllAttempts(t *testing.T) {
	cmd, counter := fakeKcl(t, 99, `response status code 429: toomanyrequests`)
	if code, _ := runRetry(t, cmd); code == 0 {
		t.Fatal("a permanent rate limit must fail")
	}
	if n := calls(t, counter); n != renderAttempts {
		t.Fatalf("want %d attempts, got %d", renderAttempts, n)
	}
}

// Anything but a rate limit is a real render error: no retry.
func TestWithRegistryRetry_OtherErrorsFailAtOnce(t *testing.T) {
	cmd, counter := fakeKcl(t, 1, `error[E2L23]: CompileError: undefined variable`)
	code, stderr := runRetry(t, cmd)
	if code == 0 {
		t.Fatal("a compile error must fail")
	}
	if n := calls(t, counter); n != 1 {
		t.Fatalf("a compile error must not be retried, got %d attempts", n)
	}
	if !strings.Contains(stderr, "CompileError") {
		t.Fatalf("kcl's own message must be passed through:\n%s", stderr)
	}
}

func TestRegistryHost(t *testing.T) {
	for in, want := range map[string]string{
		"ghcr.io/stuttgart-things/x-kustomize": "ghcr.io",
		"ttl.sh/abc":                           "ttl.sh",
		"registry.example.com:5000/a/b":        "registry.example.com:5000",
		"noslash":                              "ghcr.io",
	} {
		if got := registryHost(in); got != want {
			t.Errorf("registryHost(%q) = %q, want %q", in, got, want)
		}
	}
}
