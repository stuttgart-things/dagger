# Stuttgart-Things Dagger Development Decisions

## Universal Organizational Decisions

### Decision 1: Container-Use Development Environment
- **Status**: ✅ Accepted
- **Date**: 2025-01-22
- **Applies to**: All repositories
- **Context**: Standardized development environments across all Stuttgart-Things projects
- **Consequences**: Consistent tooling, reproducible builds, easier onboarding

### Decision 2: Stuttgart-Things Domain and Naming
- **Status**: ✅ Accepted
- **Date**: 2025-01-22
- **Applies to**: All repositories
- **Context**: Consistent organizational branding and API grouping
- **Consequences**: All APIs use `*.stuttgart-things.com` groups, consistent repository naming

### Decision 3: Git Workflow and Conventional Commits
- **Status**: ✅ Accepted
- **Date**: 2025-01-22
- **Applies to**: All repositories
- **Context**: Version management and automation consistency
- **Consequences**: Automated semantic versioning, standardized commit messages

### Decision 4: Testing Strategy and Quality Gates
- **Status**: ✅ Accepted
- **Date**: 2025-01-22
- **Applies to**: All repositories
- **Context**: Quality assurance standards across all technology stacks
- **Consequences**: Mandatory testing before merge, consistent CI/CD pipelines

### Decision 5: Documentation Requirements
- **Status**: ✅ Accepted
- **Date**: 2025-01-22
- **Applies to**: All repositories
- **Context**: Knowledge sharing and onboarding standardization
- **Consequences**: Comprehensive README.md files, testing documentation, API examples

### Decision 6: OCI Registry for Modules and Packages
- **Status**: ✅ Accepted
- **Date**: 2025-01-22
- **Applies to**: All repositories
- **Context**: Centralized distribution of reusable components
- **Consequences**: All modules published to `ghcr.io/stuttgart-things/*`, version management

---

## Dagger Module Development Decisions

### Decision DG-1: Dagger Module Repository Structure
- **Status**: ✅ Accepted
- **Date**: 2025-01-22
- **Context**: Standardized structure for all Dagger modules within the repository
- **Decision**: All new Dagger modules must be placed in the `tests/` directory at repository root
- **Rationale**:
  - Clear separation between module code and tests
  - Consistent across all Stuttgart-Things Dagger repositories
  - Easy discovery and navigation
  - Follows established patterns from other repositories
- **Consequences**:
  - New modules: `tests/{module-name}/`
  - Existing modules to be migrated to `tests/` structure
  - Taskfile tasks must reference `tests/{module-name}` paths

### Decision DG-2: Mandatory Taskfile Test Integration
- **Status**: ✅ Accepted
- **Date**: 2025-01-22
- **Context**: Consistent testing and automation across all Dagger modules
- **Decision**: Every Dagger module MUST have a corresponding `test-{module-name}` task in the root Taskfile.yaml
- **Requirements**:
  ```yaml
  test-{module-name}:
    desc: Test {module-name} functions
    cmds:
      - |
        echo "Testing {module-name} version..."
        dagger call -m {{ .MODULE }} \
        {module}-version \
        --progress plain
      - |
        echo "Testing {module-name} basic functionality..."
        dagger call -m {{ .MODULE }} \
        test-{module} \
        --progress plain
      # Additional module-specific tests
    vars:
      MODULE: tests/{module-name}
      # Module-specific variables
  ```
- **Integration**: Module must be added to the main `test` task selection:
  ```yaml
  test:
    desc: Select test to run
    cmds:
      - |
        totest=$(gum choose "ansible" "go" "helm" "crossplane" "kcl" "{module-name}")
        echo "Selected: ${totest}"
        task test-${totest}
  ```
- **Consequences**:
  - All modules discoverable via `task test`
  - Consistent testing interface across modules
  - Automated CI/CD integration possible
  - Quality gates enforced before merging

### Decision DG-3: Dagger Module Testing Standards
- **Status**: ✅ Accepted
- **Date**: 2025-01-22
- **Context**: Quality assurance for all Dagger modules
- **Decision**: Every Dagger module must implement standardized testing patterns
- **Required Test Functions**:
  - `{Module}Version()` - Version information test
  - `Test{Module}()` - Basic functionality test
  - Module-specific functionality tests (e.g., `RunKcl()`, `ValidateKcl()`)
- **Test Structure Pattern**:
  ```go
  // Version test - verify tool installation
  func (m *{Module}) {Module}Version(ctx context.Context) (string, error)

  // Basic functionality test - verify core features
  func (m *{Module}) Test{Module}(ctx context.Context) (string, error)

  // Advanced functionality tests - module-specific features
  func (m *{Module}) Run{Module}(ctx context.Context, source *dagger.Directory, ...) (string, error)
  func (m *{Module}) Validate{Module}(ctx context.Context, source *dagger.Directory) (string, error)
  ```
- **Test Data Requirements**:
  - Each module must include `test-{module}-project/` directory with sample files
  - Test data must demonstrate real-world usage scenarios
  - Examples must be runnable and produce expected output
- **Consequences**:
  - Reliable module functionality verification
  - Clear API contract documentation
  - Regression testing capabilities
  - User examples for onboarding

### Decision DG-4: Container Base Image Standards
- **Status**: ✅ Accepted
- **Date**: 2025-01-22
- **Context**: Consistent, secure, and reliable container environments
- **Decision**: Use proven base images that match container-use environment standards
- **Approved Base Images**:
  - **Primary**: `ubuntu:24.04` (for complex installations requiring apt packages)
  - **Secondary**: `cgr.dev/chainguard/wolfi-base:latest` (for security-focused deployments)
  - **Alpine**: `alpine:latest` (for minimal footprint requirements)
- **Installation Patterns**:
  ```go
  func (m *{Module}) container() *dagger.Container {
      if m.BaseImage == "" {
          m.BaseImage = "ubuntu:24.04"  // Default to proven base
      }

      ctr := dag.Container().From(m.BaseImage)

      // Use official installation scripts when available
      ctr = ctr.WithExec([]string{"apt-get", "update"})
      ctr = ctr.WithExec([]string{"apt-get", "install", "-y", "curl", "wget", "git", "ca-certificates"})
      ctr = ctr.WithExec([]string{"bash", "-c", "curl -fsSL https://official-tool.io/script/install.sh | bash"})

      return ctr
  }
  ```
- **Configuration Principles**:
  - Match container-use environment setup for consistency
  - Use official installation scripts when available
  - Prefer stable, well-tested installation methods
  - Avoid complex manual installations
- **Consequences**:
  - Consistent behavior across development and Dagger environments
  - Reduced installation complexity and failures
  - Easier troubleshooting and debugging
  - Proven stability in container-use environments

### Decision DG-5: Module Documentation Requirements
- **Status**: ✅ Accepted
- **Date**: 2025-01-22
- **Context**: Comprehensive documentation for all Dagger modules
- **Decision**: Every Dagger module must include comprehensive documentation
- **Required Documentation**:
  - **README.md** in module directory with:
    - Purpose and functionality overview
    - Installation and usage instructions
    - API reference with all functions documented
    - Examples for common use cases
    - Testing instructions
  - **Go code documentation** with:
    - Function-level comments explaining purpose and parameters
    - Module-level comments describing overall functionality
    - Example usage in comments where helpful
  - **Test examples** with:
    - Sample input files and expected outputs
    - Real-world usage scenarios
    - Performance and limitation notes
- **Documentation Standards**:
  ```go
  // {Module} provides functionality for working with {tool}
  //
  // The module includes functions for:
  // - Version checking and tool validation
  // - Basic functionality testing
  // - Advanced operations like validation and execution
  //
  // Example usage:
  //   dagger call -m tests/{module} {module}-version
  //   dagger call -m tests/{module} test-{module}
  type {Module} struct {
      BaseImage string
  }

  // {Module}Version returns the installed {tool} version
  func (m *{Module}) {Module}Version(ctx context.Context) (string, error) {
      // Implementation with clear purpose
  }
  ```
- **Consequences**:
  - Self-documenting modules for easier adoption
  - Clear API contracts and expectations
  - Reduced support overhead
  - Better onboarding experience

### Decision DG-6: Error Handling and Stability Patterns
- **Status**: ✅ Accepted
- **Date**: 2025-01-22
- **Context**: Reliable and predictable Dagger module behavior
- **Decision**: Implement robust error handling and stability patterns in all modules
- **Error Handling Requirements**:
  - Graceful handling of tool installation failures
  - Clear error messages for debugging
  - Fallback strategies for common issues
  - Input validation and sanitization
- **Stability Patterns**:
  ```go
  // Avoid complex operations that cause segfaults or crashes
  func (m *{Module}) Test{Module}(ctx context.Context) (string, error) {
      // Use simple, stable configurations
      testConfig := `simple = "configuration"
  value = "stable"`

      return m.container().
          WithNewFile("/tmp/simple-test.ext", testConfig).
          WithWorkdir("/tmp").
          WithExec([]string{"{tool}", "run", "simple-test.ext"}).
          Stdout(ctx)
  }

  // Implement validation through compilation/execution
  func (m *{Module}) Validate{Module}(ctx context.Context, source *dagger.Directory) (string, error) {
      return m.container().
          WithMountedDirectory("/src", source).
          WithWorkdir("/src").
          WithExec([]string{"sh", "-c", "{tool} run main.ext > /dev/null && echo 'Validation successful'"}).
          Stdout(ctx)
  }
  ```
- **Testing Principles**:
  - Start with simple configurations to verify basic functionality
  - Avoid complex syntax that may cause crashes
  - Provide meaningful validation feedback
  - Use stable, proven patterns from container-use environments
- **Consequences**:
  - More reliable module execution
  - Better debugging experience
  - Reduced support issues
  - Consistent behavior across modules

---

## Implementation Guidelines

### Module Creation Checklist
- [ ] Create module in `tests/{module-name}/` directory
- [ ] Implement required test functions (`*Version`, `Test*`, etc.)
- [ ] Add `test-{module-name}` task to root Taskfile.yaml
- [ ] Include module in main `test` task selection
- [ ] Create comprehensive README.md documentation
- [ ] Add test data in `test-{module}-project/` directory
- [ ] Use approved base images and installation patterns
- [ ] Implement robust error handling and validation
- [ ] Test all functionality thoroughly before merging

### Quality Gates
- All tests must pass via `task test-{module-name}`
- Documentation must be complete and accurate
- Module must be accessible via `task test` selection
- Installation must be stable and reproducible
- Examples must work as documented

### Maintenance Standards
- Keep modules updated with latest tool versions
- Monitor for security updates in base images
- Update documentation when functionality changes
- Maintain backward compatibility when possible
- Follow semantic versioning for breaking changes

---

## Release Management Process

### Decision RM-1: Automated Release Workflow
- **Status**: ✅ Accepted
- **Date**: 2025-10-22
- **Context**: Standardized release process using Task automation and semantic versioning
- **Decision**: All releases must follow the automated branch → test → release → semantic versioning workflow

### Release Process Steps

#### 1. Branch Creation (`task branch`)
```bash
task branch
```
- **Purpose**: Create a feature branch from main for development
- **Process**:
  1. Switches to `main` branch
  2. Pulls latest changes
  3. Prompts for new branch name (interactive input)
  4. Creates and pushes new branch to origin
  5. Sets upstream tracking to origin/main

#### 2. Development and Testing
- Develop features/fixes in the branch
- Use `task test` to run module-specific tests:
  ```bash
  task test  # Interactive selection via gum
  # Choose from: ansible, go, helm, crossplane, kcl
  ```
- Use `task do` for other development tasks:
  ```bash
  task do  # Interactive task selection via gum
  ```

#### 3. Commit and Push (`task commit`)
```bash
task commit
```
- **Purpose**: Commit changes with conventional commit messages
- **Process**:
  1. Cleans dist/ directory
  2. Sets upstream tracking to current branch
  3. Pulls latest changes
  4. Stages all changes with `git add *`
  5. Interactive commit message selection via **gum**:
     - `feat: {branch-name}`
     - `fix: {branch-name}`
     - `BREAKING CHANGE: {branch-name}`
     - `ENTER CUSTOM COMMIT MESSAGE` (custom input)
  6. Commits and pushes to origin

#### 4. Release Creation (`task release`)
```bash
task release
```
- **Purpose**: Automated testing, PR creation, and semantic release
- **Process**:
  1. **Comprehensive Testing**: Runs all module tests
     - `task test-go`
     - `task test-helm`
     - `task test-ansible`
     - `task test-docker`
     - `task test-hugo`
     - `task test-terraform`
  2. **Pull Request**: Executes `task pr`
     - Creates PR with branch name as title/description
     - Waits 20 seconds for checks
     - Auto-merges with rebase and deletes branch
     - Switches back to main and pulls
  3. **Semantic Release**:
     - Dry-run to validate release
     - Actual release with semantic-release
     - Creates GitHub release with changelog
     - Updates version tags

### Interactive Task System (Gum Integration)

#### Decision RM-2: Gum-Based Interactive Task Selection
- **Status**: ✅ Accepted
- **Date**: 2025-10-22
- **Context**: User-friendly task execution with visual selection
- **Decision**: All complex task selections use `gum choose` for better UX

#### Available Interactive Tasks

1. **`task test`** - Module Testing Selection
   ```bash
   task test
   # Gum selection: ansible, go, helm, crossplane, kcl
   ```

2. **`task do`** - General Task Selection
   ```bash
   task do
   # Gum selection from all available tasks in Taskfile.yaml
   ```

3. **`task create`** - Module Creation
   ```bash
   task create
   # Interactive inputs:
   # - Module name (gum input)
   # - SDK selection (gum choose: go, python)
   ```

4. **`task commit`** - Commit Message Selection
   ```bash
   task commit
   # Gum selection:
   # - feat: {branch}
   # - fix: {branch}
   # - BREAKING CHANGE: {branch}
   # - ENTER CUSTOM COMMIT MESSAGE
   ```

5. **`task switch-remote`** - Remote Branch Selection
   ```bash
   task switch-remote
   # Gum selection from remote branches
   ```

6. **`task switch-local`** - Local Branch Selection
   ```bash
   task switch-local
   # Gum selection from local branches
   ```

### Release Automation Features

#### Conventional Commits and Semantic Versioning
- **Commit Types**: `feat:`, `fix:`, `BREAKING CHANGE:`
- **Version Bumps**:
  - `feat:` → Minor version (0.x.0)
  - `fix:` → Patch version (0.0.x)
  - `BREAKING CHANGE:` → Major version (x.0.0)

#### GitHub Integration
- Automated PR creation and merging
- GitHub release creation with changelog
- Issue/PR commenting with release information
- Tag management with `v{version}` format

#### Quality Assurance
- All tests must pass before release
- Dry-run validation before actual release
- Branch protection and automated cleanup
- Comprehensive test coverage across all modules

### Consequences
- **Streamlined Workflow**: Developers can focus on coding rather than release mechanics
- **Consistent Releases**: Automated process ensures no steps are missed
- **Quality Control**: All tests run before any release
- **User-Friendly**: Gum interface makes complex tasks approachable
- **Semantic Versioning**: Automated version management based on conventional commits
- **Documentation**: Automatic changelog generation and GitHub release notes