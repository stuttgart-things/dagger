package main

import "strings"

// splitValues splits comma-separated key=value pairs.
func splitValues(values string) []string {
	var result []string
	for _, pair := range strings.Split(values, ",") {
		pair = strings.TrimSpace(pair)
		if pair != "" {
			result = append(result, pair)
		}
	}
	return result
}
