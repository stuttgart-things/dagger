# Slidev Dagger Module

This module provides Dagger functions for [Slidev](https://sli.dev) presentations: scaffolding a deck, running a live dev server, and producing a static build — all from a single `slides.md` (plus optional `style.css`), without keeping `package.json`, `node_modules/`, lockfiles or `dist/` in git.

## Features

- ✅ Slidev project scaffolding driven by a single Markdown file
- ✅ Local development server with hot reload (`slidev` + `--remote`)
- ✅ Static build to `dist/` for any HTTP host (nginx, MinIO, S3, Pages, ...)
- ✅ Custom theme and addon installation via npm package names
- ✅ Optional `style.css` overlay for branding tweaks

## Prerequisites

- Dagger CLI installed
- Docker runtime available
- A `slides.md` (Slidev Markdown source)

## Quick Start

### Local Development

```bash
# Serve the deck on http://0.0.0.0:3030 with hot reload
dagger call -m slidev serve \
  --slides ./slides.md \
  --port 3030 \
  up --progress plain
```

### Static Build

```bash
# Build the deck into a static dist/ directory
dagger call -m slidev build \
  --slides ./slides.md \
  export --path /tmp/slidev/dist
```

### Initialize Deck Only

```bash
# Scaffold the deck (package.json + node_modules + slides.md) without running it
dagger call -m slidev init-deck \
  --slides ./slides.md \
  export --path /tmp/slidev/deck
```

## API Reference

### Serve

```bash
dagger call -m slidev serve \
  --slides ./slides.md \
  --style ./style.css \
  --theme @slidev/theme-seriph \
  --addons '["@slidev/addon-anything"]' \
  --port 3030 \
  up --progress plain
```

| Flag       | Type      | Default                  | Notes                                |
|------------|-----------|--------------------------|--------------------------------------|
| `--slides` | `File`    | required                 | Slidev Markdown source.              |
| `--style`  | `File`    | optional                 | Drops `style.css` next to `slides.md`. |
| `--theme`  | `string`  | `@slidev/theme-default`  | Any npm theme package name.          |
| `--addons` | `[]string`| optional                 | Extra npm packages to install.       |
| `--extras` | `Directory`| optional                | Overlay onto `/deck/` for partials, components, public assets, etc. `--slides`/`--style` always win. |
| `--port`   | `string`  | `3030`                   | Bound to `0.0.0.0` via `--remote`.   |

### Build

```bash
dagger call -m slidev build \
  --slides ./slides.md \
  --base / \
  export --path /tmp/slidev/dist
```

| Flag      | Type     | Default | Notes                                                              |
|-----------|----------|---------|--------------------------------------------------------------------|
| `--base`  | `string` | `/`     | Public URL prefix passed to `slidev build --base`.                 |

### InitDeck

```bash
dagger call -m slidev init-deck \
  --slides ./slides.md \
  --theme @slidev/theme-default \
  export --path /tmp/slidev/deck
```

Returns the populated `/deck` directory (with `package.json`, `node_modules/`, `slides.md`, and `style.css` when provided). Useful for inspection or chaining into custom workflows.

### Multi-file Decks (`--extras`)

For decks that split chapters into separate files via Slidev's `src:` includes — a typical workshop layout looks like:

```
my-workshop/
├── slides.md          # entry deck with `src: ./slides/NN_*.md` includes
├── slides/            # chapter partials
│   ├── 00_agenda.md
│   ├── 01_intro.md
│   └── ...
└── style.css          # global style overrides
```

#### Live dev server

```bash
dagger call -m /path/to/dagger/slidev serve \
  --slides /path/to/my-workshop/slides.md \
  --style  /path/to/my-workshop/style.css \
  --extras /path/to/my-workshop \
  --port 3030 \
  up --progress plain
```

#### Static build

```bash
dagger call -m /path/to/dagger/slidev build \
  --slides /path/to/my-workshop/slides.md \
  --style  /path/to/my-workshop/style.css \
  --extras /path/to/my-workshop \
  export --path /tmp/slidev/dist
```

`--extras` is overlaid on `/deck/` after dependency install, so anything below the directory you point at lands in the deck root with the same layout. The explicit `--slides`/`--style` files are written last and always win, so it's safe to point `--extras` at a presentation root that also contains those files. This means only the Markdown sources and `style.css` need to live in git — `package.json`, lockfiles, `node_modules/` and `dist/` are all regenerated in the container.

## Serving the Static Build

After `build`, any static file server works:

```bash
chmod -R o+rX /tmp/slidev/dist

docker run --rm -p 8080:80 \
  -v "/tmp/slidev/dist:/usr/share/nginx/html" nginx
```

## Why This Module

Slidev decks normally drag a full Vite/Vue project into git. This module flips it around: only the content you authored (`slides.md`, optional `style.css`) lives in git — everything else (`package.json`, lockfile, `node_modules/`, `dist/`) is regenerated on demand inside the container.

## Resources

- [Slidev Documentation](https://sli.dev)
- [Slidev Themes](https://sli.dev/resources/theme-gallery)
- [Slidev Addons](https://sli.dev/resources/addons)
