# CLAUDE.md

Repo-specific guidance for Claude Code. Keep terse.

## Repo shape

Monorepo of independent Dagger modules (one per top-level directory: `kcl/`, `ansible/`, `terraform/`, …). Each module is a self-contained Go package with its own `dagger.json` and generated `internal/` client. Touch only the module(s) relevant to the task; don't modify generated files under `*/internal/`.

## Release flow

Releases are driven by [semantic-release](https://github.com/semantic-release/semantic-release) on conventional commits against `main`.

Commit prefix → bump:
- `fix:` / `fix(scope):` → patch
- `feat:` / `feat(scope):` → minor
- `BREAKING CHANGE:` footer or `feat!:` → major
- `chore:`, `docs:`, `refactor:`, `test:` → no release

### Standard flow (CI cuts the release)

1. Branch off `main` with `fix/<slug>` or `feat/<slug>`.
2. Commit with a conventional message. Scope by module: `fix(kcl): ...`.
3. `gh pr create --base main`, merge when green.
4. CI runs semantic-release and publishes the tag + GitHub release + `chore(release)` bump commit. No manual step.

### Local release (when CI is down or you need to force a cut)

```bash
git checkout main && git pull
npx semantic-release --debug --no-ci
```

Requires `GITHUB_TOKEN` in env. `--no-ci` bypasses the CI-env guard; semantic-release still refuses to run on a dirty tree or non-release branch.

## Dagger module gotchas

- **Flag names are per-module, not per-function.** Two exported functions in the same module cannot both declare a parameter named `workdir`, even with different Go types (`string` vs `*dagger.Directory`). The CLI fails at flag registration with `flag already exists: <name>` *before* dispatch, so every call into the module breaks — not just the colliding one. When adding a parameter, grep the whole module for the name first.
- **Renaming an existing exported parameter is a breaking change** for anyone calling the module by version. If only one of the colliding names is new/unreleased, rename that one.
- **Generated code under `*/internal/` is not hand-edited.** Regenerate via `dagger develop` in the module directory if signatures change.

## Testing a module change end-to-end

```bash
dagger call -m ./kcl <function> --<flag> <value> ...
```

Use `-m ./<module>` (local path) to test uncommitted changes. Use `-m github.com/stuttgart-things/dagger/<module>@<ref>` to test a published tag.
