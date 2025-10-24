
# Linting Module

Lint YAML and Markdown files for best practices and syntax errors using Dagger.

## 🚀 Quick Start

### Prerequisites
- Dagger CLI ([Installation](https://docs.dagger.io/install))
- Docker


### YAML Lint Example

Lint all YAML files in the test folder and export the report:

```bash
dagger call -m linting lint-yaml --src tests/linting/yaml/ export --path=/tmp/report.yaml
```

- `--src tests/linting/yaml/` selects the folder with YAML files
- `export --path=/tmp/report.yaml` saves the result as a report

### Markdown Lint Example

Lint all Markdown files in the test folder and export the findings:

```bash
dagger call -m linting lint-markdown --src tests/linting/markdown/ export --path=/tmp/markdown-findings.txt
```

- `--src tests/linting/markdown/` selects the folder with Markdown files
- `export --path=/tmp/markdown-findings.txt` saves the findings as a text file

## 📂 Test Data

Example test data can be found in:
- `tests/linting/yaml/valid.yaml`
- `tests/linting/yaml/invalid.yaml`

## 📖 More Modules

See the main README for more Dagger modules and examples.
