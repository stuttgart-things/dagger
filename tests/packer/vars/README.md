# packer/vars test fixture

Exercises the `vars`, `varsFile`, and `sopsFile` parameters of `Packer.Bake`
using the `null` builder (no infra required).

`template.pkr.hcl` declares five variables:

- `name`, `environment`, `owner` — set via `varsFile` / `vars`
- `api_token`, `db_password` — set via `sopsFile`

A `shell-local` provisioner echoes each value so the packer log confirms
what was passed in.

## Files

- `template.pkr.hcl` — null-builder template with variable declarations
- `vars.yaml` — plain YAML consumed by `--vars-file`
- `secrets.plain.yaml` — plaintext source for the sops fixture (not read by tests)
- `secrets.enc.yaml` — SOPS-encrypted version, created locally (see below)

## Creating the sops fixture

```sh
# one-time: generate an age keypair
age-keygen -o age.key
export SOPS_AGE_RECIPIENTS=$(grep 'public key:' age.key | awk '{print $4}')

# encrypt
sops -e --age "$SOPS_AGE_RECIPIENTS" secrets.plain.yaml > secrets.enc.yaml
```

Keep `age.key` out of git; export its contents when running the test:

```sh
export SOPS_AGE_KEY=$(cat age.key)
```

## Example invocations

**1. `vars` only (comma-separated CLI):**

```sh
dagger call -m packer bake \
  --local-dir=. \
  --build-path=tests/packer/vars/template.pkr.hcl \
  --vars="name=from-cli,environment=ci,owner=dagger"
```

**2. `varsFile` (plain YAML):**

```sh
dagger call -m packer bake \
  --local-dir=. \
  --build-path=tests/packer/vars/template.pkr.hcl \
  --vars-file=vars.yaml
```

**3. `varsFile` + `sopsFile` + `vars` override:**

```sh
dagger call -m packer bake \
  --local-dir=. \
  --build-path=tests/packer/vars/template.pkr.hcl \
  --vars-file=vars.yaml \
  --sops-file=secrets.enc.yaml \
  --sops-age-key=env:SOPS_AGE_KEY \
  --vars="environment=prod"
```

`-var-file` entries are applied before `-var`, so the last example sets
`environment=prod` (CLI wins over `vars.yaml`).

Paths passed to `--build-path`, `--vars-file`, and `--sops-file` are
relative to the packer build directory (the parent of `build-path`).
