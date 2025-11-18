# Homerun - Redis Service Module

A Dagger module for running Redis services in your CI/CD pipelines and local development.

## 🚀 TL;DR - Quick Start

```bash
# Start Redis locally with port forwarding (NO PASSWORD - insecure!)
dagger call -m homerun run-redis up

# Start Redis with password (RECOMMENDED)
dagger call -m homerun run-redis --password="mypass" up # pragma: allowlist secret

# Open interactive redis-cli shell
dagger call -m homerun redis-cli terminal
dagger call -m homerun redis-cli --password="mypass" terminal # pragma: allowlist secret

# Quick test Redis is working
dagger call -m homerun test-redis-connection

# Run your tests with Redis
dagger call -m homerun run-test-with-redis --source=./tests --test-path=.

# Generate a secure password
dagger call -m homerun generate-password --length=32
```

## 📋 Quick Reference

| Command | Description |
|---------|-------------|
| `dagger call -m homerun run-redis up` | **Start Redis without password (insecure, for dev only)** |
| `dagger call -m homerun run-redis --password="pass" up` | **Start Redis with password (recommended)** | #pragma: allowlist secret
| `dagger call -m homerun redis-cli terminal` | **Open interactive redis-cli shell** |
| `dagger call -m homerun redis-cli --password="pass" terminal` | **Open redis-cli with password** | # pragma: allowlist secret
| `dagger call -m homerun test-redis-connection` | Quick test (no password) |
| `dagger call -m homerun test-redis-connection --password="pass"` | Test with password | # pragma: allowlist secret
| `dagger call -m homerun redis-service` | Get Redis service object |
| `dagger call -m homerun generate-password` | Generate secure password |
| `dagger call -m homerun run-test-with-redis --source=./tests --test-path=.` | Full test runner |

---

## 📖 Table of Contents

<details>
<summary><b>Getting Started</b></summary>

### The Three Ways to Use Redis

#### 1. 🚀 Port Forwarding (`dagger up`)

**Best for:** Local development, debugging, manual testing

Start Redis and access it from your host machine:

```bash
# Start Redis with password
dagger call -m homerun run-redis --password="dev"  # pragma: allowlist secret

# In another terminal, connect from your machine
redis-cli -p 6379 -a dev

# Or without password (insecure, dev only)
dagger call -m homerun run-redis up
redis-cli -p 6379
```

This is exactly like running:
```bash
docker run -p 6379:6379 redis/redis-stack-server
```

But with Dagger's reproducibility and pipeline integration!

#### 2. 🔗 Service Binding (In Modules)

**Best for:** CI/CD, automated testing, module composition

Use Redis as a service in your own Dagger modules:

```go
// In your module
redis := dag.Homerun().RedisService("7.2.0-v18", "testpass")

result := dag.Container().
    From("alpine").
    WithServiceBinding("redis", redis).  // ← Redis accessible at hostname "redis"
    WithExec([]string{"redis-cli", "-h", "redis", "-a", "testpass", "ping"}).
    Stdout(ctx)
```

#### 3. 🧪 Test Runner (All-in-One)

**Best for:** Go projects with Redis integration tests

Complete test environment with Redis:

```bash
dagger call -m homerun run-test-with-redis \
  --source=./your-project \
  --test-path=./tests
```

This automatically:
- ✅ Starts Redis with random password
- ✅ Sets up Go environment
- ✅ Configures environment variables
- ✅ Runs your tests
- ✅ Cleans up everything

### Understanding Service Binding vs Port Forwarding

**With `up` (Port Forwarding)**
```
Your Host Machine ←→ localhost:6379 ←→ Dagger Redis Service
     ✅ Can connect from host with redis-cli
     ✅ Can connect from your local apps
     ⚠️  Port exposed on your machine
```

**With Service Binding (In Modules)**
```
Dagger Container ←→ hostname:redis:6379 ←→ Dagger Redis Service
     ✅ Isolated network within Dagger
     ✅ No ports exposed on host
     ✅ Perfect for CI/CD
     ⚠️  Only accessible from Dagger containers
```

</details>

<details>
<summary><b>Available Functions</b></summary>

### `run-redis`
Start a Redis service with port forwarding (use with `up`)

**Parameters:**
- `version` (optional, default: "7.2.0-v18") - Redis version
- `port` (optional, default: 6379) - Port to expose
- `password` (optional) - Redis password (⚠️ **strongly recommended**)

**Examples:**
```bash
# With password (recommended)
dagger call -m homerun run-redis --password="test" up # pragma: allowlist secret

# Without password (insecure, dev only)
dagger call -m homerun run-redis up
```

**⚠️ Security Note:** When no password is provided, Redis runs without authentication. This is **only suitable for local development**. Always use a password in any shared or production-like environment.

### `redis-cli`
Open an interactive redis-cli shell connected to Redis (use with `terminal`)

**Parameters:**
- `version` (optional, default: "7.2.0-v18") - Redis version
- `password` (optional) - Redis password

**Examples:**
```bash
# Without password
dagger call -m homerun redis-cli terminal

# With password
dagger call -m homerun redis-cli --password="mypass" terminal # pragma: allowlist secret
```

**What you can do:**
Once in the redis-cli shell, you can run any Redis commands:
```
127.0.0.1:6379> SET mykey "Hello World"
127.0.0.1:6379> GET mykey
127.0.0.1:6379> HSET user:1 name "John" age "30"
127.0.0.1:6379> HGETALL user:1
127.0.0.1:6379> KEYS *
127.0.0.1:6379> exit
```

### `redis-service`
Get a Redis service for use in other Dagger modules

**Parameters:**
- `version` (optional, default: "7.2.0-v18") - Redis version
- `password` (optional) - Redis password

**Example in Go module:**
```go
redis := dag.Homerun().RedisService("7.2.0-v18", "mypass")

result := dag.Container().
    From("alpine").
    WithServiceBinding("redis", redis).
    WithEnvVariable("REDIS_ADDR", "redis:6379").
    // ... your commands
```

### `test-redis-connection`
Test if Redis is reachable and responding

**Parameters:**
- `version` (optional, default: "7.2.0-v18") - Redis version
- `password` (optional) - Redis password

**Example:**
```bash
dagger call -m homerun test-redis-connection --password="test" # pragma: allowlist secret
```

**Output:** Should return `PONG` if Redis is working correctly.

### `generate-password`
Generate a secure random password

**Parameters:**
- `length` (optional, default: 16) - Password length

**Example:**
```bash
dagger call -m homerun generate-password --length=32
```

### `run-test-with-redis`
Run Go tests with Redis service automatically started

**Parameters:**
- `source` (required) - Directory containing your Go code
- `test-path` (required) - Path to test (e.g., "." for current dir)
- `go-version` (optional, default: "1.25.4") - Go version
- `redis-version` (optional, default: "7.2.0-v18") - Redis version

**Example:**
```bash
dagger call -m homerun run-test-with-redis \
  --source=./tests/homerun \
  --test-path=. \
  --go-version=1.25.4 \
  --redis-version=7.2.0-v18
```

</details>

<details>
<summary><b>Common Use Cases & Examples</b></summary>

### For Local Development

```bash
# Start Redis for your local app
dagger call -m homerun run-redis --password="dev123" up # pragma: allowlist secret

# Your app can now connect to localhost:6379
```

### For CI/CD Pipelines

**GitHub Actions:**
```yaml
- name: Run integration tests
  run: |
    dagger call -m homerun run-test-with-redis \
      --source=. \
      --test-path=./tests/integration
```

**GitLab CI:**
```yaml
test:
  image: registry.dagger.io/engine:latest
  script:
    - dagger call -m homerun run-test-with-redis --source=. --test-path=./tests
```

### For Multi-Module Dagger Projects

```go
// mymodule/main.go
package main

import (
    "context"
    "dagger/mymodule/internal/dagger"
)

type MyModule struct{}

func (m *MyModule) Test(ctx context.Context, source *dagger.Directory) (string, error) {
    // Get Redis from homerun module
    redis := dag.Homerun().RedisService("7.2.0-v18", "testpass")

    // Run your tests with Redis
    return dag.Container().
        From("golang:1.25-alpine").
        WithMountedDirectory("/src", source).
        WithWorkdir("/src").
        WithServiceBinding("redis", redis).
        WithEnvVariable("REDIS_ADDR", "redis:6379").
        WithEnvVariable("REDIS_PASSWORD", "testpass").
        WithExec([]string{"go", "test", "-v", "./..."}).
        Stdout(ctx)
}
```

### For Security Testing

```bash
# Generate secure password
PASSWORD=$(dagger call -m homerun generate-password --length=32)

# Use it
dagger call -m homerun run-redis --password="$PASSWORD" up
```

### For Multiple Redis Versions

Test compatibility across different Redis versions:

```bash
for version in "6.2.0-v18" "7.0.0-v18" "7.2.0-v18"; do
  echo "Testing Redis $version..."
  dagger call -m homerun test-redis-connection # pragma: allowlist secret --version="$version" --password="test"
done
```

</details>

<details>
<summary><b>Common Patterns</b></summary>

### Pattern 1: Quick Local Testing

```bash
# Start Redis in background
dagger call -m homerun run-redis --password="test" up & # pragma: allowlist secret

# Run your app
REDIS_ADDR=localhost:6379 REDIS_PASSWORD=test go run main.go

# Stop Redis
fg  # then Ctrl+C
```

### Pattern 2: CI Integration Test

```bash
# Your test expects these env vars:
# REDIS_ADDR, REDIS_PORT, REDIS_PASSWORD
dagger call -m homerun run-test-with-redis \
  --source=. \
  --test-path=./tests
```

### Pattern 3: Custom Test Pipeline with Dynamic Password

```go
func (m *MyModule) IntegrationTest(ctx context.Context, src *dagger.Directory) error {
    // Generate password
    pass, _ := dag.Homerun().GeneratePassword(16)

    // Start Redis
    redis := dag.Homerun().RedisService("7.2.0-v18", pass)

    // Run tests
    _, err := dag.Container().
        From("golang:1.25").
        WithMountedDirectory("/src", src).
        WithServiceBinding("redis", redis).
        WithEnvVariable("REDIS_PASSWORD", pass).
        WithExec([]string{"go", "test", "./..."}).
        Sync(ctx)

    return err
}
```

### Pattern 4: Version Compatibility Testing

```bash
#!/bin/bash
for version in "6.2.0-v18" "7.0.0-v18" "7.2.0-v18"; do
    echo "Testing Redis $version"
    dagger call -m homerun run-test-with-redis \
        --source=. \
        --test-path=./tests \
        --redis-version="$version" || exit 1
done
echo "All versions passed!"
```

### Pattern 5: Interactive Shell with Redis

```bash
dagger call -m homerun redis-service --password="testpass" \ # pragma: allowlist secret
  | dagger call container \
    --from alpine:latest \
    with-service-binding --alias redis --service - \
    with-exec --args "sh,-c,apk add --no-cache redis bash" \
    terminal
```

Then in the shell:
```bash
redis-cli -h redis -a testpass
> SET mykey "Hello"
> GET mykey
> KEYS *
```

</details>

<details>
<summary><b>CLI Examples</b></summary>

### Using `dagger up` (Port Forwarding to Host)

```bash
# Start Redis on default port (6379) without password
dagger call -m homerun run-redis up

# Start Redis with a password
dagger call -m homerun run-redis --password="mySecurePass" up # pragma: allowlist secret

# Start Redis on a custom port
dagger call -m homerun run-redis --port=6380 --password="test123" up # pragma: allowlist secret

# Start Redis with specific version
dagger call -m homerun run-redis --version="7.0.0-v18" --password="pass" up # pragma: allowlist secret
```

**What this does:**
- ✅ Starts Redis service in Dagger
- ✅ Forwards the port to your local machine
- ✅ You can connect from your host with: `redis-cli -p 6379 -a mySecurePass`
- ✅ Keeps running until you press Ctrl+C

**Example connecting from your local machine:**
```bash
# Terminal 1: Start Redis
dagger call -m homerun run-redis --password="test" up # pragma: allowlist secret

# Terminal 2: Connect to it
redis-cli -p 6379 -a test
> SET mykey "hello"
> GET mykey
```

### Test Redis Connection

```bash
# Test with default version (7.2.0-v18), no password
dagger call -m homerun test-redis-connection

# Test with password
dagger call -m homerun test-redis-connection --password="mySecretPass" # pragma: allowlist secret

# Test with specific version
dagger call -m homerun test-redis-connection --version="7.0.0-v18" --password="test123" # pragma: allowlist secret
```

### Use Redis Service in Pipelines

```bash
# Example: Start Redis and connect to it
dagger call -m homerun redis-service \
  | dagger call container \
    --from alpine:latest \
    with-service-binding --alias redis --service - \
    with-exec --args "sh,-c,apk add redis && redis-cli -h redis ping" \
    stdout
```

</details>

<details>
<summary><b>Module Integration Examples</b></summary>

### Example: CI Module with Redis

Create a file `ci/main.go`:

```go
package main

import (
    "context"
    "dagger/ci/internal/dagger"
)

type Ci struct{}

// RunIntegrationTests runs tests with Redis cache
func (m *Ci) RunIntegrationTests(
    ctx context.Context,
    source *dagger.Directory,
) (string, error) {
    // Get Redis service from homerun module with a secure password
    password := "test-password-123" // In production, generate this
    redis := dag.Homerun().RedisService("7.2.0-v18", password)

    // Run your tests with Redis available
    return dag.Container().
        From("golang:1.25-alpine").
        WithMountedDirectory("/src", source).
        WithWorkdir("/src").
        WithServiceBinding("redis", redis).
        WithEnvVariable("REDIS_ADDR", "redis:6379").
        WithEnvVariable("REDIS_PASSWORD", password).
        WithExec([]string{"go", "test", "-v", "./..."}).
        Stdout(ctx)
}

// RunWithGeneratedPassword shows how to generate password on-the-fly
func (m *Ci) RunWithGeneratedPassword(
    ctx context.Context,
    source *dagger.Directory,
) (string, error) {
    // Generate a secure random password
    password, err := dag.Homerun().GeneratePassword(24)
    if err != nil {
        return "", err
    }

    // Use Redis with generated password
    redis := dag.Homerun().RedisService("7.2.0-v18", password)

    return dag.Container().
        From("golang:1.25-alpine").
        WithMountedDirectory("/src", source).
        WithWorkdir("/src").
        WithServiceBinding("redis", redis).
        WithEnvVariable("REDIS_ADDR", "redis:6379").
        WithEnvVariable("REDIS_PASSWORD", password).
        WithExec([]string{"go", "test", "-v", "./..."}).
        Stdout(ctx)
}
```

Then call it:

```bash
# Initialize your CI module
dagger init --sdk=go --source=./ci --name=ci

# Test Redis connection
dagger call -m ci start-redis-test

# Run tests with Redis
dagger call -m ci run-tests-with-redis --source=./your-project

# Test with generated password
dagger call -m ci generate-and-use-password
```

### Example: Multiple Redis Versions for Compatibility Testing

```go
func (m *Ci) TestRedisCompatibility(ctx context.Context, source *dagger.Directory) error {
    versions := []string{"6.2.0-v18", "7.0.0-v18", "7.2.0-v18"}

    for _, version := range versions {
        redis := dag.Homerun().RedisService(version, "testpass")

        output, err := dag.Container().
            From("golang:1.25-alpine").
            WithMountedDirectory("/src", source).
            WithWorkdir("/src").
            WithServiceBinding("redis", redis).
            WithEnvVariable("REDIS_ADDR", "redis:6379").
            WithEnvVariable("REDIS_PASSWORD", "testpass").
            WithExec([]string{"go", "test", "-v", "./..."}).
            Stdout(ctx)

        if err != nil {
            return fmt.Errorf("tests failed for Redis %s: %w", version, err)
        }

        fmt.Printf("✓ Tests passed for Redis %s\n", version)
    }

    return nil
}
```

</details>

<details>
<summary><b>Troubleshooting</b></summary>

### "Connection refused"
```bash
# Redis might still be starting, add retry logic or wait a bit
dagger call -m homerun test-redis-connection --password="yourpass" # pragma: allowlist secret
```

The service might still be starting. The test includes retry logic automatically.

### "NOAUTH Authentication required"
```bash
# You set a password but forgot to pass it
# Add --password parameter
dagger call -m homerun run-redis --password="yourpass" up # pragma: allowlist secret
```

You're trying to connect without a password when one is set. Pass the password parameter.

### "Wrong password"
Double-check the password being passed to the service and the client.

### Port already in use
```bash
# Use different port
dagger call -m homerun run-redis --port=6380 up

# Then connect with:
redis-cli -p 6380
```

</details>

<details>
<summary><b>Best Practices</b></summary>

1. **Always use passwords in production** - Even though optional, always set passwords for security
2. **Generate passwords dynamically** - Use `generate-password` for secure random passwords
3. **Version pinning** - Specify exact Redis versions for reproducible builds
4. **Clean up** - Dagger handles this automatically, but be aware services persist during pipeline execution
5. **Use service binding in CI** - Don't expose ports in CI/CD, use service binding instead
6. **Test locally with `up`** - Use port forwarding for quick local testing and debugging

</details>

<details>
<summary><b>Comparison with Other Tools</b></summary>

| Tool | Command | Homerun Equivalent |
|------|---------|-------------------|
| Docker | `docker run -p 6379:6379 redis` | `dagger call -m homerun run-redis up` |
| Docker Compose | Service definition in YAML | `dag.Homerun().RedisService()` in code |
| Testcontainers | Go code with containers | `RunTestWithRedis` function |

**Advantages of Homerun:**
- ✅ Reproducible across all environments
- ✅ No Docker daemon required (Dagger handles it)
- ✅ Works same on CI and locally
- ✅ Version controlled configuration
- ✅ Composable with other Dagger modules

</details>

---

## ✨ Key Features

✅ **Port forwarding with `dagger up`** - Just like hugo serve
✅ **Standalone Redis service** - Use in any Dagger module
✅ **Optional authentication** - With or without password
✅ **Version control** - Test against multiple Redis versions
✅ **Password generation** - Built-in secure password generator
✅ **Ready-to-use test runner** - For Go projects
✅ **CI/CD friendly** - Easy integration with any pipeline

## 📦 Requirements

- Dagger installed
- For local `dagger up` usage: redis-cli (optional, for testing connection)

## 🧪 Test Suite

See the [tests/homerun](../../tests/homerun) directory for complete working examples with:
- Go test code
- Redis connection tests
- Basic operations tests
- Redis Streams tests

## 🔧 Task Integration

<details>
<summary><b>Taskfile Commands - Quick Reference</b></summary>

### Available Tasks

#### `task test-homerun`
Run the full homerun test suite with Redis

**Usage:**
```bash
task test-homerun
```

**What it does:**
- Starts Redis service automatically
- Runs Go tests from `tests/homerun/`
- Tests Redis connection, basic operations, and streams
- Uses Go 1.25.4 and Redis 7.2.0-v18 by default

---

#### `task run-redis`
Start Redis with port forwarding to your host machine

**Usage:**
```bash
# Start without password (insecure, dev only)
task run-redis

# Start with password (recommended)
task run-redis PASSWORD=mySecurePass

# Custom port and password
task run-redis PORT=6380 PASSWORD=secret

# Specific Redis version
task run-redis PASSWORD=test REDIS_VERSION=7.0.0-v18
```

**Variables:**
- `PORT` - Port to expose (default: 6379)
- `PASSWORD` - Redis password (default: none/empty)

**What it does:**
- Starts Redis using `dagger up`
- Forwards Redis port to your localhost
- Keeps running until you press Ctrl+C
- Shows connection info and password

---

#### `task redis-cli`
Connect to Redis using redis-cli

**Usage:**
```bash
# Connect to localhost:6379 without password
task redis-cli

# Connect with password
task redis-cli PASSWORD=mypass

# Connect to remote host
task redis-cli HOST=maverick.example.com PORT=6379 PASSWORD=hello

# Connect to custom port
task redis-cli HOST=localhost PORT=6380 PASSWORD=secret
```

**Variables:**
- `HOST` - Redis host (default: localhost)
- `PORT` - Redis port (default: 6379)
- `PASSWORD` - Redis password (default: none/empty)

**What it does:**
- Opens interactive redis-cli session
- Automatically includes `-a` flag if PASSWORD is set
- Connects to specified host and port

---

### Common Workflows

#### Workflow 1: Local Development

**Terminal 1 - Start Redis:**
```bash
task run-redis PASSWORD=dev123
```

**Terminal 2 - Connect and use:**
```bash
task redis-cli PASSWORD=dev123

# In redis-cli:
127.0.0.1:6379> SET user:1 "John Doe"
127.0.0.1:6379> GET user:1
127.0.0.1:6379> HSET session:abc token "xyz123"
127.0.0.1:6379> HGETALL session:abc
127.0.0.1:6379> exit
```

**Terminal 3 - Run your app:**
```bash
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=dev123
go run main.go
```

---

#### Workflow 2: Testing

**Run the full test suite:**
```bash
task test-homerun
```

**Or test manually:**
```bash
# Terminal 1: Start Redis
task run-redis PASSWORD=test

# Terminal 2: Run your tests
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=test
go test -v ./...
```

---

#### Workflow 3: Remote Connection

Connect to a Redis instance running elsewhere:

```bash
task redis-cli HOST=maverick.tiab.labda.sva.de PORT=6379 PASSWORD=hello
```

---

### Tips

1. **Always use passwords** except for quick local testing
2. **Use environment variables** for sensitive passwords:
   ```bash
   task run-redis PASSWORD=$REDIS_PASSWORD
   task redis-cli PASSWORD=$REDIS_PASSWORD
   ```

3. **Check if Redis is running** before connecting:
   ```bash
   redis-cli -h localhost -p 6379 -a mypass ping
   ```

4. **Stop Redis** gracefully:
   - Press `Ctrl+C` in the terminal where `task run-redis` is running

5. **Multiple Redis instances**:
   ```bash
   # Terminal 1: Redis on port 6379
   task run-redis PORT=6379 PASSWORD=redis1

   # Terminal 2: Redis on port 6380
   task run-redis PORT=6380 PASSWORD=redis2

   # Terminal 3: Connect to first instance
   task redis-cli PORT=6379 PASSWORD=redis1

   # Terminal 4: Connect to second instance
   task redis-cli PORT=6380 PASSWORD=redis2
   ```

---

### Troubleshooting Task Commands

#### Can't connect with redis-cli

**Problem:** `Could not connect to Redis`

**Solutions:**
```bash
# Check if Redis is running
task redis-cli

# If it fails, start Redis first:
task run-redis PASSWORD=test

# Then try again:
task redis-cli PASSWORD=test
```

#### Wrong password error

**Problem:** `NOAUTH Authentication required` or `ERR invalid password`

**Solution:**
```bash
# Make sure passwords match:
task run-redis PASSWORD=correctpass
task redis-cli PASSWORD=correctpass  # Must be the same!
```

#### Port already in use

**Problem:** `address already in use`

**Solution:**
```bash
# Use a different port
task run-redis PORT=6380 PASSWORD=test
task redis-cli PORT=6380 PASSWORD=test
```

#### Can't find redis-cli command

**Problem:** `redis-cli: command not found`

**Solution:**
Install redis-cli:
```bash
# macOS
brew install redis

# Ubuntu/Debian
sudo apt-get install redis-tools

# Fedora/RHEL
sudo dnf install redis
```

---

### Quick Task Command Reference

| Task Command | What It Does |
|-------------|--------------|
| `task test-homerun` | Run full test suite |
| `task run-redis` | Start Redis (no password) |
| `task run-redis PASSWORD=pass` | Start Redis (with password) |
| `task redis-cli` | Connect to localhost:6379 |
| `task redis-cli PASSWORD=pass` | Connect with password |
| `task redis-cli HOST=remote` | Connect to remote Redis |

</details>

## 📚 Service Binding Concepts

When you use `redis-service`, it returns a Dagger Service. This service:
- ✅ Runs in the background
- ✅ Is automatically networked to containers that bind it
- ✅ Is accessible via hostname (e.g., "redis")
- ✅ Shuts down automatically when no longer needed
- ✅ Can be shared across multiple containers in a pipeline

The service is NOT exposed to your host machine - it only exists within the Dagger pipeline execution context.

## 🤝 Contributing

Feel free to open issues or submit pull requests!

## 📝 License

See the main repository LICENSE file.
