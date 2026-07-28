/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de
*/

package collections

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type galaxyFile struct {
	Namespace   string   `yaml:"namespace"`
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Readme      string   `yaml:"readme"`
	Authors     []string `yaml:"authors"`
	Description string   `yaml:"description"`
	License     []string `yaml:"license"`
	LicenseFile string   `yaml:"license_file"`
	Tags        []string `yaml:"tags"`
	Repository  string   `yaml:"repository"`
	Homepage    string   `yaml:"homepage"`
	Issues      string   `yaml:"issues"`
}

// RENDERS GalaxyConfig FROM A collection.yaml AND PARSES THE RESULT BACK
func renderGalaxy(t *testing.T, collectionFile string) galaxyFile {
	t.Helper()

	meta := map[string]string{}
	empty := map[string]string{}

	_, _, _, _, meta, _ = ProcessCollectionFile([]byte(collectionFile), empty, empty, empty, empty, meta, empty)

	rendered := RenderTemplate(GalaxyConfig, BuildGalaxyMeta(meta, "26.4.379"))

	var parsed galaxyFile
	if err := yaml.Unmarshal([]byte(rendered), &parsed); err != nil {
		t.Fatalf("rendered galaxy.yml is not valid YAML: %v\n---\n%s", err, rendered)
	}

	return parsed
}

// A COLLECTION THAT DECLARES NOTHING BUT name/namespace MUST NOT SHIP
// PLACEHOLDER METADATA OR THE WRONG LICENSE
func TestBuildGalaxyMetaAppliesDefaults(t *testing.T) {
	parsed := renderGalaxy(t, `---
name: baseos
namespace: sthings
`)

	if parsed.Namespace != "sthings" || parsed.Name != "baseos" || parsed.Version != "26.4.379" {
		t.Errorf("unexpected identity: %+v", parsed)
	}

	if len(parsed.License) != 1 || parsed.License[0] != "Apache-2.0" {
		t.Errorf("license = %v, want [Apache-2.0]", parsed.License)
	}

	for _, author := range parsed.Authors {
		if strings.Contains(author, "example.com") {
			t.Errorf("placeholder author leaked: %q", author)
		}
	}

	for field, value := range map[string]string{
		"description": parsed.Description,
		"repository":  parsed.Repository,
		"homepage":    parsed.Homepage,
		"issues":      parsed.Issues,
	} {
		if strings.Contains(value, "example.com") || strings.Contains(value, "your collection") {
			t.Errorf("placeholder %s leaked: %q", field, value)
		}
	}
}

// VALUES DECLARED IN collection.yaml MUST WIN OVER THE DEFAULTS
func TestBuildGalaxyMetaHonoursDeclaredValues(t *testing.T) {
	parsed := renderGalaxy(t, `---
name: container
namespace: sthings
description: container collection
license:
  - MIT
tags:
  - container
  - k8s
repository: https://github.com/stuttgart-things/ansible
issues: https://github.com/stuttgart-things/ansible/issues
`)

	if len(parsed.License) != 1 || parsed.License[0] != "MIT" {
		t.Errorf("license = %v, want [MIT]", parsed.License)
	}

	if parsed.Description != "container collection" {
		t.Errorf("description = %q", parsed.Description)
	}

	if len(parsed.Tags) != 2 || parsed.Tags[0] != "container" || parsed.Tags[1] != "k8s" {
		t.Errorf("tags = %v, want [container k8s]", parsed.Tags)
	}
}

// REGRESSION: RENDERING RAN THROUGH html/template, WHICH ESCAPED THE QUOTES
// AND ANGLE BRACKETS OF AN AUTHOR ENTRY INTO &#34; / &lt; AND BROKE THE YAML
func TestRenderTemplateDoesNotHTMLEscape(t *testing.T) {
	parsed := renderGalaxy(t, `---
name: baseos
namespace: sthings
authors:
  - patrick hermann <patrick.hermann@sva.de>
`)

	if len(parsed.Authors) != 1 {
		t.Fatalf("authors = %v, want 1 entry", parsed.Authors)
	}

	if parsed.Authors[0] != "patrick hermann <patrick.hermann@sva.de>" {
		t.Errorf("author was mangled: %q", parsed.Authors[0])
	}
}
