package main

import "testing"

// An inline `:tag` in an OCI source is not a pin -- kcl reads the version only
// from a `tag` query parameter, so the suffix stays part of the repository
// path and the reference resolves to the newest published tag. These cases
// pin down the rewrite, and just as importantly the references it must leave
// alone. See stuttgart-things/kcl#231.
func TestNormalizeOciSource(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"inline tag is rewritten", "ghcr.io/stuttgart-things/harvester-vm:0.3.0", "oci://ghcr.io/stuttgart-things/harvester-vm?tag=0.3.0"},
		{"scheme is preserved", "oci://ghcr.io/stuttgart-things/harvester-vm:0.3.0", "oci://ghcr.io/stuttgart-things/harvester-vm?tag=0.3.0"},
		{"prerelease tag survives", "ghcr.io/o/m:0.3.0-rc.1", "oci://ghcr.io/o/m?tag=0.3.0-rc.1"},
		{"scheme is added when missing", "ghcr.io/stuttgart-things/harvester-vm", "oci://ghcr.io/stuttgart-things/harvester-vm"},

		{"existing query is left alone", "ghcr.io/stuttgart-things/harvester-vm?tag=0.2.0", "oci://ghcr.io/stuttgart-things/harvester-vm?tag=0.2.0"},
		{"digest is left alone", "ghcr.io/o/m@sha256:abc123", "oci://ghcr.io/o/m@sha256:abc123"},
		{"trailing colon is left alone", "ghcr.io/o/m:", "oci://ghcr.io/o/m:"},
		{"no repository path is left alone", "harvester-vm:0.3.0", "oci://harvester-vm:0.3.0"},

		// The colon in a registry port is not a tag. Only the final path
		// segment is examined, so the port survives either way.
		{"registry port is not a tag", "localhost:5000/foo/bar", "oci://localhost:5000/foo/bar"},
		{"registry port plus a real tag", "localhost:5000/foo/bar:1.2.3", "oci://localhost:5000/foo/bar?tag=1.2.3"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeOciSource(tc.in); got != tc.want {
				t.Errorf("normalizeOciSource(%q)\n got %q\nwant %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestShellQuote(t *testing.T) {
	// `?` is a glob character, so the rewritten source must reach `sh -c`
	// quoted rather than relying on an unmatched glob passing through.
	if got, want := shellQuote("oci://ghcr.io/o/m?tag=1.0"), `'oci://ghcr.io/o/m?tag=1.0'`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := shellQuote("it's"), `'it'"'"'s'`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
