# Hugo Dagger Module

This module provides Dagger functions for Hugo static site generation including site initialization, building, and deployment to various targets.

## Features

- ✅ Hugo site initialization with themes
- ✅ Static content generation and building
- ✅ Local development server with live reload
- ✅ MinIO bucket synchronization for assets
- ✅ Multi-stage build and export workflows
- ✅ Theme management and configuration

## Prerequisites

- Dagger CLI installed
- Docker runtime available
- Hugo configuration files

## Quick Start

### Initialize Site

```bash
# Initialize Hugo site structure
dagger call -m hugo init-site \
  --name test \
  --config tests/hugo/hugo.toml \
  --content tests/hugo/content \
  export --path /tmp/hugo/test
```

### Build and Export

```bash
# Build and export static content
dagger call -m hugo build-and-export \
  --name blog \
  --config tests/hugo/hugo.toml \
  --content tests/hugo/content \
  export --path /tmp/blog/static
```

### Local Development

```bash
# Serve site locally with live reload
dagger call -m hugo serve \
  --config tests/hugo/hugo.toml \
  --content tests/hugo/content \
  --port 4144 \
  up --progress plain
```

### Test Module

```bash
# Run comprehensive tests
task test-hugo
```

## API Reference

### Site Initialization

```bash
dagger call -m hugo init-site \
  --name mysite \
  --config config/hugo.toml \
  --content content/ \
  export --path /tmp/hugo/
```

### Static Building

```bash
dagger call -m hugo build-and-export \
  --name mysite \
  --config config/hugo.toml \
  --content content/ \
  export --path /tmp/static/
```

### Development Server

```bash
dagger call -m hugo serve \
  --config config/hugo.toml \
  --content content/ \
  --port 1313 \
  up --progress plain
```

### MinIO Integration

```bash
# Sync assets from MinIO bucket
dagger call -m hugo sync-minio-bucket \
  --endpoint https://artifacts.automation.example.com \
  --bucket-name images \
  --insecure true \
  --access-key env:MINIO_USER \
  --secret-key env:MINIO_PASSWORD \
  --alias-name artifacts \
  export --path /tmp/images

# Build with MinIO asset sync
dagger call -m hugo build-sync-export \
  --name blog \
  --config config/hugo.toml \
  --content content/ \
  --endpoint https://artifacts.example.com \
  --bucket-name assets \
  --access-key env:MINIO_USER \
  --secret-key env:MINIO_PASSWORD \
  --alias-name artifacts \
  export --path /tmp/site/
```

## Serving Static Content

After building, you can serve the static content:

```bash
# Make content readable
chmod -R o+rX /tmp/blog/static

# Serve with nginx container
docker run --rm -p 8080:80 \
  -v "/tmp/blog/static:/usr/share/nginx/html" nginx
```

## Examples

See the [main README](../README.md#hugo) for detailed usage examples.

## Testing

```bash
task test-hugo
```

## Resources

- [Hugo Documentation](https://gohugo.io/documentation/)
- [Hugo Themes](https://themes.gohugo.io/)
- [MinIO Documentation](https://min.io/docs/)