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
authors:
- your name <example@domain.com>

### OPTIONAL
description: your collection description
license:
- GPL-2.0-or-later
license_file: ''
tags: []
dependencies: {}
repository: http://example.com/repository
documentation: http://docs.example.com
homepage: http://example.com
issues: http://example.com/issue/tracker
build_ignore: []
`

func GenerateSemanticVersion() string {
	// Get the current date and time
	currentDate := time.Now()

	// Major: Year in two digits (2025 -> 25)
	major := currentDate.Year() % 100

	// Minor: Day of the week (0 for Sunday to 6 for Saturday)
	minor := int(currentDate.Weekday())

	// Patch: A number derived from the hour and minute (to ensure uniqueness within a day)
	patch := currentDate.Hour()*60 + currentDate.Minute() // Total minutes since midnight

	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}
