# Linting Module

Lint YAML, Markdown files, and run pre-commit hooks for best practices and syntax errors using Dagger.

## 🚀 Quick Start

### Prerequisites
- Dagger CLI ([Installation](https://docs.dagger.io/install))
- Docker

### YAML Lint Example

Lint all YAML files in the test folder and export the report:
```bash
dagger call -m linting lint-yaml --src tests/linting/yaml/ export --path=/tmp/yaml-findings.txt
```

- `--src tests/linting/yaml/` selects the folder with YAML files
- `export --path=/tmp/yaml-findings.txt` saves the result as a report

### Markdown Lint Example

Lint all Markdown files in the test folder and export the findings:
```bash
dagger call -m linting lint-markdown --src tests/linting/markdown/ export --path=/tmp/markdown-findings.txt
```

- `--src tests/linting/markdown/` selects the folder with Markdown files
- `export --path=/tmp/markdown-findings.txt` saves the findings as a text file

### Pre-Commit Hooks Example

Run pre-commit hooks on your repository and export the findings:
```bash
dagger call -m linting lint-pre-commit --src . export --path=/tmp/precommit-findings.txt
```

- `--src .` runs pre-commit on the current directory
- `export --path=/tmp/precommit-findings.txt` saves the findings as a text file

#### Custom Pre-Commit Config

Use a custom pre-commit configuration file:
```bash
dagger call -m linting lint-pre-commit --src . --config-path .pre-commit-config.yaml export --path=/tmp/precommit-findings.txt
```

#### Skip Specific Hooks

Skip hooks that require Docker or other unavailable resources:
```bash
dagger call -m linting lint-pre-commit --src . --skip-hooks hadolint-docker export --path=/tmp/precommit-findings.txt
```

Skip multiple hooks:
```bash
dagger call -m linting lint-pre-commit --src . --skip-hooks hadolint-docker --skip-hooks another-hook export --path=/tmp/precommit-findings.txt
```

### Secret Scanning Example

Scan for secrets in your repository and export the JSON findings:
```bash
dagger call -m linting scan-secrets --src . export --path=/tmp/secret-findings.json
```

- `--src .` scans the current directory
- `export --path=/tmp/secret-findings.json` saves the detect-secrets JSON report

#### Exclude Files

Exclude files matching a regex pattern:
```bash
dagger call -m linting scan-secrets --src . --exclude-files "\.git|node_modules" export --path=/tmp/secret-findings.json
```

### JSON Schema Syntax Validation

Run `jq empty` over every matching `*.schema.json` under the source tree — a
fast syntax gate that fails CI if any schema file won't parse:
```bash
dagger call -m linting validate-json-schema --src tests/linting/schemas/
```

- `--src` — directory to recurse into
- `--glob` — optional basename pattern (default `**/*.schema.json`; the leading
  `**/` is stripped since `find` already recurses). Example:
  `--glob '*.json'` to check every JSON file.

Exits non-zero on the first parse failure; stdout lists every file checked.

The module's `test-linting` task exercises local happy + regression fixtures;
`test-linting-schemas SCHEMAS_SRC=/path/to/tree` runs the same function against
an external source tree (defaults to `../argocd`).

### AI-Powered Secret Auto-Fix

Use the AI agent to automatically add `pragma: allowlist secret` comments to flagged lines:
```bash
dagger call -m linting auto-fix-secrets --src . export --path=/tmp/fixed/
```

- Analyzes detect-secrets findings and adds inline pragma comments
- Uses correct comment syntax per file type (`//` for Go, `#` for YAML/Python/Shell/Terraform)
- Skips `.json` files and lines that already have the pragma

## 📂 Test Data

Example test data can be found in:
- `tests/linting/yaml/valid.yaml`
- `tests/linting/yaml/invalid.yaml`
- `tests/linting/markdown/` - Markdown test files
- `tests/linting/schemas/` - valid JSON schema fixtures (happy path)
- `tests/linting/schemas-invalid/` - malformed schema (regression path)
- `.pre-commit-config.yaml` - Pre-commit configuration

## 🔧 Supported Pre-Commit Hooks

The module supports common pre-commit hooks including:
- trailing-whitespace
- end-of-file-fixer
- check-added-large-files
- check-merge-conflict
- check-yaml
- detect-private-key
- shellcheck
- hadolint (use `hadolint` instead of `hadolint-docker`)
- check-github-workflows
- detect-secrets

## 📖 More Modules

See the main README for more Dagger modules and examples.
