# Boolean Variable Testing for Dagger Templating Module

## Test Summary

This document describes the boolean handling tests for the Dagger templating module.

## Features Tested

### 1. Boolean Parsing from Command-Line Variables ✅

Booleans passed via `--variables` are now properly parsed as Go boolean types (not strings).

**Example:**
```bash
dagger call -m templating render \
  --src=tests/templating \
  --templates="boolean-test.yaml.tmpl" \
  --variables="debug=true,enableLogging=false" \
  export --path=/tmp/output
```

- `true` → boolean `true`
- `false` → boolean `false`
- Other values → remain as strings

### 2. Boolean Handling from YAML Variables File ✅

YAML files naturally support booleans, and these are preserved correctly.

**Example YAML:**
```yaml
debug: false
enableLogging: true
productionMode: true
features:
  logging: true
  monitoring: true
```

**Usage:**
```bash
dagger call -m templating render \
  --src=tests/templating \
  --templates="boolean-test.yaml.tmpl" \
  --variables-file=example-vars.yaml \
  export --path=/tmp/output
```

### 3. Strict Mode for Missing Variables ✅

Added `--strict-mode` flag to control behavior when variables are missing:

- `--strict-mode=false` (default): Missing variables render as `<no value>`
- `--strict-mode=true`: Rendering fails with clear error message

**Example with strict mode:**
```bash
# This will fail if any variable is missing
dagger call -m templating render \
  --src=tests/templating \
  --templates="boolean-test.yaml.tmpl" \
  --strict-mode=true \
  export --path=/tmp/output
```

**Error message when variable is missing:**
```
template rendering failed for boolean-test.yaml.tmpl:
template: boolean-test.yaml.tmpl:5:16: executing "boolean-test.yaml.tmpl"
at <.namespace>: map has no entry for key "namespace"
```

## Test Files

### boolean-test.yaml.tmpl
Tests basic boolean rendering and conditional logic:
- Direct boolean value rendering
- `if/else` conditionals based on boolean values
- Nested boolean values from YAML structures

### boolean-strict-test.yaml.tmpl
Tests strict boolean type comparison using `eq` function:
- Compares variables to boolean literals
- Ensures proper type parsing (bool vs string)

### example-vars.yaml
Updated to include boolean test values:
```yaml
debug: false
enableLogging: true
productionMode: true
features:
  logging: true
  monitoring: true
```

## Test Results

| Test Case | CLI Variables | YAML File | Strict Mode | Result |
|-----------|---------------|-----------|-------------|--------|
| Boolean from CLI | ✅ | - | - | Properly parsed as boolean |
| Boolean from YAML | - | ✅ | - | Properly parsed as boolean |
| Mixed variables | ✅ | ✅ | - | CLI overrides YAML |
| Strict mode with missing vars | - | - | ✅ | Fails with clear error |
| Strict mode with all vars | ✅ | - | ✅ | Renders successfully |
| Lenient mode with missing vars | - | - | ❌ | Renders with `<no value>` |

## Code Changes

### main.go
1. Added boolean type conversion for CLI variables:
   - `"true"` → `bool(true)`
   - `"false"` → `bool(false)`

2. Added `strictMode` parameter to `Render()` function

3. Configured template `missingkey` option:
   - `missingkey=error` when `strictMode=true`
   - `missingkey=default` when `strictMode=false`

## Usage Recommendations

1. **For production deployments:** Use `--strict-mode=true` to catch missing variables early
2. **For development/testing:** Use default mode (strict-mode=false) for flexibility
3. **For boolean values:** Use either CLI variables or YAML - both work correctly
4. **Variable precedence:** CLI variables always override YAML file variables
