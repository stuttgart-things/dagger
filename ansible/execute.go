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
	// +optional
	sshUser *dagger.Secret,
	// +optional
	sshPassword *dagger.Secret,
) (bool, error) {

	workDir := "/src"

	ansible := m.container(m.BaseImage)

	if src != nil {
		ansible = ansible.
			WithDirectory(workDir, src).
			WithWorkdir(workDir).
			WithEnvVariable("ANSIBLE_HOST_KEY_CHECKING", "False")
	}

	// OPTIONAL VAULT ENVS
	if vaultAppRoleID != nil {
		ansible = ansible.WithSecretVariable("VAULT_ROLE_ID", vaultAppRoleID)
	}
	if vaultSecretID != nil {
		ansible = ansible.WithSecretVariable("VAULT_SECRET_ID", vaultSecretID)
	}
	if vaultUrl != nil {
		ansible = ansible.WithSecretVariable("VAULT_ADDR", vaultUrl)
	}

	// Set SSH credentials as env vars for lookup inside extra-vars
	if sshUser != nil {
		ansible = ansible.WithSecretVariable("ANSIBLE_USER", sshUser)
	}
	if sshPassword != nil {
		ansible = ansible.WithSecretVariable("ANSIBLE_PASSWORD", sshPassword)
	}

	// MOUNT AND INSTALL REQUIREMENTS
	if requirements != nil {
		reqPath := workDir + "/requirements.yml"
		ansible = ansible.WithMountedFile(reqPath, requirements).
			WithExec([]string{"ansible-galaxy", "install", "-r", reqPath})
	}

	// MOUNT INVENTORY
	if inventory != nil {
		ansible = ansible.WithMountedFile(workDir+"/inventory", inventory)
	}

	// SPLIT PLAYBOOKS
	playbookList := strings.Split(playbooks, ",")

	// RUN EACH PLAYBOOK
	for _, playbook := range playbookList {
		playbook = strings.TrimSpace(playbook)
		if playbook == "" {
			continue
		}

		cmd := []string{"ansible-playbook", playbook, "-vv"}
		if inventory != nil {
			cmd = append(cmd, "-i", "inventory")
		}

		// Build extra-vars string
		var extraVars []string

		if sshUser != nil && sshPassword != nil {
			extraVars = append(extraVars,
				"ansible_user='{{ lookup(\"env\", \"ANSIBLE_USER\") }}'",
				"ansible_password='{{ lookup(\"env\", \"ANSIBLE_PASSWORD\") }}'",
			)
		}

		if parameters != "" {
			extraVars = append(extraVars, parameters)
		}

		if len(extraVars) > 0 {
			cmd = append(cmd, "--extra-vars", strings.Join(extraVars, " "))
		}

		var err error
		ansible, err = ansible.WithExec(cmd).Sync(ctx)
		if err != nil {
			return false, fmt.Errorf("failed to execute playbook %s: %w", playbook, err)
		}
	}

	return true, nil
}
