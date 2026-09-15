// The generated client panics at init without a Dagger session, so run these
// with a placeholder: DAGGER_SESSION_PORT=1 DAGGER_SESSION_TOKEN=unused go test

package main

import (
	"encoding/base64"
	"strings"
	"testing"
)

// line builds one line of verify-attestation output around a predicate.
func line(predicate string) string {
	payload := `{"_type":"https://in-toto.io/Statement/v0.1",` +
		`"predicateType":"https://cyclonedx.org/bom",` +
		`"predicate":` + predicate + `}`
	return `{"payloadType":"application/vnd.in-toto+json","payload":"` +
		base64.StdEncoding.EncodeToString([]byte(payload)) + `","signatures":[]}`
}

func TestSinglePredicate(t *testing.T) {
	a := `{"bomFormat":"CycloneDX","serialNumber":"urn:uuid:a"}`
	b := `{"bomFormat":"CycloneDX","serialNumber":"urn:uuid:b"}`

	tests := []struct {
		name    string
		output  string
		want    string
		wantErr string
	}{
		{name: "one attestation", output: line(a) + "\n", want: a},
		{name: "the same predicate twice is one answer", output: line(a) + "\n" + line(a) + "\n", want: a},
		{name: "different predicates are an error", output: line(a) + "\n" + line(b) + "\n", wantErr: "2 different predicates"},
		{name: "no output", output: "\n", wantErr: "verified no attestation"},
		{name: "not an envelope", output: "Verification for x --\n", wantErr: "not a DSSE envelope"},
		{name: "no predicate", output: line("null"), wantErr: "carries no predicate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := singlePredicate(tt.output)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("want error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRegistryOf(t *testing.T) {
	for ref, want := range map[string]string{
		"ghcr.io/org/app@sha256:abc":     "ghcr.io",
		"localhost:5000/app@sha256:abc":  "localhost:5000",
		"localhost/app@sha256:abc":       "localhost",
		"alpine@sha256:abc":              "index.docker.io",
		"library/alpine:3.20@sha256:abc": "index.docker.io",
	} {
		if got := registryOf(ref); got != want {
			t.Errorf("registryOf(%q) = %q, want %q", ref, got, want)
		}
	}
}

func TestRequireDigest(t *testing.T) {
	digest := "@sha256:" + strings.Repeat("a", 64)
	for ref, ok := range map[string]bool{
		"ghcr.io/org/app" + digest:       true,
		"ghcr.io/org/app:1.2.3" + digest: true,
		"ghcr.io/org/app:1.2.3":          false,
		"ghcr.io/org/app@sha256:abc":     false,
		"ghcr.io/org/app" + digest + "x": false,
	} {
		if err := requireDigest(ref); (err == nil) != ok {
			t.Errorf("requireDigest(%q) = %v, want ok=%v", ref, err, ok)
		}
	}
}
