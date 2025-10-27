package main

import (
	"context"
)

// TestKcl runs a basic KCL test to verify the container and CLI are working
func (m *Kcl) TestKcl(ctx context.Context) (string, error) {
	// Create a very simple KCL test file to avoid complex parsing
	testKcl := `name = "hello-kcl"
version = "1.0.0"
`

	return m.container().
		WithNewFile("/tmp/simple.k", testKcl).
		WithWorkdir("/tmp").
		WithExec([]string{"kcl", "run", "simple.k"}).
		Stdout(ctx)
}
