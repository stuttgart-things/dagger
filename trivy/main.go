// A generated module for Trivy security scanning
//
// This module provides Dagger functions for scanning Docker images and filesystem
// directories using Trivy, the open-source vulnerability scanner from Aqua Security.
// It demonstrates how to configure containers with Trivy, accept optional inputs
// like scan severity and credentials, and return results as Dagger file outputs or
// raw text.
//
// The `ScanFilesystem` function scans a directory for vulnerabilities and returns
// a Trivy report as a file. The `ScanImage` function scans a container image by
// reference and returns the vulnerability report as plain text. These functions
// serve as a reference for integrating Trivy into secure CI/CD pipelines using Dagger.

package main

type Trivy struct{}
