package main

import (
	"dagger/oci/internal/dagger"
	"fmt"
)

// tlsCerts generates a self-signed TLS certificate and returns the cert and key as a directory.
func (m *Oci) tlsCerts() *dagger.Directory {
	return dag.Container().
		From("cgr.dev/chainguard/wolfi-base:latest").
		WithExec([]string{"apk", "add", "--no-cache", "openssl"}).
		WithExec([]string{"mkdir", "-p", "/tls"}).
		WithExec([]string{
			"openssl", "req", "-x509", "-newkey", "rsa:2048", "-nodes",
			"-keyout", "/tls/key.pem",
			"-out", "/tls/cert.pem",
			"-days", "1",
			"-subj", "/CN=registry",
			"-addext", "subjectAltName=DNS:registry",
		}).
		Directory("/tls")
}

// RegistryService starts a local OCI registry (zot) with TLS as a Dagger service container.
// A self-signed certificate is generated for the registry hostname.
// Useful for testing OCI artifact pushes without requiring external registry credentials.
func (m *Oci) RegistryService(
	// Zot registry image to use
	// +optional
	// +default="ghcr.io/project-zot/zot-linux-amd64:latest"
	image string,
	// Port to expose the registry on
	// +optional
	// +default=5000
	port int,
) *dagger.Service {
	if image == "" {
		image = "ghcr.io/project-zot/zot-linux-amd64:latest"
	}

	if port == 0 {
		port = 5000
	}

	zotConfig := fmt.Sprintf(`{
  "distSpecVersion": "1.1.0",
  "storage": { "rootDirectory": "/var/lib/registry" },
  "http": {
    "address": "0.0.0.0",
    "port": "%d",
    "tls": {
      "cert": "/etc/zot/tls/cert.pem",
      "key": "/etc/zot/tls/key.pem"
    }
  },
  "log": { "level": "warn" }
}`, port)

	tlsDir := m.tlsCerts()

	return dag.Container().
		From(image).
		WithNewFile("/etc/zot/tls-config.json", zotConfig).
		WithDirectory("/etc/zot/tls", tlsDir).
		WithExposedPort(port).
		WithDefaultArgs([]string{"/usr/local/bin/zot-linux-amd64", "serve", "/etc/zot/tls-config.json"}).
		AsService()
}
