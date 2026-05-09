# Templating Dagger Module

Render Go `text/template` files with variables from YAML, JSON, or CLI key=value pairs. Templates and data files can live in a local directory or be fetched over HTTPS. [sprig](https://masterminds.github.io/sprig/) functions are available in every template.

## Functions

| Function | Purpose |
|----------|---------|
| `render` | Render one or more templates using `--variables` (CLI) and/or `--variables-file` (YAML). CLI vars win. |
| `render-from-file` | Render one or more templates using a YAML or JSON `--data-file`. |
| `render-inline` | Render a template passed as a string and return the rendered string. |

`.tmpl` is stripped from output filenames automatically (e.g. `values.yaml.tmpl` → `values.yaml`). With `--strict-mode`, missing keys fail; otherwise they render as `<no value>`.

## Quick Start

```bash
# Render with CLI vars
dagger call -m templating render \
  --src tests/templating \
  --templates configmap.yaml.tmpl \
  --variables "name=patrick,env=dev" \
  export --path /tmp/rendered
```

```bash
# Render with a YAML/JSON data file (CLI vars override)
dagger call -m templating render \
  --src tests/templating \
  --templates "configmap.yaml.tmpl,deployment.yaml.tmpl" \
  --variables-file values.yaml \
  --variables "image_tag=v2" \
  export --path /tmp/rendered
```

```bash
# Render templates and data file from HTTPS URLs (no --src needed)
dagger call -m templating render-from-file \
  --templates https://example.com/tmpl/values.yaml.tmpl \
  --data-file https://example.com/data/prod.yaml \
  export --path /tmp/rendered
```

```bash
# One-shot inline render (returns string)
dagger call -m templating render-inline \
  --template-data 'hello {{ .name | upper }}' \
  --variables '{"name":"patrick"}'
```
