package main

import (
	"context"
	"dagger/hugo/internal/dagger"
)

func (m *Hugo) ExportStaticContent(
	ctx context.Context,
	siteDir *dagger.Directory,
	// The Theme to use
	// +optional
	// +default="github.com/joshed-io/reveal-hugo"
	theme string,
) (*dagger.Directory, error) {
	// Create container with mounted site directory
	ctr := m.container().
		WithMountedDirectory("/src", siteDir).
		WithWorkdir("/src")

	ctr = ctr.WithExec([]string{
		"hugo",
		"mod",
		"get", // Fixed quote escaping
		theme,
	})

	ctr = ctr.WithExec([]string{
		"hugo",
		"mod",
		"vendor", // Fixed quote escaping
	})

	// Run Hugo build command with proper arguments
	ctr = ctr.WithExec([]string{
		"hugo",
		"--minify",
		"--baseURL=/",
		"--cleanDestinationDir",
		"--theme", theme,
	})

	// Use sed to replace /%22/%22/ with /
	ctr = ctr.WithExec([]string{
		"sh", "-c",
		`sed -i 's|/%22/%22/|/|g' public/index.html`,
	})

	// RETURN THE GENERATED PUBLIC DIRECTORY WITH STATIC CONTENT
	return ctr.Directory("public"), nil
}

func (m *Hugo) BuildAndExport(
	ctx context.Context,
	name string,
	config *dagger.File,
	content *dagger.Directory,
	// The Theme to use
	// +optional
	// +default="github.com/joshed-io/reveal-hugo"
	theme string,
) (*dagger.Directory, error) {
	// Initialize the site
	siteDir, err := m.InitSite(ctx, name, config, content, theme)
	if err != nil {
		return nil, err
	}

	// Export static content
	return m.ExportStaticContent(ctx, siteDir, theme)
}
