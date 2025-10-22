# KCL Dagger Module

This Dagger module provides comprehensive KCL (KCL Configuration Language) functionality for cloud-native configuration management and automated CRD-to-KCL conversion workflows.

## Features

- ✅ **KCL CLI Integration**: Run KCL configurations in containerized environments
- ✅ **Configuration Validation**: Syntax and semantic validation of KCL files
- ✅ **CRD to KCL Conversion**: Automatic conversion of Kubernetes CRDs to type-safe KCL schemas
- ✅ **Ubuntu 24.04 Base**: Uses official Ubuntu LTS with KCL CLI installation
- ✅ **Web & Local Sources**: Support for both remote URLs and local files
- ✅ **Project Directory Support**: Mount and process entire KCL projects

## Quick Start

### Basic KCL Operations

```bash
# Test KCL installation and basic functionality
dagger call -m kcl test-kcl

# Get KCL version
dagger call -m kcl kcl-version

# Run KCL configuration from directory
dagger call -m kcl run-kcl --source ./my-kcl-project --entrypoint main.k

# Validate KCL configuration
dagger call -m kcl validate-kcl --source ./my-kcl-project
```

### CRD to KCL Conversion

```bash
# Convert CRD from web source (Terraform Provider example)
dagger call -m kcl convert-crd \
  --crd-source "https://raw.githubusercontent.com/crossplane-contrib/provider-terraform/main/package/crds/tf.upbound.io_workspaces.yaml" \
  export --path=./generated-models

# Convert local CRD file
dagger call -m kcl convert-crd \
  --crd-file ./my-crd.yaml \
  export --path=./generated-models
```

## Functions

### Core KCL Functions

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `KclVersion()` | Get installed KCL version | None | Version string |
| `TestKcl()` | Basic functionality test | None | Test output |
| `RunKcl()` | Execute KCL configuration | `source`, `entrypoint` | KCL output |
| `ValidateKcl()` | Validate KCL syntax | `source` | Validation result |

### CRD Conversion Functions

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `ConvertCrd()` | Convert single CRD to KCL models | `crdSource?`, `crdFile?` | Generated models directory |
| `ConvertCrdToDirectory()` | Convert CRD with custom output structure | `workdir`, `crdSource?`, `crdFile?`, `outputPath?` | Updated directory |

## CRD to KCL Conversion

### Overview
The `ConvertCRD` functionality enables automatic conversion of Kubernetes Custom Resource Definitions (CRDs) into type-safe KCL schemas. This is particularly useful for creating KCL modules for Crossplane providers.

### Usage Options

#### Via Taskfile (Recommended)

```bash
# Convert CRD from web source
task convert-crd CRD_URL=https://raw.githubusercontent.com/crossplane-contrib/provider-terraform/main/package/crds/tf.upbound.io_workspaces.yaml

# Convert local CRD file
task convert-crd CRD_FILE=./my-crd.yaml
```

#### Direct Dagger Usage

```bash
# Web source
dagger call -m kcl convert-crd \
  --crd-source "https://raw.githubusercontent.com/crossplane-contrib/provider-terraform/main/package/crds/tf.upbound.io_workspaces.yaml" \
  export --path=./generated-models

# Local file
dagger call -m kcl convert-crd \
  --crd-file ./my-crd.yaml \
  export --path=./generated-models

# Custom output directory
dagger call -m kcl convert-crd-to-directory \
  --workdir . \
  --crd-file ./my-crd.yaml \
  --output-path models \
  export --path=./updated-project
```

### Generated Structure

After conversion, the following structure is created:

```
models/
├── kcl.mod                                    # KCL module definition
├── k8s/                                       # Kubernetes core types
│   └── apimachinery/pkg/apis/meta/v1/
│       ├── managed_fields_entry.k
│       ├── object_meta.k
│       └── owner_reference.k
└── v1beta1/                                   # Generated CRD schema
    └── <crd-name>_v1beta1_<resource>.k        # Main schema
```

### Example Output

When converting a Terraform Provider CRD:

```kcl
# v1beta1/tf_upbound_io_v1beta1_workspace.k
schema Workspace:
    r"""
    A Workspace of Terraform Configuration.
    """
    apiVersion: "tf.upbound.io/v1beta1" = "tf.upbound.io/v1beta1"
    kind: "Workspace" = "Workspace"
    metadata?: v1.ObjectMeta
    spec: TfUpboundIoV1beta1WorkspaceSpec
    status?: TfUpboundIoV1beta1WorkspaceStatus
```

### Integration in KCL Modules

Generated models can be directly used in KCL modules:

```kcl
# main.k
import models.v1beta1.tf_upbound_io_v1beta1_workspace as workspace

# Simplified wrapper schema
schema TerraformWorkspace:
    name: str
    module: str
    variables?: {str: str}

# Helper function
generateWorkspace = lambda config: TerraformWorkspace -> [workspace.Workspace] {
    [
        workspace.Workspace {
            metadata = {name = config.name}
            spec = {
                forProvider = {
                    source = "Remote"
                    module = config.module
                    vars = [{key = k, value = v} for k, v in config.variables] if config.variables else Undefined
                }
            }
        }
    ]
}
```

## Container Configuration

- **Base Image**: `ubuntu:24.04`
- **KCL Installation**: Official KCL CLI installation script
- **Required Packages**: `curl`, `wget`, `git`, `ca-certificates`
- **Entrypoint**: `kcl`

## Taskfile Integration

Use the convenient Taskfile tasks for common operations:

```bash
# Test all KCL functionality
task test-kcl

# Convert CRD from web source
task convert-crd CRD_URL=https://raw.githubusercontent.com/crossplane-contrib/provider-terraform/main/package/crds/tf.upbound.io_workspaces.yaml

# Convert local CRD file
task convert-crd CRD_FILE=./my-crd.yaml
```

## Project Structure

```
kcl/                             # Main KCL Dagger module
├── README.md                    # This documentation
├── main.go                      # Main Dagger module implementation
├── container.go                 # Container configuration
├── dagger.json                  # Dagger module configuration
└── internal/                    # Generated Dagger types

tests/kcl/                       # Test files and examples
├── README.md                    # Test documentation
├── test-crd.yaml               # Example CRD for testing
└── test-kcl-project/           # Example KCL project
    └── main.k                  # Sample KCL configuration
```

## Use Cases

### 1. Configuration Management
```bash
# Validate configurations before deployment
dagger call -m kcl validate-kcl --source ./configs

# Render configurations for different environments
dagger call -m kcl run-kcl --source ./configs --entrypoint prod.k
```

### 2. Crossplane Module Development
```bash
# Convert Crossplane provider CRDs to KCL schemas
task convert-crd CRD_URL=https://raw.githubusercontent.com/crossplane-contrib/provider-terraform/main/package/crds/tf.upbound.io_workspaces.yaml

# Use generated schemas in KCL modules
dagger call -m kcl run-kcl --source ./crossplane-module
```

### 3. CI/CD Integration
```bash
# In your CI pipeline
dagger call -m kcl validate-kcl --source .
dagger call -m kcl run-kcl --source . --entrypoint deployment.k | kubectl apply -f -
```

## Best Practices

### CRD Conversion
1. **Use Latest CRDs**: Always fetch CRDs directly from upstream repositories
2. **Create Wrapper Schemas**: Build simplified schemas for common use cases
3. **Write Helper Functions**: Implement utility functions for common configuration patterns
4. **Add Tests**: Test generated schemas with practical examples

### General Development
1. **Follow Existing Patterns**: Maintain consistency with existing function patterns
2. **Comprehensive Testing**: Add tests for new functionality
3. **Update Documentation**: Keep examples and guides current
4. **Taskfile Integration**: Ensure new tasks are properly integrated

## Examples

### Simple KCL Configuration
```kcl
# test-kcl-project/main.k
app = "nginx"
version = "1.21"
port = 80
replicas = 3
```

### CRD Conversion Workflow
1. Download or reference a CRD
2. Convert to KCL schemas using `ConvertCrd`
3. Create wrapper schemas for simplified usage
4. Build helper functions for common patterns
5. Test with practical examples

## Troubleshooting

### Error: "No CRD source provided"
- Ensure either `CRD_URL` or `CRD_FILE` is specified

### Error: "wget: command not found"
- The container installs `wget` automatically - check network connectivity

### Error: "kcl import failed"
- Verify the CRD file contains valid YAML
- Ensure it's an actual CRD (not a regular Kubernetes resource)

### Comparison with Manual Conversion

**Before (manual):**
```bash
# Multiple error-prone steps
wget -O crd.yaml https://example.com/crd.yaml
kcl import -m crd crd.yaml
# Manual cleanup and structuring
```

**After (automated):**
```bash
# Single command, everything handled
task convert-crd CRD_URL=https://example.com/crd.yaml
```

## Advanced Usage

### Custom Container Configuration
```go
// Override base image if needed
kcl := dag.Kcl().WithBaseImage("custom:image")
```

### Batch CRD Processing
```bash
# Process multiple CRDs
for crd in *.yaml; do
  dagger call -m kcl convert-crd --crd-file "$crd" export --path="./models-$(basename $crd .yaml)"
done
```

## Development Standards

This module follows Stuttgart-Things development standards:

- ✅ **Repository Structure**: Module in root `/kcl/`, tests in `/tests/kcl/`
- ✅ **Taskfile Integration**: Automated tasks for all functions
- ✅ **Testing Requirements**: Comprehensive test coverage
- ✅ **Documentation**: Detailed function and usage documentation
- ✅ **Container Standards**: Ubuntu 24.04 base with official installation scripts
- ✅ **Error Handling**: Robust error handling and validation## Related Documentation

- [KCL Official Documentation](https://kcl-lang.io/) - KCL language reference
- [Crossplane Documentation](https://crossplane.io/docs/) - For Crossplane provider CRDs
- [Stuttgart-Things Standards](../../.container-use/decisions.md) - Development standards and patterns

## Contributing

When extending this module:
1. Follow the existing function patterns
2. Add comprehensive tests for new functionality
3. Update documentation and examples
4. Ensure Taskfile integration for new tasks
5. Test with both local and remote sources

This task automates the complete CRD-to-KCL conversion process and follows Stuttgart-Things standards for KCL modules.