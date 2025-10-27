package main

import (
	"context"
	"dagger/hugo/internal/dagger"
	"fmt"
	"strconv"
	"strings"
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

func (m *Hugo) BuildSyncExport(
	ctx context.Context,
	name string,
	config *dagger.File,
	content *dagger.Directory,
	// The Theme to use
	// +optional
	// +default="github.com/joshed-io/reveal-hugo"
	theme string,
	endpoint string,
	accessKey *dagger.Secret,
	secretKey *dagger.Secret,
	bucketName string,
	aliasName string,
	insecure bool,
) (*dagger.Directory, error) {
	// Initialize the site
	siteDir, err := m.InitSite(ctx, name, config, content, theme)
	if err != nil {
		return nil, err
	}

	staticContentDir, err := m.ExportStaticContent(ctx, siteDir, theme)

	syncedBucket, err := m.SyncMinioBucket(ctx, endpoint, accessKey, secretKey, bucketName, aliasName, insecure)

	if err != nil {
		return nil, fmt.Errorf("FAILED TO SYNC MINIO BUCKET: %w", err)
	}

	finalDir := staticContentDir.WithDirectory("images", syncedBucket)

	// Escape slashes in endpoint for use in sed
	safeEndpoint := strings.ReplaceAll(endpoint, "/", `\/`)
	safeBucket := strings.ReplaceAll(bucketName, "/", `\/`)

	// Build the full match string
	matchPrefix := fmt.Sprintf("%s/%s/", safeEndpoint, safeBucket)

	replaceCmd := fmt.Sprintf(
		`find . -name '*.html' -exec sed -i 's|%s|images/|g' {} +`,
		matchPrefix,
	)

	ctr := m.container().
		WithMountedDirectory("/src", finalDir).
		WithWorkdir("/src").
		WithExec([]string{"sh", "-c", replaceCmd})
	// ctr := m.container().
	// 	WithMountedDirectory("/src", finalDir).
	// 	WithWorkdir("/src").
	// 	WithExec([]string{
	// 		"sh", "-c",
	// 		`find . -name '*.html' -exec sed -i 's|https://artifacts.automation.sthings-vsphere.labul.sva.de/idp/|images/|g' {} +`,
	// 	})

	// Get the modified directory from the container
	updatedDir := ctr.Directory("/src")

	// Export static content
	return updatedDir, nil
}

func (m *Hugo) SyncMinioBucket(
	ctx context.Context,
	endpoint string,
	accessKey *dagger.Secret,
	secretKey *dagger.Secret,
	bucketName string,
	aliasName string,
	insecure bool,
) (*dagger.Directory, error) {

	endpoint = strings.TrimPrefix(endpoint, "https://")
	notSecure := strconv.FormatBool(insecure)

	accessKeyStr, err := accessKey.Plaintext(ctx)
	if err != nil {
		return nil, fmt.Errorf("FAILED TO GET ACCESS KEY SECRET: %w", err)
	}

	secretKeyStr, err := secretKey.Plaintext(ctx)
	if err != nil {
		return nil, fmt.Errorf("FAILED TO GET SECRET KEY SECRET: %w", err)
	}

	var repoContent *dagger.Directory
	repoContent = dag.Directory()

	ctr := m.container().
		WithMountedDirectory("/sync", repoContent).
		From("minio/mc:latest").
		WithEnvVariable("MC_INSECURE", notSecure).
		WithEnvVariable("MC_HOST_"+strings.ToLower(aliasName), fmt.Sprintf("https://%s:%s@%s", accessKeyStr, secretKeyStr, endpoint)) // # pragma: allowlist secret

	output, err := ctr.WithExec([]string{
		"mc", "ls",
		aliasName,
	}).Stdout(ctx)

	if err != nil {
		panic(err)
	}

	fmt.Println("ALL BUCKETS: ", output)

	ctr = ctr.
		WithMountedDirectory("/sync", repoContent).
		WithExec([]string{
			"mc", "mirror",
			aliasName + "/" + bucketName, "/sync",
		})

	if err != nil {
		panic(err)
	}

	outputDir := ctr.Directory("/sync")

	return outputDir, nil
}

// pragma: allowlist secret
