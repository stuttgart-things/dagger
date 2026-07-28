/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package collections

import (
	"fmt"
	"time"
)

var GalaxyConfig = `### REQUIRED
namespace: {{ .namespace }}
name: {{ .name }}
version: {{ .version }}
readme: README.md
authors: {{ .authors }}

### OPTIONAL
description: {{ .description }}
license: {{ .license }}
license_file: {{ .license_file }}
tags: {{ .tags }}
dependencies: {}
repository: {{ .repository }}
documentation: {{ .documentation }}
homepage: {{ .homepage }}
issues: {{ .issues }}
build_ignore: []
`

// FALLBACKS FOR GALAXY FIELDS A COLLECTION DOES NOT DECLARE ITSELF.
// HELD AS NATIVE GO VALUES AND ENCODED ON RENDER, LIKE THE DECLARED ONES.
var GalaxyDefaults = map[string]interface{}{
	"authors":       []string{"stuttgart-things"},
	"description":   "ansible collection built by stuttgart-things",
	"license":       []string{"Apache-2.0"},
	"license_file":  "",
	"tags":          []string{},
	"repository":    "https://github.com/stuttgart-things/ansible",
	"documentation": "https://github.com/stuttgart-things/ansible",
	"homepage":      "https://github.com/stuttgart-things/ansible",
	"issues":        "https://github.com/stuttgart-things/ansible/issues",
}

// BUILDS THE RENDER DATA FOR GalaxyConfig FROM THE PARSED COLLECTION META,
// FALLING BACK TO GalaxyDefaults FOR EVERY FIELD THE COLLECTION LEAVES UNSET.
// OPTIONAL FIELDS ARE ENCODED YAML FRAGMENTS SO A COLON, '#' OR QUOTE IN A
// VALUE CANNOT BREAK OR TRUNCATE THE RENDERED galaxy.yml
func BuildGalaxyMeta(meta map[string]string, version string) map[string]interface{} {
	galaxyMeta := map[string]interface{}{
		"namespace": meta["namespace"],
		"name":      meta["name"],
		"version":   version,
	}

	for field, fallback := range GalaxyDefaults {
		// meta HOLDS VALUES ALREADY ENCODED BY setMetaString/setMetaList
		if value := meta[field]; value != "" {
			galaxyMeta[field] = value
			continue
		}

		encoded, ok := encodeGalaxyValue(field, fallback)
		if !ok {
			continue
		}
		galaxyMeta[field] = encoded
	}

	return galaxyMeta
}

func GenerateSemanticVersion() string {
	// GET THE CURRENT DATE AND TIME
	currentDate := time.Now()

	// MAJOR: YEAR IN TWO DIGITS (2025 -> 25)
	major := currentDate.Year() % 100

	// MINOR: DAY OF THE WEEK (0 FOR SUNDAY TO 6 FOR SATURDAY)
	minor := int(currentDate.Weekday())

	// PATCH: A NUMBER DERIVED FROM THE HOUR AND MINUTE (TO ENSURE UNIQUENESS WITHIN A DAY)
	patch := currentDate.Hour()*60 + currentDate.Minute() // Total minutes since midnight

	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}
