package main

import (
	"context"
	"dagger/slidev/internal/dagger"
)

// Export runs `slidev export` to produce a printable artifact from the deck:
// a single PDF (default), a PPTX file, or a directory of PNGs (one per slide).
//
// Export needs Playwright + headless Chromium, which is not officially
// supported on Alpine. To keep the rest of the module on the lightweight
// node:22-alpine base, Export builds its own Debian-based container
// (node:22-bookworm-slim) with the Chromium system deps and lets Playwright
// install its bundled browser via `pnpm exec playwright install chromium`.
//
// Returns the directory holding the export output: a single `slides-export.<format>`
// file for pdf/pptx, or one `<n>.png` per slide for png.
func (m *Slidev) Export(
	ctx context.Context,
	slides *dagger.File,
	// +optional
	style *dagger.File,
	// +optional
	// +default="@slidev/theme-default"
	theme string,
	// +optional
	addons []string,
	// +optional
	extras *dagger.Directory,
	// +optional
	// +default="pdf"
	format string,
	// +optional
	withClicks bool,
	// +optional
	dark bool,
) (*dagger.Directory, error) {
	output := "/deck/out/slides-export." + format
	if format == "png" {
		output = "/deck/out"
	}

	args := []string{
		"pnpm", "exec", "slidev", "export", "slides.md",
		"--format", format,
		"--output", output,
	}
	if withClicks {
		args = append(args, "--with-clicks")
	}
	if dark {
		args = append(args, "--dark")
	}

	base := dag.Container().
		From("node:22-bookworm-slim").
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "--no-install-recommends",
			"ca-certificates", "git",
			"libnss3", "libnspr4", "libatk1.0-0", "libatk-bridge2.0-0",
			"libcups2", "libdrm2", "libdbus-1-3", "libxcomposite1",
			"libxdamage1", "libxext6", "libxfixes3", "libxrandr2",
			"libgbm1", "libxkbcommon0", "libpango-1.0-0", "libcairo2",
			"libasound2", "libatspi2.0-0", "libwayland-client0",
			"fonts-liberation", "fonts-noto-color-emoji",
		}).
		WithExec([]string{"npm", "install", "-g", "pnpm"}).
		WithEnvVariable("npm_config_node_linker", "hoisted")

	return m.installDeck(base, slides, style, theme, addons, extras, []string{"playwright-chromium"}).
		WithExec([]string{"pnpm", "exec", "playwright", "install", "chromium"}).
		WithExec([]string{"mkdir", "-p", "/deck/out"}).
		WithExec(args).
		Directory("/deck/out"), nil
}
