package main

import (
	"context"
	"dagger/ansible/internal/dagger"
	"fmt"
	"strings"
)

// EXECUTE ANSIBLE
func (m *Ansible) Execute(
	ctx context.Context,
	// +optional
	src *dagger.Directory,
	playbooks string,
	// +optional
	requirements *dagger.File,
	// +optional
	inventory *dagger.File,
	// +optional
	parameters string,
	// +optional
	vaultAppRoleID *dagger.Secret,
	// +optional
	vaultSecretID *dagger.Secret,
	// +optional
	vaultUrl *dagger.Secret,
) (bool, error) {

	workDir := "/src"

	ansible := m.AnsibleContainer

	if src != nil {
		ansible = ansible.
			WithDirectory(workDir, src).
			WithWorkdir(workDir)
	}

	// Optional Vault envs
	if vaultAppRoleID != nil {
		ansible = ansible.WithSecretVariable("VAULT_ROLE_ID", vaultAppRoleID)
	}
	if vaultSecretID != nil {
		ansible = ansible.WithSecretVariable("VAULT_SECRET_ID", vaultSecretID)
	}
	if vaultUrl != nil {
		ansible = ansible.WithSecretVariable("VAULT_ADDR", vaultUrl)
	}

	// Mount and install requirements
	if requirements != nil {
		reqPath := workDir + "/requirements.yml"
		ansible = ansible.WithMountedFile(reqPath, requirements).
			WithExec([]string{"ansible-galaxy", "install", "-r", reqPath})
	}

	// Mount inventory once
	if inventory != nil {
		ansible = ansible.WithMountedFile(workDir+"/inventory", inventory)
	}

	// Split playbooks and parameters
	playbookList := strings.Split(playbooks, ",")
	paramList := strings.Fields(parameters) // comma-split, but parameters are typically space-based

	// Run each playbook
	for _, playbook := range playbookList {
		playbook = strings.TrimSpace(playbook)
		if playbook == "" {
			continue
		}

		cmd := []string{"ansible-playbook", playbook, "-vv"}
		if inventory != nil {
			ansible = ansible.WithMountedFile(workDir+"/inventory", inventory)
			cmd = append(cmd, "-i", "inventory")
		}
		if len(paramList) > 0 {
			cmd = append(cmd, paramList...)
		}

		var err error
		ansible, err = ansible.WithExec(cmd).Sync(ctx)
		if err != nil {
			return false, fmt.Errorf("failed to execute playbook %s: %w", playbook, err)
		}
	}

	return true, nil
}
