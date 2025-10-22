# Stuttgart-Things Dagger Development Environment

This directory contains the development decisions, standards, and documentation for the Stuttgart-Things Dagger repository.

## 📋 Contents

- **[decisions.md](./decisions.md)** - Comprehensive architectural and development decisions
- **[release-guide.md](./release-guide.md)** - Step-by-step release process guide
- **[module-development.md](./module-development.md)** - Guidelines for creating new Dagger modules

## 🚀 Quick Start

### Development Workflow

1. **Create Feature Branch**
   ```bash
   task branch
   # Interactive branch name input
   ```

2. **Develop and Test**
   ```bash
   task test  # Choose module to test
   task do    # Choose from all available tasks
   ```

3. **Commit Changes**
   ```bash
   task commit
   # Interactive commit message selection with gum
   ```

4. **Create Release**
   ```bash
   task release
   # Automated testing, PR creation, and semantic release
   ```

### Interactive Task System

All complex operations use **gum** for user-friendly selection:

- `task test` - Module testing with visual selection
- `task do` - All tasks selection from Taskfile.yaml
- `task commit` - Conventional commit message selection
- `task create` - New module creation wizard
- `task switch-remote/local` - Branch switching

## 🏗️ Module Development

### Current Modules

- **ansible** - Ansible playbook and collection automation
- **go** - Go application building, testing, and security scanning
- **helm** - Helm chart operations (lint, package, push, validate)
- **crossplane** - Crossplane package management
- **kcl** - KCL configuration language tools and CRD conversion
- **terraform** - Terraform infrastructure automation
- **docker** - Container image building and publishing
- **hugo** - Static site generation

### Creating New Modules

1. Use `task create` for interactive module creation
2. Follow the [module development guidelines](./module-development.md)
3. Implement required test functions
4. Add to main test selection in Taskfile.yaml
5. Create comprehensive documentation

## 📦 Release Management

### Semantic Versioning

- `feat:` commits → Minor version bump (0.x.0)
- `fix:` commits → Patch version bump (0.0.x)
- `BREAKING CHANGE:` commits → Major version bump (x.0.0)

### Automated Release Process

The `task release` command provides fully automated releases:

1. **Testing Phase**: All modules tested comprehensively
2. **PR Phase**: Automatic pull request creation and merging
3. **Release Phase**: Semantic-release with GitHub integration
4. **Cleanup Phase**: Branch deletion and main branch sync

## 🔧 Development Standards

### Repository Structure
```
dagger/
├── .container-use/          # Development decisions and docs
├── kcl/                     # KCL module (main modules in root)
├── ansible/                 # Ansible module
├── go/                      # Go module
├── helm/                    # Helm module
├── tests/                   # Test files and data
│   ├── kcl/                # KCL test files
│   ├── ansible/            # Ansible test files
│   └── ...                 # Other module test files
├── Taskfile.yaml           # Task automation
└── README.md              # Main documentation
```

### Quality Gates

- ✅ All tests pass via `task test-{module}`
- ✅ Documentation complete and accurate
- ✅ Module accessible via `task test` selection
- ✅ Stable and reproducible installation
- ✅ Examples work as documented

## 🤝 Contributing

1. Follow the established development decisions
2. Use the interactive task system for consistency
3. Ensure comprehensive testing before releases
4. Maintain clear documentation and examples
5. Follow conventional commit format for automated versioning

## 📚 Additional Resources

- [Dagger Documentation](https://docs.dagger.io/)
- [Stuttgart-Things Standards](https://github.com/stuttgart-things)
- [Task Documentation](https://taskfile.dev/)
- [Gum TUI Toolkit](https://github.com/charmbracelet/gum)
- [Semantic Release](https://semantic-release.gitbook.io/)