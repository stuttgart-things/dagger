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
