#!/bin/bash
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
