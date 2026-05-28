// Vault provides small helpers shared across modules that need to talk to
// HashiCorp Vault — currently focused on attaching AppRole environment
// variables to a container so downstream tools (vals, ansible playbooks, …)
// can authenticate. The intent is to keep a single canonical place for
// Vault-related plumbing rather than reimplementing the same env-var dance
// in every consumer module.

package main

import (
	"dagger/vault/internal/dagger"
)

type Vault struct{}

// WithAppRoleEnv returns the given container with Vault AppRole environment
// variables (VAULT_ROLE_ID, VAULT_SECRET_ID, VAULT_ADDR) set as masked
// secrets. Any nil argument is skipped — pass only what the downstream tool
// needs. The container itself does no Vault login; this is plumbing for
// tools like `vals` and ansible playbooks that read these envs.
func (m *Vault) WithAppRoleEnv(
	ctr *dagger.Container,
	// +optional
	roleID *dagger.Secret,
	// +optional
	secretID *dagger.Secret,
	// +optional
	addr *dagger.Secret,
) *dagger.Container {
	if roleID != nil {
		ctr = ctr.WithSecretVariable("VAULT_ROLE_ID", roleID)
	}
	if secretID != nil { // pragma: allowlist secret
		ctr = ctr.WithSecretVariable("VAULT_SECRET_ID", secretID)
	}
	if addr != nil {
		ctr = ctr.WithSecretVariable("VAULT_ADDR", addr)
	}
	return ctr
}
