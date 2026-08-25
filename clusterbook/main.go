package main

import (
	"crypto/tls"
	"net/http"
)

type Clusterbook struct {
	// Skip TLS verification when talking to an https clusterbook.
	Insecure bool
}

// New builds the module.
//
// insecure exists because the deployed clusterbook sits behind a Gateway whose
// certificate is signed by the internal Vault PKI: curl on a lab machine
// accepts it, the Dagger container does not, and the failure reads
// "certificate signed by unknown authority" rather than anything about trust
// stores. Callers on plain http are unaffected.
func New(
	// skip TLS verification (internal CA)
	// +optional
	insecure bool,
) *Clusterbook {
	return &Clusterbook{Insecure: insecure}
}

func (c *Clusterbook) httpClient() *http.Client {
	if !c.Insecure {
		return http.DefaultClient
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // #nosec G402 -- opt-in, internal CA
		},
	}
}
