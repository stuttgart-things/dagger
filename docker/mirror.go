package main

import (
	"context"
	"dagger/docker/internal/dagger"
	"fmt"
	"strings"
)

// MirrorApkPackages fetches APK packages and their dependencies from an Alpine-based image
// and creates a mirror archive suitable for offline installation
func (m *Docker) MirrorApkPackages(
	ctx context.Context,
	// Alpine-based Docker image (e.g., "python:3.13.7-alpine", "alpine:3.18")
	// +optional
	// +default="python:3.13.7-alpine"
	image string,
	// Source directory containing packages.yaml file
	// +optional
	source *dagger.Directory,
	// Comma-separated list of APK packages to fetch (e.g., "curl,wget")
	// +optional
	apkPackages string,
	// YAML file containing package list (packages: [pkg1, pkg2])
	// +optional
	packagesFile *dagger.File,
	// Alpine repository: main, community, or testing
	// +optional
	// +default="main"
	apkRepo string,
	// Name of the output archive file (without extension)
	// +optional
	// +default="apk-packages"
	archiveName string,
) (*dagger.File, error) {

	// Set default values if not provided
	// Validate repository
	if apkRepo != "main" && apkRepo != "community" && apkRepo != "testing" {
		return nil, fmt.Errorf("invalid repository '%s'. Must be one of: main, community, testing", apkRepo)
	}

	var apkPackagesSpace string

	// Read packages from source directory if provided
	if source != nil {
		// Look for packages.yaml in the source directory
		packagesFile = source.File("packages.yaml")
	}

	// Read packages from YAML file if provided
	if packagesFile != nil {
		yamlContent, err := packagesFile.Contents(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to read packages file: %w", err)
		}

		// Parse YAML to extract packages
		// Expected format: packages: [pkg1, pkg2, pkg3] or packages:\n  - pkg1\n  - pkg2
		packages, err := parsePackagesFromYAML(yamlContent)
		if err != nil {
			return nil, fmt.Errorf("failed to parse packages YAML: %w", err)
		}
		apkPackagesSpace = strings.Join(packages, " ")
	} else if apkPackages != "" {
		// Convert comma-separated packages to space-separated
		apkPackagesSpace = strings.ReplaceAll(apkPackages, ",", " ")
	} else {
		return nil, fmt.Errorf("either apkPackages, packagesFile, or source (with packages.yaml) must be provided")
	}

	// STEP 1: Get Alpine VERSION_ID from base image
	baseContainer := dag.Container().From(image)
	versionID, err := baseContainer.
		WithExec([]string{"sh", "-c", "grep '^VERSION_ID=' /etc/os-release | cut -d= -f2 | tr -d '\"' | cut -d. -f1,2"}).
		Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Alpine version: %w", err)
	}
	versionID = strings.TrimSpace(versionID)

	// STEP 2: Prepare mirror directory structure
	mirrorPath := fmt.Sprintf("/mirror/v%s/%s/x86_64", versionID, apkRepo)

	// STEP 3: Fetch APK packages and APKINDEX
	fetchScript := fmt.Sprintf(`
cd %s
echo "[INFO] Using Alpine v$VERSION_ID repository: $APK_REPO"

# Configure repositories
if [ "$APK_REPO" = "main" ]; then
  echo "http://dl-cdn.alpinelinux.org/alpine/v$VERSION_ID/main" > /etc/apk/repositories
else
  # For community/testing, we need main as well for dependencies
  echo "http://dl-cdn.alpinelinux.org/alpine/v$VERSION_ID/main" > /etc/apk/repositories
  echo "http://dl-cdn.alpinelinux.org/alpine/v$VERSION_ID/$APK_REPO" >> /etc/apk/repositories
fi

echo "[INFO] Repository configuration:"
cat /etc/apk/repositories

apk update
echo "[INFO] Fetching APK_PACKAGES: $APK_PACKAGES"
apk fetch --recursive --output . $APK_PACKAGES

echo "[INFO] Fetching APKINDEX for $APK_REPO repository"
wget -q http://dl-cdn.alpinelinux.org/alpine/v$VERSION_ID/$APK_REPO/x86_64/APKINDEX.tar.gz

# Also fetch main APKINDEX if we're not using main repo (for dependencies)
if [ "$APK_REPO" != "main" ]; then
  echo "[INFO] Also fetching main APKINDEX for dependencies"
  wget -q -O APKINDEX-main.tar.gz http://dl-cdn.alpinelinux.org/alpine/v$VERSION_ID/main/x86_64/APKINDEX.tar.gz
fi
`, mirrorPath)

	// Execute fetch in container
	fetchContainer := baseContainer.
		WithEnvVariable("VERSION_ID", versionID).
		WithEnvVariable("APK_REPO", apkRepo).
		WithEnvVariable("APK_PACKAGES", apkPackagesSpace).
		WithExec([]string{"sh", "-c", fmt.Sprintf("mkdir -p %s", mirrorPath)}).
		WithExec([]string{"sh", "-euxc", fetchScript})

	// Get the mirror directory
	mirrorDir := fetchContainer.Directory("/mirror")

	// Get list of downloaded .apk files for summary
	downloadedFiles, err := fetchContainer.
		WithExec([]string{"sh", "-c", fmt.Sprintf("ls -1 %s/*.apk 2>/dev/null | xargs -n1 basename || echo 'No packages downloaded'", mirrorPath)}).
		Stdout(ctx)
	if err != nil {
		downloadedFiles = "Unable to list downloaded packages"
	}

	// Get current timestamp
	timestamp, err := dag.Container().
		From("alpine:latest").
		WithExec([]string{"date", "+%Y-%m-%d %H:%M:%S UTC"}).
		Stdout(ctx)
	if err != nil {
		timestamp = "Unknown"
	}
	timestamp = strings.TrimSpace(timestamp)

	// STEP 4: Create summary file
	summaryContent := fmt.Sprintf(`APK Mirror Summary
==================

Generated: %s
Base Image: %s
Alpine Version: %s
Repository: %s
Requested Packages: %s

Downloaded APK Files:
%s

Mirror Structure:
- v%s/%s/x86_64/
  - *.apk (package files)
  - APKINDEX.tar.gz (package index)

Usage:
------
1. Extract this archive to your Alpine-based system
2. Configure APK to use the local mirror:
   echo "/path/to/mirror" > /etc/apk/repositories
3. Install packages: apk add <package-name>
`,
		timestamp,
		image,
		versionID,
		apkRepo,
		apkPackagesSpace,
		downloadedFiles,
		versionID,
		apkRepo,
	)

	// STEP 5: Create zip archive with summary
	zipContainer := dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "--no-cache", "zip"}).
		WithDirectory("/work", mirrorDir).
		WithWorkdir("/work").
		WithNewFile("/work/SUMMARY.txt", summaryContent).
		WithExec([]string{"zip", "-r", fmt.Sprintf("/tmp/%s.zip", archiveName), "."})

	// Return the zip file
	return zipContainer.File(fmt.Sprintf("/tmp/%s.zip", archiveName)), nil
}

// parsePackagesFromYAML extracts package names from YAML content
// Supports both array format: packages: [pkg1, pkg2]
// and list format: packages:\n  - pkg1\n  - pkg2
func parsePackagesFromYAML(yamlContent string) ([]string, error) {
	var packages []string
	lines := strings.Split(yamlContent, "\n")

	inPackagesList := false
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for array format: packages: [pkg1, pkg2, pkg3]
		if strings.HasPrefix(line, "packages:") {
			// Extract content after "packages:"
			content := strings.TrimPrefix(line, "packages:")
			content = strings.TrimSpace(content)

			// Check if it's an array format [...]
			if strings.HasPrefix(content, "[") && strings.HasSuffix(content, "]") {
				content = strings.Trim(content, "[]")
				// Split by comma and clean up
				for _, pkg := range strings.Split(content, ",") {
					pkg = strings.TrimSpace(pkg)
					pkg = strings.Trim(pkg, "\"'")
					if pkg != "" {
						packages = append(packages, pkg)
					}
				}
				return packages, nil
			}
			// Otherwise it's list format, continue reading
			inPackagesList = true
			continue
		}

		// Parse list format: - pkg1
		if inPackagesList && strings.HasPrefix(line, "-") {
			pkg := strings.TrimPrefix(line, "-")
			pkg = strings.TrimSpace(pkg)
			pkg = strings.Trim(pkg, "\"'")
			if pkg != "" {
				packages = append(packages, pkg)
			}
		} else if inPackagesList && !strings.HasPrefix(line, "-") && line != "" {
			// End of packages list
			break
		}
	}

	if len(packages) == 0 {
		return nil, fmt.Errorf("no packages found in YAML file")
	}

	return packages, nil
}
