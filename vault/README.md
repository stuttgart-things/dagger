# Vault Dagger Module

Small helpers shared across modules that need to talk to HashiCorp Vault.

Currently exposes a single function that attaches Vault AppRole environment
variables to a container so downstream tools (vals, ansible playbooks, …) can
authenticate. The module itself does no Vault login — it's pure plumbing.

## Consumers

Used as a Dagger module dependency by:

- [`ansible`](../ansible/README.md) — `Execute*` wires AppRole envs for vals lookups during playbook runs
- [`helm`](../helm/README.md) — `HelmfileOperation` wires AppRole envs for vals-backed kubeconfig retrieval

To add a new consumer:

```bash
cd <your-module>
dagger install ../vault
```

Then in your Go code:

```go
import "<your-module>/internal/dagger"

ctr = dag.Vault().WithAppRoleEnv(ctr, dagger.VaultWithAppRoleEnvOpts{
    RoleID:   roleID,   // *dagger.Secret — sets VAULT_ROLE_ID
    SecretID: secretID, // *dagger.Secret — sets VAULT_SECRET_ID
    Addr:     addr,     // *dagger.Secret — sets VAULT_ADDR
})
```

All three options are optional; nil values are skipped so callers pass only
what the downstream tool needs.

## Functions

### `with-app-role-env`

Returns the given container with Vault AppRole environment variables
(`VAULT_ROLE_ID`, `VAULT_SECRET_ID`, `VAULT_ADDR`) set as masked secrets.

```bash
# Not typically called from the CLI — consumed via module dependency.
# Example for inspection / debugging:
dagger functions -m ./vault
```
