package main

import (
	"context"
	"fmt"
)

// Version reports the Ansible release installed in the module container: the
// pinned package version (defaultAnsibleVersion) followed by `ansible --version`.
func (m *Ansible) Version(ctx context.Context) (string, error) {
	out, err := m.container(m.BaseImage, "").
		WithExec([]string{
			"sh", "-c",
			"pip3 show ansible | sed -n 's/^Version: /ansible package: /p'; ansible --version",
		}).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("ansible version failed: %w", err)
	}

	return out, nil
}
