package main

import (
	"context"
	"dagger/ansible/internal/dagger"
	"fmt"
	"path/filepath"
	"strings"
)

// executePlaybooks is a private helper that sets up and runs Ansible playbooks.
// It returns the container after all playbooks have been executed.
func (m *Ansible) executePlaybooks(
	ctx context.Context,
	src *dagger.Directory,
	playbooks string,
	requirements *dagger.File,
	inventory *dagger.File,
	parameters string,
	vaultAppRoleID *dagger.Secret,
	vaultSecretID *dagger.Secret,
	vaultUrl *dagger.Secret,
	sshUser *dagger.Secret,
	sshPassword *dagger.Secret, // pragma: allowlist secret
	ansibleVersion string,
) (*dagger.Container, error) {

	workDir := "/src"

	ansible := m.container(m.BaseImage, ansibleVersion)

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
	if vaultSecretID != nil { // pragma: allowlist secret
		ansible = ansible.WithSecretVariable("VAULT_SECRET_ID", vaultSecretID)
	}
	if vaultUrl != nil {
		ansible = ansible.WithSecretVariable("VAULT_ADDR", vaultUrl)
	}

	// Set SSH credentials as env vars for lookup inside extra-vars
	if sshUser != nil {
		ansible = ansible.WithSecretVariable("ANSIBLE_USER", sshUser)
	}
	if sshPassword != nil { // pragma: allowlist secret
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

		if sshUser != nil && sshPassword != nil { // pragma: allowlist secret
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
			return nil, fmt.Errorf("failed to execute playbook %s: %w", playbook, err)
		}
	}

	return ansible, nil
}

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
	// The Ansible version
	// +optional
	// +default="11.11.0"
	ansibleVersion string,
) (bool, error) {

	_, err := m.executePlaybooks(ctx, src, playbooks, requirements, inventory, parameters, vaultAppRoleID, vaultSecretID, vaultUrl, sshUser, sshPassword, ansibleVersion)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ExecuteAndExport runs Ansible playbooks and exports specified files from the container.
// Returns a flat directory containing the exported files (using basenames).
func (m *Ansible) ExecuteAndExport(
	ctx context.Context,
	playbooks string,
	// Comma-separated list of file paths to export from the container
	exportPaths string,
	// +optional
	src *dagger.Directory,
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
	// The Ansible version
	// +optional
	// +default="11.11.0"
	ansibleVersion string,
) (*dagger.Directory, error) {

	ctr, err := m.executePlaybooks(ctx, src, playbooks, requirements, inventory, parameters, vaultAppRoleID, vaultSecretID, vaultUrl, sshUser, sshPassword, ansibleVersion)
	if err != nil {
		return nil, err
	}

	// Extract files from the container into a flat directory
	exportDir := dag.Directory()

	paths := strings.Split(exportPaths, ",")
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		basename := filepath.Base(p)
		exportedFile := ctr.File(p)
		exportDir = exportDir.WithFile(basename, exportedFile)
	}

	return exportDir, nil
}
