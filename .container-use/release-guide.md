# Release Guide - Stuttgart-Things Dagger

This guide provides step-by-step instructions for creating releases in the Stuttgart-Things Dagger repository.

## 🚀 Release Process Overview

The release process is fully automated using Task and follows semantic versioning principles.

### Prerequisites

- ✅ Access to the repository with push permissions
- ✅ GitHub token configured for semantic-release
- ✅ All dependencies installed (task, gum, semantic-release)
- ✅ Clean working directory on `main` branch

## 📋 Step-by-Step Release Process

### Step 1: Create Feature Branch

```bash
task branch
```

**What happens:**
1. Switches to `main` branch automatically
2. Pulls latest changes from origin
3. Prompts for new branch name (interactive)
4. Creates new branch and pushes to origin
5. Sets upstream tracking

**Example:**
```bash
$ task branch
Switched to branch 'main'
Your branch is up to date with 'origin/main'.
Already up to date.
Enter to be created (remote) branch:
feat/add-new-module
```

### Step 2: Development Phase

Develop your features/fixes in the branch:

```bash
# Test specific modules during development
task test
# Select from: ansible, go, helm, crossplane, kcl

# Use interactive task selection for other operations
task do
# Select from all available tasks
```

**Testing Options:**
- **Individual Module Testing**: `task test-{module-name}`
- **Interactive Selection**: `task test` (uses gum for selection)
- **All Tasks**: `task do` (shows all available tasks)

### Step 3: Commit Changes

```bash
task commit
```

**Interactive Commit Process:**
1. Cleans temporary directories
2. Sets upstream tracking
3. Pulls latest changes
4. Stages all changes
5. **Gum-powered commit message selection:**
   - `feat: {branch-name}` - New features (minor version bump)
   - `fix: {branch-name}` - Bug fixes (patch version bump)
   - `BREAKING CHANGE: {branch-name}` - Breaking changes (major version bump)
   - `ENTER CUSTOM COMMIT MESSAGE` - Custom conventional commit

**Example:**
```bash
$ task commit
committing changes
ENTER COMMIT MESSAGE
> feat: feat/add-new-module
  fix: feat/add-new-module
  BREAKING CHANGE: feat/add-new-module
  ENTER CUSTOM COMMIT MESSAGE
```

### Step 4: Create Release

```bash
task release
```

**Automated Release Pipeline:**

#### Phase 1: Comprehensive Testing
```bash
# All modules tested automatically:
task test-go          # Go module functionality
task test-helm        # Helm chart operations
task test-ansible     # Ansible automation
task test-docker      # Container operations
task test-hugo        # Static site generation
task test-terraform   # Infrastructure automation
```

#### Phase 2: Pull Request Automation
```bash
task pr               # Automated PR workflow:
                     # - Creates PR with branch name
                     # - Waits for checks (20s)
                     # - Auto-merges with rebase
                     # - Deletes remote branch
                     # - Switches to main and pulls
```

#### Phase 3: Semantic Release
```bash
semantic-release --dry-run      # Validates release
semantic-release --debug --no-ci # Creates actual release
```

**Release Outputs:**
- ✅ New version tag (e.g., `v0.35.0`)
- ✅ GitHub release with changelog
- ✅ Updated CHANGELOG.md
- ✅ PR comments with release information

## 🔄 Release Types and Versioning

### Conventional Commits → Version Bumps

| Commit Type | Example | Version Impact |
|-------------|---------|----------------|
| `feat:` | `feat: add kcl module` | **Minor** (0.x.0) |
| `fix:` | `fix: resolve helm lint issues` | **Patch** (0.0.x) |
| `BREAKING CHANGE:` | `BREAKING CHANGE: remove deprecated API` | **Major** (x.0.0) |

### Version Examples

**Current Version:** `v0.34.0`

- **feat: add new module** → `v0.35.0`
- **fix: update dependencies** → `v0.34.1`
- **BREAKING CHANGE: remove old API** → `v1.0.0`

## 🧪 Testing Strategy

### Module-Specific Testing

Each module includes comprehensive testing:

```bash
# KCL Module Testing
task test-kcl
# Tests: version, basic functionality, project files, validation, CRD conversion

# Ansible Module Testing
task test-ansible
# Tests: playbook execution, collection building

# Go Module Testing
task test-go
# Tests: linting, binary building, container building, security scanning
```

### Interactive Testing

```bash
task test
```
**Gum Selection Interface:**
```
? Select test to run:
  > ansible
    go
    helm
    crossplane
    kcl
```

## 🛠️ Troubleshooting

### Common Issues

#### 1. Test Failures
```bash
# Run specific module test to debug
task test-{module-name}

# Check individual functions
dagger call -m {module} {function-name} --progress plain
```

#### 2. Branch Issues
```bash
# Switch to existing remote branch
task switch-remote

# Switch to local branch
task switch-local
```

#### 3. Commit Message Issues
```bash
# Use custom commit message
task commit
# Select "ENTER CUSTOM COMMIT MESSAGE"
# Format: type(scope): description
# Example: feat(kcl): add CRD conversion functionality
```

#### 4. Release Failures
```bash
# Check semantic-release dry run
semantic-release --dry-run

# Verify GitHub token
echo $GITHUB_TOKEN

# Check repository status
git status
git log --oneline -5
```

### Recovery Procedures

#### Failed Release
1. Check the error in semantic-release output
2. Fix issues (usually commit format or token)
3. Re-run `task release`

#### Branch Conflicts
1. Use `task switch-local` to change branches
2. Resolve conflicts manually
3. Continue with `task commit`

#### Test Failures
1. Debug with individual module tests
2. Fix issues in the module code
3. Re-run `task test` to verify

## 📊 Release Metrics

### Automatic Changelog Generation

Each release includes:
- **Features Added**: List of new functionality
- **Bugs Fixed**: Resolved issues
- **Breaking Changes**: API changes requiring user action
- **Commit Links**: Direct links to GitHub commits

### GitHub Integration

- **Release Notes**: Automatic generation from conventional commits
- **Asset Publishing**: Automated for configured file patterns
- **PR Comments**: Automatic notification on included PRs
- **Issue Linking**: Automatic closure of related issues

## 🎯 Best Practices

### Development Workflow
1. **Always start with `task branch`** for clean feature branches
2. **Test frequently** using `task test` during development
3. **Use conventional commits** for automated versioning
4. **Keep commits focused** on single features/fixes

### Release Workflow
1. **Let automation handle the process** - don't skip steps
2. **Wait for all tests** to pass before releasing
3. **Review generated changelog** after release
4. **Monitor GitHub release** for any issues

### Quality Assurance
1. **All modules must pass testing** before release
2. **Documentation must be updated** with new features
3. **Examples must work** as documented
4. **Breaking changes must be clearly communicated**

## 📚 Additional Resources

- [Conventional Commits Specification](https://conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [Semantic Release Documentation](https://semantic-release.gitbook.io/)
- [Task Documentation](https://taskfile.dev/)
- [Gum TUI Components](https://github.com/charmbracelet/gum)