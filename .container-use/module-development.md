# Module Development Guide - Stuttgart-Things Dagger

This guide provides comprehensive instructions for developing new Dagger modules following Stuttgart-Things standards.

## 🏗️ Module Development Overview

### Architecture Decisions

All modules follow established patterns documented in [decisions.md](./decisions.md):

- **Repository Structure**: Modules in root, tests in `tests/{module}/`
- **Testing Standards**: Comprehensive test functions for all features
- **Container Standards**: Proven base images with stable installations
- **Documentation Requirements**: Complete README.md and code documentation

## 🚀 Creating a New Module

### Step 1: Interactive Module Creation

```bash
task create
```

**Interactive Wizard:**
1. **Module Name Input**: Enter the module name (e.g., `python`, `rust`, `terraform`)
2. **SDK Selection**: Choose between `go` or `python`
3. **Automatic Setup**: Creates module structure and dagger.json

**Example:**
```bash
$ task create
? Module name: › python
? SDK:
  > go
    python
```

### Step 2: Module Structure Setup

**Generated Structure:**
```
{module-name}/
├── dagger.json          # Dagger module configuration
├── main.go              # Main module implementation (for Go)
├── go.mod               # Go dependencies (for Go)
├── go.sum               # Go dependencies lock (for Go)
├── requirements.txt     # Python dependencies (for Python)
└── README.md            # Module documentation
```

### Step 3: Implement Core Functions

#### Required Functions (Go Example)

```go
package main

import (
    "context"
    "dagger/module-name/internal/dagger"
)

// ModuleName provides functionality for working with {tool}
//
// The module includes functions for:
// - Version checking and tool validation
// - Basic functionality testing
// - Advanced operations like validation and execution
//
// Example usage:
//   dagger call -m {module-name} {module-name}-version
//   dagger call -m {module-name} test-{module-name}
type ModuleName struct {
    BaseImage string
}

// Container returns a configured container with the tool installed
func (m *ModuleName) container() *dagger.Container {
    if m.BaseImage == "" {
        m.BaseImage = "ubuntu:24.04"  // Default to proven base
    }

    return dag.Container().From(m.BaseImage).
        WithExec([]string{"apt-get", "update"}).
        WithExec([]string{"apt-get", "install", "-y", "curl", "wget", "git", "ca-certificates"}).
        WithExec([]string{"bash", "-c", "curl -fsSL https://tool-official-site.io/install.sh | bash"}).
        WithEntrypoint([]string{"tool-name"})
}

// ModuleNameVersion returns the installed tool version
func (m *ModuleName) ModuleNameVersion(ctx context.Context) (string, error) {
    return m.container().
        WithExec([]string{"tool-name", "version"}).
        Stdout(ctx)
}

// TestModuleName performs basic functionality testing
func (m *ModuleName) TestModuleName(ctx context.Context) (string, error) {
    // Use simple, stable configuration
    testConfig := `simple = "configuration"
value = "stable"`

    return m.container().
        WithNewFile("/tmp/simple-test.ext", testConfig).
        WithWorkdir("/tmp").
        WithExec([]string{"tool-name", "run", "simple-test.ext"}).
        Stdout(ctx)
}

// RunModuleName executes tool with provided source directory
func (m *ModuleName) RunModuleName(
    ctx context.Context,
    // Source directory containing tool files
    source *dagger.Directory,
    // Entry point file (optional)
    // +optional
    entrypoint string,
) (string, error) {
    if entrypoint == "" {
        entrypoint = "main.ext"
    }

    return m.container().
        WithMountedDirectory("/src", source).
        WithWorkdir("/src").
        WithExec([]string{"tool-name", "run", entrypoint}).
        Stdout(ctx)
}

// ValidateModuleName validates tool files in source directory
func (m *ModuleName) ValidateModuleName(
    ctx context.Context,
    // Source directory to validate
    source *dagger.Directory,
) (string, error) {
    return m.container().
        WithMountedDirectory("/src", source).
        WithWorkdir("/src").
        WithExec([]string{"sh", "-c", "tool-name validate . > /dev/null && echo 'Validation successful'"}).
        Stdout(ctx)
}
```

### Step 4: Add to Taskfile.yaml

#### Add Test Task

```yaml
test-{module-name}:
  desc: Test {module-name} functions
  cmds:
    - |
      echo "Testing {module-name} version..."
      dagger call -m {{ .MODULE }} \
      {module-name}-version \
      --progress plain
    - |
      echo "Testing {module-name} basic functionality..."
      dagger call -m {{ .MODULE }} \
      test-{module-name} \
      --progress plain
    - |
      echo "Testing {module-name} with project files..."
      dagger call -m {{ .MODULE }} \
      run-{module-name} \
      --source {{ .TEST_PROJECT }} \
      --entrypoint {{ .TEST_ENTRYPOINT }} \
      --progress plain
    - |
      echo "Testing {module-name} validation..."
      dagger call -m {{ .MODULE }} \
      validate-{module-name} \
      --source {{ .TEST_PROJECT }} \
      --progress plain
  vars:
    MODULE: {module-name}
    TEST_PROJECT: tests/{module-name}/test-{module-name}-project
    TEST_ENTRYPOINT: main.ext
```

#### Update Main Test Task

```yaml
test:
  desc: Select test to run
  cmds:
    - |
      totest=$(gum choose "ansible" "go" "helm" "crossplane" "kcl" "{module-name}")
      echo "Selected: ${totest}"
      task test-${totest}
```

### Step 5: Create Test Data

#### Test Project Structure

```
tests/{module-name}/
├── README.md                    # Test documentation
├── test-{module-name}-project/  # Sample project
│   ├── main.ext                # Main test file
│   ├── config.ext              # Configuration file
│   └── data/                   # Test data
└── test-advanced.ext           # Advanced test scenarios
```

#### Example Test Files

**tests/{module-name}/test-{module-name}-project/main.ext:**
```
# Simple, stable configuration that demonstrates basic functionality
name = "test-project"
version = "1.0.0"

# Add module-specific configuration
config = {
    "key": "value",
    "enabled": true
}
```

### Step 6: Documentation

#### README.md Template

```markdown
# {Module Name} Dagger Module

This module provides Dagger functions for working with {Tool Name}.

## Features

- ✅ {Tool} version checking and validation
- ✅ Basic functionality testing
- ✅ Project execution and validation
- ✅ Advanced {tool-specific} operations

## Prerequisites

- Dagger CLI installed
- Docker runtime available

## Usage

### Version Check

```bash
dagger call -m {module-name} {module-name}-version
```

### Basic Testing

```bash
dagger call -m {module-name} test-{module-name}
```

### Run Project

```bash
dagger call -m {module-name} run-{module-name} \
  --source ./my-project \
  --entrypoint main.ext
```

### Validate Project

```bash
dagger call -m {module-name} validate-{module-name} \
  --source ./my-project
```

## API Reference

### Functions

#### `{ModuleName}Version() string`
Returns the installed {tool} version.

#### `Test{ModuleName}() string`
Performs basic functionality testing with simple configuration.

#### `Run{ModuleName}(source Directory, entrypoint string) string`
Executes {tool} with the provided source directory.
- `source`: Directory containing {tool} files
- `entrypoint`: Main file to execute (optional, defaults to main.ext)

#### `Validate{ModuleName}(source Directory) string`
Validates {tool} files in the source directory.
- `source`: Directory to validate

## Testing

The module includes comprehensive tests accessible via:

```bash
# Interactive testing
task test
# Select: {module-name}

# Direct testing
task test-{module-name}
```

## Examples

### Example 1: Basic Usage

```bash
# Create a simple project
mkdir my-{module-name}-project
echo 'name = "hello"' > my-{module-name}-project/main.ext

# Run with Dagger
dagger call -m {module-name} run-{module-name} \
  --source ./my-{module-name}-project
```

### Example 2: Validation

```bash
# Validate project structure
dagger call -m {module-name} validate-{module-name} \
  --source ./my-{module-name}-project
```

## Troubleshooting

### Common Issues

1. **Installation failures**: Check base image compatibility
2. **Execution errors**: Verify file syntax and structure
3. **Validation failures**: Ensure project follows {tool} standards

### Debug Mode

Run with verbose output for troubleshooting:

```bash
dagger call -m {module-name} {function-name} --progress plain -vv
```

## Contributing

1. Follow Stuttgart-Things development standards
2. Ensure all tests pass via `task test-{module-name}`
3. Update documentation for any API changes
4. Include examples for new functionality

## Resources

- [{Tool} Official Documentation](https://tool-site.io/docs)
- [Dagger Documentation](https://docs.dagger.io/)
- [Stuttgart-Things Standards](https://github.com/stuttgart-things)
```

## 🧪 Testing Standards

### Required Test Functions

Every module must implement:

1. **Version Test**: `{Module}Version()` - Verify tool installation
2. **Basic Test**: `Test{Module}()` - Verify core functionality
3. **Execution Test**: `Run{Module}()` - Execute with real files
4. **Validation Test**: `Validate{Module}()` - Validate project structure

### Test Data Requirements

- **Real-world scenarios**: Test files should demonstrate actual usage
- **Stable configurations**: Avoid complex syntax that may cause crashes
- **Multiple examples**: Include basic, intermediate, and advanced examples
- **Error cases**: Test validation with invalid inputs

### Testing Integration

```bash
# Module testing is integrated into main workflow
task test-{module-name}    # Individual module testing
task test                  # Interactive selection
task release              # Includes all module tests
```

## 🔧 Advanced Features

### Custom Base Images

```go
// Allow custom base image override
func (m *ModuleName) WithBaseImage(image string) *ModuleName {
    m.BaseImage = image
    return m
}

// Usage:
// dagger call -m module-name with-base-image --image alpine:latest {function}
```

### Environment Configuration

```go
// Support environment variables
func (m *ModuleName) WithEnv(name, value string) *ModuleName {
    // Implementation for environment configuration
    return m
}
```

### Advanced File Operations

```go
// Export generated files
func (m *ModuleName) Generate{Something}(
    ctx context.Context,
    source *dagger.Directory,
) (*dagger.Directory, error) {
    return m.container().
        WithMountedDirectory("/src", source).
        WithWorkdir("/src").
        WithExec([]string{"tool-name", "generate", "output"}).
        Directory("/src/output"), nil
}
```

## 📋 Quality Checklist

Before submitting a new module:

### Code Quality
- [ ] All required functions implemented
- [ ] Proper error handling and validation
- [ ] Clean, readable code with comments
- [ ] Follows Go/Python best practices

### Testing
- [ ] All test functions work correctly
- [ ] Test data includes real-world examples
- [ ] Validation works with various inputs
- [ ] Module passes `task test-{module-name}`

### Documentation
- [ ] Complete README.md with examples
- [ ] Function documentation in code
- [ ] API reference is accurate
- [ ] Troubleshooting section included

### Integration
- [ ] Added to main test task selection
- [ ] Test task properly configured in Taskfile.yaml
- [ ] Works with interactive `task test`
- [ ] Follows repository structure standards

### Stability
- [ ] Uses proven base images
- [ ] Stable tool installation process
- [ ] Graceful error handling
- [ ] No segfaults or crashes in testing

## 🤝 Contributing Guidelines

1. **Follow established patterns** from existing modules
2. **Use interactive task system** for development (`task do`, `task test`)
3. **Test thoroughly** before submitting
4. **Document comprehensively** with examples
5. **Follow conventional commits** for version management

## 📚 Resources

- [Dagger Go SDK Documentation](https://docs.dagger.io/sdk/go)
- [Dagger Python SDK Documentation](https://docs.dagger.io/sdk/python)
- [Stuttgart-Things Standards](https://github.com/stuttgart-things)
- [Task Documentation](https://taskfile.dev/)
- [Container Best Practices](https://docs.docker.com/develop/dev-best-practices/)
