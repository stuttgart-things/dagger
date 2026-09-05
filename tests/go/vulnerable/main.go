// Package main is a fixture for the go module's govulncheck function.
//
// It exists to prove that --fail-on-finding actually fires. The dependency
// below is pinned to a version with a known vulnerability that is reachable
// from this call graph: yaml.Unmarshal in gopkg.in/yaml.v2 v2.2.1 is affected
// by GO-2020-0036 (unbounded alias expansion, CVE-2019-11253).
//
// Do not "fix" this by bumping the dependency -- a clean tree here would make
// the regression test silently pass forever. The happy path is covered by
// tests/go/calculator, which has no dependencies at all.
package main

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

func main() {
	var out map[string]string

	// The reachable call. govulncheck reports GO-2020-0036 because of this
	// line; delete it and the finding drops to "not reachable".
	if err := yaml.Unmarshal([]byte("greeting: hello\n"), &out); err != nil {
		panic(err)
	}

	fmt.Println(out["greeting"])
}
