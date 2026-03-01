# KCL Module Tests

This directory contains test files and examples for the KCL Dagger module located in `/kcl/`.

## Test Files

- `test-crd.yaml` - Example CRD for testing CRD-to-KCL conversion functionality
- `test-kcl-project/` - Sample KCL project for testing basic KCL operations
  - `main.k` - Simple KCL configuration example
- `test-kustomize-project/` - KCL project for testing kustomize base rendering
  - `main.k` - Produces a Deployment and Service as multi-document YAML
  - `params.yaml` - Test parameters (name, namespace, image, port, replicas)

## Running Tests

All tests are executed via the main Taskfile in the repository root:

```bash
# Run all KCL tests (includes CRD conversion)
task test-kcl

# Test kustomize base rendering (render + verify output files)
task test-kcl-kustomize

# Render kustomize base with custom source
task render-kustomize-base SOURCE=./my-project PROFILE=./my-params.yaml

# Convert CRD using web source
task convert-crd CRD_URL=https://raw.githubusercontent.com/crossplane-contrib/provider-terraform/main/package/crds/tf.upbound.io_workspaces.yaml

# Convert local CRD file
task convert-crd CRD_FILE=tests/kcl/test-crd.yaml
```

## Module Structure

- **Module Location**: `/kcl/` (main Dagger module)
- **Tests Location**: `/tests/kcl/` (test files and examples)
- **Taskfile Integration**: Main `Taskfile.yaml` in repository root

This follows Stuttgart-Things standards where modules are in the root and tests are under `tests/<module-name>/`.
