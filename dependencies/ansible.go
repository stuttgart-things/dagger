package main

import (
	"context"
	"dagger/dependencies/internal/dagger"
	"fmt"
	"strings"
)

// GalaxyCollection represents an Ansible Galaxy collection metadata
type GalaxyCollection struct {
	Name           string `json:"name"`
	HighestVersion struct {
		Version string `json:"version"`
	} `json:"highest_version"`
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

// UpdateAnsibleRequirements checks for updates to Ansible collections in requirements.yaml
// and returns a report of available updates
func (m *Dependencies) UpdateAnsibleRequirements(
	ctx context.Context,
	// Path to the requirements.yaml file to check
	requirementsFile *dagger.File,
	// Optional: GitHub token for checking custom collection releases
	// +optional
	githubToken *dagger.Secret,
) (string, error) {
	// Create a container with required tools using Wolfi
	container := dag.Container().
		From("cgr.dev/chainguard/wolfi-base:latest").
		WithExec([]string{"apk", "add", "--no-cache", "bash", "curl", "jq", "yq", "python-3", "py3-pip"}).
		WithExec([]string{"pip", "install", "--break-system-packages", "ansible"})

	// Mount the requirements file
	container = container.
		WithMountedFile("/work/requirements.yaml", requirementsFile).
		WithWorkdir("/work")

	// Add GitHub token if provided
	if githubToken != nil {
		container = container.WithSecretVariable("GITHUB_TOKEN", githubToken)
	}

	// Create a script to check for updates
	script := `#!/bin/bash
set -e

echo "=== Checking Ansible Collection Updates ==="
echo ""

# Read the requirements file
REQUIREMENTS_FILE="/work/requirements.yaml"

# Function to get latest version from Ansible Galaxy
get_galaxy_version() {
  local collection=$1
  local namespace=$(echo $collection | cut -d'.' -f1)
  local name=$(echo $collection | cut -d'.' -f2)

  version=$(curl -s "https://galaxy.ansible.com/api/v3/plugin/ansible/content/published/collections/index/${namespace}/${name}/" \
    | jq -r '.highest_version.version // "unknown"' 2>/dev/null || echo "unknown")

  echo "$version"
}

# Function to get latest GitHub release version
get_github_release_version() {
  local collection_prefix=$1
  local auth_header=""

  if [ -n "$GITHUB_TOKEN" ]; then
    auth_header="-H \"Authorization: Bearer $GITHUB_TOKEN\""
  fi

  version=$(curl -s $auth_header "https://api.github.com/repos/stuttgart-things/ansible/releases" \
    | jq -r "[.[] | select(.tag_name | contains(\"${collection_prefix}\")) | .tag_name] | first" 2>/dev/null || echo "unknown")

  echo "$version"
}

# Check Galaxy collections
echo "### Ansible Galaxy Collections ###"
echo ""

yq e '.collections[] | select(.name | test("^[a-z]+\\.[a-z_]+$")) | .name' "$REQUIREMENTS_FILE" | while read -r collection; do
  if [ -n "$collection" ]; then
    current_version=$(yq e ".collections[] | select(.name == \"$collection\") | .version" "$REQUIREMENTS_FILE")
    latest_version=$(get_galaxy_version "$collection")

    if [ "$latest_version" != "unknown" ]; then
      if [ "$current_version" != "$latest_version" ]; then
        echo "🔼 $collection"
        echo "   Current: $current_version"
        echo "   Latest:  $latest_version"
        echo "   Update:  yq e -i '(.collections[] | select(.name == \"$collection\") | .version) = \"$latest_version\"' requirements.yaml"
        echo ""
      else
        echo "✅ $collection: $current_version (up-to-date)"
        echo ""
      fi
    else
      echo "⚠️  $collection: $current_version (could not check)"
      echo ""
    fi
  fi
done

# Check GitHub release collections
echo ""
echo "### Stuttgart-Things Custom Collections (GitHub Releases) ###"
echo ""

for collection_prefix in sthings-container sthings-baseos sthings-awx sthings-rke; do
  current_url=$(yq e ".collections[] | select(.name | contains(\"$collection_prefix\")) | .name" "$REQUIREMENTS_FILE")

  if [ -n "$current_url" ]; then
    # Extract version from URL pattern like: sthings-container-25.1.168.tar.gz
    current_version=$(echo "$current_url" | grep -oP "${collection_prefix}-\K[0-9.]+" || echo "unknown")

    latest_tag=$(get_github_release_version "$collection_prefix")

    if [ "$latest_tag" != "unknown" ] && [ "$latest_tag" != "null" ]; then
      latest_version=$(echo "$latest_tag" | grep -oP "${collection_prefix}-\K[0-9.]+" || echo "unknown")

      if [ "$latest_version" != "unknown" ] && [ "$current_version" != "$latest_version" ]; then
        new_url="https://github.com/stuttgart-things/ansible/releases/download/${latest_tag}/${latest_tag}.tar.gz"
        echo "🔼 $collection_prefix"
        echo "   Current: $current_version"
        echo "   Latest:  $latest_version"
        echo "   Update:  yq e -i '(.collections[] | select(.name | contains(\"$collection_prefix\")) | .name) = \"$new_url\"' requirements.yaml"
        echo ""
      else
        echo "✅ $collection_prefix: $current_version (up-to-date)"
        echo ""
      fi
    else
      echo "⚠️  $collection_prefix: $current_version (could not check GitHub releases)"
      echo ""
    fi
  fi
done

echo ""
echo "=== Check Complete ==="
`

	// Execute the script
	result, err := container.
		WithNewFile("/tmp/check-updates.sh", script, dagger.ContainerWithNewFileOpts{
			Permissions: 0755,
		}).
		WithExec([]string{"/bin/bash", "/tmp/check-updates.sh"}).
		Stdout(ctx)

	if err != nil {
		// Try to get stderr for debugging
		stderr, _ := container.Stderr(ctx)
		if stderr != "" {
			return result + "\n\nSTDERR:\n" + stderr, err
		}
		return result, err
	}

	return result, nil
}

// UpdateAnsibleRequirementsAndApply checks for updates and applies them to the requirements file
func (m *Dependencies) UpdateAnsibleRequirementsAndApply(
	ctx context.Context,
	// Path to the requirements.yaml file to update
	requirementsFile *dagger.File,
	// Collections to update (comma-separated, e.g., "community.general,kubernetes.core")
	// Use "all" to update all collections
	collectionsToUpdate string,
	// Optional: GitHub token for checking custom collection releases
	// +optional
	githubToken *dagger.Secret,
) (*dagger.File, error) {
	// Create a container with required tools using Wolfi
	container := dag.Container().
		From("cgr.dev/chainguard/wolfi-base:latest").
		WithExec([]string{"apk", "add", "--no-cache", "bash", "curl", "jq", "yq", "python-3", "py3-pip"}).
		WithExec([]string{"pip", "install", "--break-system-packages", "ansible"})

	// Mount the requirements file and copy it to a writable location
	container = container.
		WithMountedFile("/tmp/input/requirements.yaml", requirementsFile).
		WithWorkdir("/work").
		WithExec([]string{"cp", "/tmp/input/requirements.yaml", "/work/requirements.yaml"})

	// Add GitHub token if provided
	if githubToken != nil {
		container = container.WithSecretVariable("GITHUB_TOKEN", githubToken)
	}

	// Validate and prepare collections list
	updateAll := strings.ToLower(collectionsToUpdate) == "all"
	collections := []string{}
	if !updateAll {
		collections = strings.Split(collectionsToUpdate, ",")
	}

	// Create update script
	updateScript := fmt.Sprintf(`#!/bin/bash
set -e

REQUIREMENTS_FILE="/work/requirements.yaml"
UPDATE_ALL=%v
COLLECTIONS_TO_UPDATE=(%s)

# Function to check if collection should be updated
should_update() {
  local collection=$1

  if [ "$UPDATE_ALL" = "true" ]; then
    return 0
  fi

  for col in "${COLLECTIONS_TO_UPDATE[@]}"; do
    if [ "$col" = "$collection" ]; then
      return 0
    fi
  done

  return 1
}

# Function to get latest version from Ansible Galaxy
get_galaxy_version() {
  local collection=$1
  local namespace=$(echo $collection | cut -d'.' -f1)
  local name=$(echo $collection | cut -d'.' -f2)

  version=$(curl -s "https://galaxy.ansible.com/api/v3/plugin/ansible/content/published/collections/index/${namespace}/${name}/" \
    | jq -r '.highest_version.version // "unknown"' 2>/dev/null || echo "unknown")

  echo "$version"
}

# Function to get latest GitHub release version
get_github_release_version() {
  local collection_prefix=$1
  local auth_header=""

  if [ -n "$GITHUB_TOKEN" ]; then
    auth_header="-H \"Authorization: Bearer $GITHUB_TOKEN\""
  fi

  version=$(curl -s $auth_header "https://api.github.com/repos/stuttgart-things/ansible/releases" \
    | jq -r "[.[] | select(.tag_name | contains(\"${collection_prefix}\")) | .tag_name] | first" 2>/dev/null || echo "unknown")

  echo "$version"
}

echo "Updating Ansible collections..."
echo ""

# Update Galaxy collections
yq e '.collections[] | select(.name | test("^[a-z]+\\.[a-z_]+$")) | .name' "$REQUIREMENTS_FILE" | while read -r collection; do
  if [ -n "$collection" ] && should_update "$collection"; then
    latest_version=$(get_galaxy_version "$collection")

    if [ "$latest_version" != "unknown" ]; then
      yq e -i "(.collections[] | select(.name == \"$collection\") | .version) = \"$latest_version\"" "$REQUIREMENTS_FILE"
      echo "Updated $collection to $latest_version"
    fi
  fi
done

# Update GitHub release collections
for collection_prefix in sthings-container sthings-baseos sthings-awx sthings-rke; do
  if should_update "$collection_prefix"; then
    latest_tag=$(get_github_release_version "$collection_prefix")

    if [ "$latest_tag" != "unknown" ] && [ "$latest_tag" != "null" ]; then
      new_url="https://github.com/stuttgart-things/ansible/releases/download/${latest_tag}/${latest_tag}.tar.gz"
      yq e -i "(.collections[] | select(.name | contains(\"$collection_prefix\")) | .name) = \"$new_url\"" "$REQUIREMENTS_FILE"
      echo "Updated $collection_prefix to $latest_tag"
    fi
  fi
done

echo ""
echo "Update complete!"
`, updateAll, strings.Join(collections, " "))

	// Execute the update script
	container = container.
		WithNewFile("/tmp/update.sh", updateScript, dagger.ContainerWithNewFileOpts{
			Permissions: 0755,
		}).
		WithExec([]string{"/bin/bash", "/tmp/update.sh"})

	// Return the updated file
	return container.File("/work/requirements.yaml"), nil
}

// ApplyAnsibleUpdates checks for updates and applies ALL available updates to the requirements file
// This is a convenience function that automatically updates all collections to their latest versions
func (m *Dependencies) ApplyAnsibleUpdates(
	ctx context.Context,
	// Path to the requirements.yaml file to update
	requirementsFile *dagger.File,
	// Optional: GitHub token for checking custom collection releases
	// +optional
	githubToken *dagger.Secret,
) (*dagger.File, error) {
	return m.UpdateAnsibleRequirementsAndApply(ctx, requirementsFile, "all", githubToken)
}
