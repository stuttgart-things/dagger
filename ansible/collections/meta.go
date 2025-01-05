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

	// GET THE CURRENT DATE
	currentDate := time.Now()
	version := fmt.Sprintf("%d.%02d.%02d-%02d%02d", currentDate.Year(), currentDate.Month(), currentDate.Day(), currentDate.Hour(), currentDate.Minute())

	return version
}
