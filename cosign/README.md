# Cosign Dagger Module

A Dagger module that signs, attests and verifies container images and OCI
artefacts keylessly with Sigstore's [cosign](https://github.com/sigstore/cosign).

## Features

- Sign an artefact by digest with a Fulcio certificate and a Rekor entry
- Attach a predicate, such as a CycloneDX SBOM, as a signed attestation
- Verify signatures and attestations against an OIDC issuer and an identity
- Prove that verification refuses an identity that is not yours
- A pinned cosign release, copied from `gcr.io/projectsigstore/cosign:v<version>`

## Module Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `base-image` | string | `cgr.dev/chainguard/wolfi-base:latest` | Image the cosign binary is copied into |

Every function takes `--cosign-version` (default `2.6.5`). It stays on 2.x
until someone has checked that Kyverno's verifier reads what cosign 3 writes by
default.

Every function also takes `--registry`, `--registry-username` and
`--registry-password`. The registry is derived from `--ref` when not given. The
login runs in the same exec as cosign, with its config on a temporary mount, so
credentials never end up in a cached layer.

## Decisions

- **Digests only for signing.** `sign` and `attest` refuse a ref without
  `@sha256:`. A signature on a tag is a signature on a name somebody can move.
  Resolve the tag first with `crane digest`; for a multi-arch release that is
  the index digest. A tag in front of the digest is harmless: cosign's payload
  names the repository and the digest only.
- **The identity is the calling workflow.** Fulcio writes the workflow that
  requested the OIDC token into the certificate. Request the token in the
  repository's own workflow and pass it in. A reusable workflow template that
  signs would make every repository using it sign as the template.
- **The token is passed in, not requested.** `--identity-token` takes a token
  with audience `sigstore`. It reaches cosign as `SIGSTORE_ID_TOKEN`, never as
  an argument, and `ACTIONS_ID_TOKEN_REQUEST_*` never enter the container.
- **No default identity.** Verification needs `--certificate-oidc-issuer` and
  exactly one of `--certificate-identity` or `--certificate-identity-regexp`.
- **Verification has to be able to refuse.** `verify-refuses` succeeds only when
  cosign refuses an identity because it does not match. It fails when the
  verification passes, and also when cosign fails for any other reason, such as
  an artefact with no signature at all.
- **Never cached.** Every function is `+cache="never"` and runs cosign fresh.
  Signing writes to a registry and Rekor, and a verification is only as current
  as the state it read.
- **Locally, only verification works.** Signing needs a CI OIDC token.

## Sign

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ref` | string | Yes | Artefact to sign, pinned to a digest |
| `identity-token` | Secret | Yes | OIDC token with audience `sigstore` |

The registry credentials need push access, since the signature is stored in
the artefact's repository. Returns cosign's report, which includes the Rekor
entry and where the signature was pushed.

## Attest

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `ref` | string | Yes | - | Artefact to attest, pinned to a digest |
| `predicate` | File | Yes | - | Predicate document, e.g. a CycloneDX SBOM |
| `predicate-type` | string | Yes | - | cosign type name (`cyclonedx`, `spdxjson`, `slsaprovenance`, …) or URI |
| `identity-token` | Secret | Yes | - | OIDC token with audience `sigstore` |
| `replace` | bool | No | `false` | Replace earlier attestations of the same type |

An empty predicate is refused. `verify-attestation` refuses to choose between
attestations of one type that carry different predicates, so pass `--replace`
from any job that can attest the same digest twice, e.g. on a re-run.

## Verify

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ref` | string | Yes | Artefact to verify |
| `certificate-oidc-issuer` | string | Yes | Issuer the certificate must name |
| `certificate-identity` | string | One of | Exact identity (certificate SAN) |
| `certificate-identity-regexp` | string | One of | Identity regexp; anchor and escape it |

Returns the verified payloads as JSON. Fails when no signature verifies.

```bash
dagger call -m cosign verify \
  --ref ghcr.io/stuttgart-things/schmetterpause@sha256:72be6d814478927f4894ac586fd88bd19cb0a570792e84dda3bfc4147a07c852 \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp '^https://github\.com/stuttgart-things/schmetterpause/\.github/workflows/ci\.yml@refs/'
```

## VerifyAttestation

Takes the parameters of `verify` plus `predicate-type`, and returns the
predicate as `predicate.json`.

```bash
dagger call -m cosign verify-attestation \
  --ref ghcr.io/stuttgart-things/schmetterpause@sha256:72be6d814478927f4894ac586fd88bd19cb0a570792e84dda3bfc4147a07c852 \
  --predicate-type cyclonedx \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp '^https://github\.com/stuttgart-things/schmetterpause/\.github/workflows/ci\.yml@refs/' \
  export --path sbom.cdx.json
```

## VerifyRefuses

Takes the parameters of `verify`, with an identity that must **not** match, and
an optional `predicate-type` to check an attestation instead of a signature.

```bash
dagger call -m cosign verify-refuses \
  --ref ghcr.io/stuttgart-things/schmetterpause@sha256:72be6d814478927f4894ac586fd88bd19cb0a570792e84dda3bfc4147a07c852 \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp '^https://github\.com/stuttgart-things/not-this-repository/'
```

## GitHub Actions

Request the token and call the module in the same step, so the token lives no
longer than it has to and is not written to `GITHUB_ENV`.

```yaml
jobs:
  sign:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write   # the signature and attestation are pushed next to the image
      id-token: write   # without it there is no identity to sign as
    steps:
      - uses: actions/checkout@v5
      - name: Sign and attest
        env:
          REF: ghcr.io/myorg/app@sha256:...   # e.g. from crane digest
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          set -euo pipefail
          SIGSTORE_TOKEN=$(curl -fsSL \
            -H "Authorization: bearer ${ACTIONS_ID_TOKEN_REQUEST_TOKEN}" \
            "${ACTIONS_ID_TOKEN_REQUEST_URL}&audience=sigstore" | jq -r .value)
          echo "::add-mask::${SIGSTORE_TOKEN}"
          export SIGSTORE_TOKEN

          dagger call -m github.com/stuttgart-things/dagger/cosign@<version> sign \
            --ref "${REF}" \
            --identity-token env:SIGSTORE_TOKEN \
            --registry-username "${GITHUB_ACTOR}" \
            --registry-password env:GITHUB_TOKEN

          dagger call -m github.com/stuttgart-things/dagger/trivy@<version> \
            sbom --image-ref "${REF}" export --path sbom.cdx.json

          dagger call -m github.com/stuttgart-things/dagger/cosign@<version> attest \
            --ref "${REF}" \
            --predicate sbom.cdx.json \
            --predicate-type cyclonedx \
            --replace \
            --identity-token env:SIGSTORE_TOKEN \
            --registry-username "${GITHUB_ACTOR}" \
            --registry-password env:GITHUB_TOKEN
```

## Testing

```bash
task test-cosign   # runs tests/smoke/cosign.sh
```

## Resources

- [cosign](https://github.com/sigstore/cosign)
- [Fulcio certificate extensions](https://github.com/sigstore/fulcio/blob/main/docs/oid-info.md)
- [Dagger Documentation](https://docs.dagger.io/)
