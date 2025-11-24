# TEMPLATING

This module provides template rendering capabilities using Go's text/template package.
It supports loading templates from local files or HTTPS URLs, and accepts variables from YAML files or command-line parameters.

## Features

- ✅ **Boolean Type Support**: CLI variables `true`/`false` are parsed as Go booleans (not strings)
- ✅ **Strict Mode**: Optional flag to fail on missing variables instead of rendering `<no value>`
- ✅ **Variable Precedence**: CLI variables override YAML file variables
- ✅ **Remote Templates**: Support for HTTPS URLs
- ✅ **Multiple Templates**: Render multiple templates in a single call

## Basic Usage

```bash
dagger call -m templating render \
  --src=tests/templating \
  --templates="config.yaml.tmpl,deployment.yaml.tmpl,service.yaml.tmpl" \
  --variables="name=myapp,namespace=default,version=1.0.0,replicas=3,image=nginx,port=8080,serviceType=ClusterIP,debug=false" \
  --variables-file=example-vars.yaml \
  export --path=./rendered
```

## Strict Mode (Recommended for Production)

Fail fast when variables are missing:

```bash
dagger call -m templating render \
  --src=tests/templating \
  --templates="config.yaml.tmpl" \
  --variables-file=vars.yaml \
  --strict-mode=true \
  export --path=./rendered
```

## Render w/ complex structures

```bash
# data.yaml
---
database:
  host: "db.example.com"
  port: 5432
  name: "myapp_db"
  username: "admin"

features:
  - authentication
  - logging
  - monitoring
```

```bash
# config.yaml.tmpl
---
data:
  app.yaml: |
    database:
      host: {{.database.host}}
      port: {{.database.port}}
      name: {{.database.name}}
      username: {{.database.username}}
    features:
    {{- range .features}}
      - {{.}}
    {{- end}}
```

```bash
dagger call render-from-file \
--src ../tests/templating \
--templates config.yaml.tmpl \
--data-file data.yaml \
export --path ./output
```

## Boolean Variables

Booleans work from both CLI and YAML:

**CLI:**
```bash
--variables="debug=true,enableLogging=false"
```

**YAML:**
```yaml
debug: true
enableLogging: false
```

**Template:**
```yaml
{{- if .debug }}
log_level: debug
{{- else }}
log_level: info
{{- end }}
```

See [tests/templating/BOOLEAN_TESTS.md](tests/templating/BOOLEAN_TESTS.md) for detailed boolean testing documentation.
