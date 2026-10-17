# techdocs

Builds a repository's [Backstage TechDocs](https://backstage.io/docs/features/techdocs/)
site in CI and publishes it to S3-compatible storage (MinIO). Backstage then
runs with `techdocs.builder: external` and an `awsS3` publisher: it only reads
the finished HTML and never runs mkdocs in its own pod.

Tools (pinned, Renovate-managed in `container.go`): `mkdocs-techdocs-core`,
`@techdocs/cli`, Node.js from nodejs.org (checked against `SHASUMS256.txt`) on
`python:3.13-slim`.

## Functions

| Function | What it does |
|---|---|
| `generate` | `techdocs-cli generate --no-docker` on `--src` (mkdocs.yml in `--source-dir`, default `.`); returns the site directory |
| `entity` | `<namespace>/<kind>/<name>` of the entity in `catalog-info.yaml` that carries `backstage.io/techdocs-ref` (namespace defaults to `default`) |
| `publish` | `techdocs-cli publish --publisher-type awsS3 --awsS3ForcePathStyle` of a site to `<bucket>/<namespace>/<kind>/<name>/`; removes stale files of the previous publish |
| `generate-and-publish` | all of the above; `--entity` is read from the catalog file unless given |

`publish` never comes from cache: an upload is a side effect.

## Usage

```bash
# Build only (e.g. in a PR), export the site to look at it
dagger call -m github.com/stuttgart-things/dagger/techdocs@<version> \
  generate --src . export --path ./site

# Which path will Backstage read?
dagger call -m github.com/stuttgart-things/dagger/techdocs@<version> \
  entity --src .

# Build and publish (main branch)
dagger call -m github.com/stuttgart-things/dagger/techdocs@<version> \
  generate-and-publish \
  --src . \
  --bucket techdocs \
  --endpoint https://minio.example.com \
  --access-key env:TECHDOCS_S3_ACCESS_KEY \
  --secret-key env:TECHDOCS_S3_SECRET_KEY \
  --ca-cert ./ca.crt   # only for an endpoint signed by a private CA
```

The S3 user needs `s3:ListBucket` and `s3:GetBucketLocation` on the bucket and
`s3:GetObject`, `s3:PutObject`, `s3:DeleteObject` on its objects (delete is
used to remove stale pages).

## Backstage side

```yaml
techdocs:
  builder: external
  publisher:
    type: awsS3
    awsS3:
      bucketName: techdocs
      endpoint: https://minio.example.com
      s3ForcePathStyle: true
      region: us-east-1
      credentials:
        accessKeyId: ${TECHDOCS_S3_ACCESS_KEY_ID}
        secretAccessKey: ${TECHDOCS_S3_SECRET_ACCESS_KEY}
```

Backstage needs read access only.
