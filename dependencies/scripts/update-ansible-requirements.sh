#!/bin/bash
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
