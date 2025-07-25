// Release module for automating semantic-release workflows and Git tag management.
//
// This module provides two main functions:
//   - Semantic: Runs semantic-release to automate versioning, changelog generation, and publishing.
//   - DeleteTag: Deletes a Git tag locally and remotely, refreshing the repository state before and after.

package main

type Release struct {
	// Base Wolfi image to use
	// +optional
	// +default="1.0.18-light"
	SemanticReleaseVersion string
	// +optional
	// +default="hoppr/semantic-release"
	BaseImage string
}
