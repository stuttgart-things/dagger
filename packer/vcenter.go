package main

import (
	"context"
	"dagger/packer/internal/dagger"
	"fmt"
)

func (m *Packer) Vcenteroperation(
	ctx context.Context,
	// The Vm Operation to perform
	// +optional
	operation string,
	// +optional
	source string,
	// +optional
	target string,
	vcenter *dagger.Secret,
	username *dagger.Secret,
	password *dagger.Secret,
) {
	var cmd []string

	ctr := m.container("1.13.1", "linux_amd64").
		WithSecretVariable("GOVC_URL", vcenter).
		WithSecretVariable("GOVC_USERNAME", username).
		WithSecretVariable("GOVC_PASSWORD", password).
		WithEnvVariable("GOVC_INSECURE", "true")

	switch operation {
	case "move":
		if source == "" || target == "" {
			panic("source and target must be specified for move operation")
		}
		cmd = []string{
			"govc",
			"object.mv",
			source,
			target}

	case "rename":
		if source == "" || target == "" {
			panic("source (current VM name) and target (new VM name) must be specified for rename operation")
		}
		cmd = []string{
			"govc",
			"vm.change",
			"-vm", source,
			"-name", target,
		}

	case "delete":
		if source == "" {
			panic("source (VM or template path) must be specified for delete operation")
		}

		switch target {
		case "template":
			// Convert template to VM first
			unmarkCmd := []string{"govc", "vm.markastemplate", "-u=false", source}
			if _, err := ctr.WithExec(unmarkCmd).Stdout(ctx); err != nil {
				panic(fmt.Errorf("failed to unmark template: %w", err))
			}

			// Now destroy the VM
			destroyCmd := []string{"govc", "vm.destroy", source}
			cmd = destroyCmd

		case "vm", "":
			cmd = []string{"govc", "vm.destroy", source}

		default:
			panic(fmt.Errorf("unsupported delete target type: %s", target))
		}

	default:
		panic(fmt.Errorf("unsupported operation: %s", operation))
	}

	exec := ctr.WithExec(cmd)

	stdout, err := exec.Stdout(ctx)
	if err != nil {
		stderr, _ := exec.Stderr(ctx) // Try to get stderr even if stdout failed
		panic(fmt.Errorf("govc %s failed: %w\nstderr: %s", operation, err, stderr))
	}

	fmt.Printf("govc %s succeeded:\n%s\n", operation, stdout)
}
