package main

import (
	"context"
	"dagger/linting/internal/dagger"
	"strings"
)

// ValidateJsonSchema runs `jq empty` over every file under src whose basename
// matches glob. Acts as a fast JSON-syntax gate for *.schema.json files.
// Exits non-zero if any matched file fails to parse.
// Exported as ValidateJsonSchema (not ValidateJSONSchema) so Dagger's name
// normalizer produces the kebab form `validate-json-schema` rather than
// `validate-jsonschema` — keeps the CLI call matching the issue's spec.
func (m *Linting) ValidateJsonSchema(
	ctx context.Context,
	src *dagger.Directory,
	// Basename glob passed to `find -name`. A leading `**/` is stripped since
	// find already recurses by default.
	// +optional
	// +default="**/*.schema.json"
	glob string,
) (string, error) {
	if glob == "" {
		glob = "**/*.schema.json"
	}
	pattern := strings.TrimPrefix(glob, "**/")

	script := `set -e
MATCHED=$(find . -type f -name "$GLOB" | wc -l)
echo "ValidateJsonSchema: glob=$GLOB matched=$MATCHED"
if [ "$MATCHED" -eq 0 ]; then
  echo "no files matched — nothing to validate"
  exit 0
fi
find . -type f -name "$GLOB" -print0 | xargs -0 -I{} sh -c 'echo "  check {}"; jq empty "{}"'
echo "all $MATCHED schema(s) parsed OK"
`

	return m.container().
		WithMountedDirectory("/mnt", src).
		WithWorkdir("/mnt").
		WithEnvVariable("GLOB", pattern).
		WithExec([]string{"sh", "-c", script}).
		Stdout(ctx)
}
