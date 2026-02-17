package main

import (
	"context"
	"dagger/linting/internal/dagger"
)

// ScanSecrets runs detect-secrets scan on the provided directory and returns a JSON findings report.
func (m *Linting) ScanSecrets(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	// +default="secret-findings.json"
	outputFile string,
	// +optional
	excludeFiles string,
) (*dagger.File, error) {
	if outputFile == "" {
		outputFile = "secret-findings.json"
	}

	ctr := m.container().
		WithMountedDirectory("/mnt", src).
		WithWorkdir("/mnt")

	cmd := "detect-secrets scan --all-files ."
	if excludeFiles != "" {
		cmd += " --exclude-files '" + excludeFiles + "'"
	}
	cmd += " > /tmp/" + outputFile + " 2>&1 || true"

	ctr = ctr.WithExec([]string{"sh", "-c", cmd})

	return ctr.File("/tmp/" + outputFile), nil
}

// AutoFixSecrets uses a Dagger AI agent to analyze detect-secrets findings
// and add pragma: allowlist secret comments to flagged lines.
func (m *Linting) AutoFixSecrets(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	excludeFiles string,
) (*dagger.Directory, error) {
	ctr := m.container().
		WithMountedDirectory("/mnt", src).
		WithWorkdir("/mnt")

	// Run detect-secrets scan to produce findings
	scanCmd := "detect-secrets scan --all-files ."
	if excludeFiles != "" {
		scanCmd += " --exclude-files '" + excludeFiles + "'"
	}
	scanCmd += " > /tmp/findings.json 2>&1 || true"

	ctr = ctr.WithExec([]string{"sh", "-c", scanCmd})

	environment := dag.Env(dagger.EnvOpts{Privileged: true}).
		WithContainerInput("workspace", ctr,
			"Container with source code at /mnt and detect-secrets findings at /tmp/findings.json. "+
				"The findings JSON has a 'results' object keyed by file path, each containing an array of findings with 'line_number' and 'type' fields.").
		WithContainerOutput("result",
			"The same container with 'pragma: allowlist secret' comments appended to the end of each flagged line in the source files under /mnt. "+
				"Do not modify the findings JSON file.")

	prompt := `You are a security code reviewer. Your task is to read /tmp/findings.json and add inline pragma comments to suppress false-positive secret findings.

Instructions:
1. Read /tmp/findings.json to get the list of flagged files and line numbers.
2. For each finding, open the flagged file under /mnt and append the appropriate pragma comment to the END of the flagged line.
3. Use the correct comment syntax based on file extension:
   - .go files: // pragma: allowlist secret
   - .yaml, .yml, .py, .sh, .tf, .toml files: # pragma: allowlist secret
   - .json files: SKIP (JSON does not support comments)
4. If a line already contains "pragma: allowlist secret", do NOT add it again.
5. Do not modify any other lines. Only append the pragma comment to flagged lines.
6. Make sure to preserve the original file content and indentation exactly.

Process all findings and return the modified container as the result.`

	work, err := dag.LLM().
		WithEnv(environment).
		WithPrompt(prompt).
		Loop().
		Sync(ctx)
	if err != nil {
		return nil, err
	}

	return work.Env().Output("result").AsContainer().Directory("/mnt"), nil
}
