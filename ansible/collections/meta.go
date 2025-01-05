/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package collections

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
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

	// Major: Last two digits of the year
	major := currentDate.Year() % 100

	// Minor: Day of the week (0 for Sunday, 6 for Saturday)
	minor := int(currentDate.Weekday())

	// Patch: Hash of the Unix timestamp, truncated to 3 digits
	timestamp := strconv.FormatInt(currentDate.UnixNano(), 10) // Unix timestamp with nanoseconds
	hash := sha256.Sum256([]byte(timestamp))
	hashString := hex.EncodeToString(hash[:])
	patch, _ := strconv.Atoi(hashString[:3]) // Convert first 3 hex digits to an integer

	// Format the version
	version := fmt.Sprintf("%02d.%01d.%03d", major, minor, patch)

	return version
}
