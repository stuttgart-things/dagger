// A Dagger module for Backstage TechDocs
//
// Builds a repository's TechDocs site in CI and publishes it to S3-compatible
// storage (MinIO), so Backstage runs with `techdocs.builder: external` and
// only reads the finished HTML instead of running mkdocs in its own pod.
//
// The site is built with `techdocs-cli generate --no-docker` (mkdocs +
// mkdocs-techdocs-core) and uploaded with `techdocs-cli publish
// --publisher-type awsS3` to <bucket>/<namespace>/<kind>/<name>/, the path
// Backstage's awsS3 publisher reads.
//
// Functions:
//   - Generate: builds the site of a source directory
//   - Entity: reads <namespace>/<kind>/<name> from catalog-info.yaml
//   - Publish: uploads a generated site for an entity
//   - GenerateAndPublish: all of the above in one call

package main

import (
	"context"
	"dagger/techdocs/internal/dagger"
	"fmt"
	"time"
)

type Techdocs struct {
	BaseImage string
}

// Generate builds the TechDocs site of src (the directory holding mkdocs.yml)
// and returns it.
func (m *Techdocs) Generate(
	ctx context.Context,
	// Repository (or docs) directory containing mkdocs.yml
	src *dagger.Directory,
	// Directory of mkdocs.yml inside src
	// +optional
	// +default="."
	sourceDir string,
) (*dagger.Directory, error) {
	ctr, err := m.container(ctx)
	if err != nil {
		return nil, err
	}

	// techdocs-cli rewrites mkdocs.yml (it adds the techdocs-core plugin if it
	// is missing), so src is copied in rather than mounted read-only.
	return ctr.
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{
			"techdocs-cli", "generate",
			"--no-docker",
			"--source-dir", sourceDir,
			"--output-dir", "/site",
			"--verbose",
		}).
		Directory("/site"), nil
}

// Publish uploads a site built by Generate to an S3-compatible bucket, under
// the path Backstage reads for entity. Stale files of a previous publish of
// the same entity are removed.
func (m *Techdocs) Publish(
	ctx context.Context,
	// Site directory returned by Generate
	site *dagger.Directory,
	// Entity as <namespace>/<kind>/<name>, e.g. default/component/flux
	entity string,
	// Bucket Backstage reads (techdocs.publisher.awsS3.bucketName)
	bucket string,
	// S3 endpoint, e.g. https://minio.example.com
	endpoint string,
	// S3 access key of a user that may write to bucket
	accessKey *dagger.Secret,
	// S3 secret key
	secretKey *dagger.Secret,
	// Region; MinIO accepts any value
	// +optional
	// +default="us-east-1"
	region string,
	// PEM CA certificate for an endpoint signed by a private CA
	// +optional
	caCert *dagger.File,
) (string, error) {
	ctr, err := m.container(ctx)
	if err != nil {
		return "", err
	}

	if caCert != nil {
		ctr = ctr.
			WithMountedFile("/etc/ssl/custom/ca.crt", caCert).
			WithEnvVariable("NODE_EXTRA_CA_CERTS", "/etc/ssl/custom/ca.crt")
	}

	return ctr.
		WithMountedDirectory("/site", site).
		WithSecretVariable("AWS_ACCESS_KEY_ID", accessKey).
		WithSecretVariable("AWS_SECRET_ACCESS_KEY", secretKey).
		WithEnvVariable("AWS_REGION", region).
		// An upload is a side effect: a cached run would skip it after the
		// bucket was emptied or another commit published in between.
		WithEnvVariable("TECHDOCS_PUBLISH_AT", time.Now().UTC().Format(time.RFC3339Nano)).
		WithExec([]string{
			"techdocs-cli", "publish",
			"--publisher-type", "awsS3",
			"--storage-name", bucket,
			"--entity", entity,
			"--directory", "/site",
			"--awsEndpoint", endpoint,
			// MinIO serves buckets by path, not by virtual host.
			"--awsS3ForcePathStyle",
		}).
		// techdocs-cli logs to stderr.
		CombinedOutput(ctx)
}

// GenerateAndPublish builds the site of src and publishes it. The entity is
// read from catalog-info.yaml unless given.
func (m *Techdocs) GenerateAndPublish(
	ctx context.Context,
	// Repository directory containing mkdocs.yml and catalog-info.yaml
	src *dagger.Directory,
	// Bucket Backstage reads (techdocs.publisher.awsS3.bucketName)
	bucket string,
	// S3 endpoint, e.g. https://minio.example.com
	endpoint string,
	// S3 access key of a user that may write to bucket
	accessKey *dagger.Secret,
	// S3 secret key
	secretKey *dagger.Secret,
	// Entity as <namespace>/<kind>/<name>; read from catalogFile if empty
	// +optional
	entity string,
	// Catalog file inside src to read the entity from
	// +optional
	// +default="catalog-info.yaml"
	catalogFile string,
	// Directory of mkdocs.yml inside src
	// +optional
	// +default="."
	sourceDir string,
	// Region; MinIO accepts any value
	// +optional
	// +default="us-east-1"
	region string,
	// PEM CA certificate for an endpoint signed by a private CA
	// +optional
	caCert *dagger.File,
) (string, error) {
	if entity == "" {
		var err error
		entity, err = m.Entity(ctx, src, catalogFile)
		if err != nil {
			return "", err
		}
	}

	site, err := m.Generate(ctx, src, sourceDir)
	if err != nil {
		return "", fmt.Errorf("generate: %w", err)
	}

	out, err := m.Publish(ctx, site, entity, bucket, endpoint, accessKey, secretKey, region, caCert)
	if err != nil {
		return "", fmt.Errorf("publish %s: %w", entity, err)
	}
	return fmt.Sprintf("published %s to %s/%s\n%s", entity, endpoint, bucket, out), nil
}
