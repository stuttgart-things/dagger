// A Dagger module for rendering Go templates with variable substitution
//
// This module provides template rendering capabilities using Go's text/template package.
// It supports loading templates from local files or HTTPS URLs, and accepts variables
// from YAML files or command-line parameters.
//
// Templates can use any file extension, with optional .tmpl suffix that will be
// automatically removed in the output. Variables can be provided via a YAML file,
// command-line key=value pairs, or both (command-line takes precedence).
//
// The module can be called from the dagger CLI or from any of the Dagger SDKs.

package main

type Templating struct{}
