// Homerun provides Redis-backed test scaffolding for other Dagger modules.
//
// It exposes a Redis service (optionally password-protected), a randomly
// generated password helper, an interactive redis-cli container, a one-shot
// connectivity probe, and a `RunTestWithRedis` runner that wires a Redis
// service into a Go test container — all without requiring an external Redis.

package main

import (
	"context"
	"crypto/rand"
	"dagger/homerun/internal/dagger"
	"encoding/base64"
	"fmt"
)

type Homerun struct{}

// RedisService creates a Redis service with optional password protection
//
// Returns a Redis service that can be bound to other containers for testing
func (m *Homerun) RedisService(
	// +optional
	// +default="7.2.0-v18"
	version string,
	// +optional
	password string,
) *dagger.Service {
	container := dag.Container().
		From("redis/redis-stack-server:" + version).
		WithExposedPort(6379)

	// Only set password if provided
	if password != "" {
		container = container.WithEnvVariable("REDIS_ARGS", "--requirepass "+password)
	}

	return container.AsService()
}

// GeneratePassword generates a random password of specified length
//
// Useful for creating secure passwords for Redis or other services
func (m *Homerun) GeneratePassword(
	// +optional
	// +default=16
	length int,
) (string, error) {
	return randomPassword(length)
}

// RunRedis starts a Redis service that can be accessed with 'dagger up'
//
// Example: dagger call -m homerun run-redis --port 6379 --password "mypass" up
func (m *Homerun) RunRedis(
	// +optional
	// +default="7.2.0-v18"
	version string,
	// +optional
	// +default=6379
	port int,
	// +optional
	password string,
) *dagger.Service {
	container := dag.Container().
		From("redis/redis-stack-server:" + version).
		WithExposedPort(port)

	// Only set password if provided
	if password != "" {
		container = container.WithEnvVariable("REDIS_ARGS", "--requirepass "+password)
	}

	return container.AsService()
}

// RedisCli opens an interactive redis-cli shell connected to a Redis service
//
// This is useful for debugging and manual interaction with Redis
func (m *Homerun) RedisCli(
	// +optional
	// +default="7.2.0-v18"
	version string,
	// +optional
	password string,
) *dagger.Container {
	redis := m.RedisService(version, password)

	// Build the redis-cli command based on whether password is set
	var args []string
	if password != "" {
		args = []string{"redis-cli", "-h", "redis", "-a", password}
	} else {
		args = []string{"redis-cli", "-h", "redis"}
	}

	return dag.Container().
		From("redis:7.4-alpine").
		WithServiceBinding("redis", redis).
		WithEntrypoint(args).
		Terminal()
}

// TestRedisConnection tests if Redis service is reachable and working
//
// This is useful to verify the Redis service is up and accessible from the CLI
func (m *Homerun) TestRedisConnection(
	ctx context.Context,
	// +optional
	// +default="7.2.0-v18"
	version string,
	// +optional
	password string,
) (string, error) {
	redis := m.RedisService(version, password)

	// Build the redis-cli command based on whether password is set
	var cmd []string
	if password != "" {
		cmd = []string{"sh", "-c", "apk add --no-cache redis && redis-cli -h redis -a " + password + " ping"}
	} else {
		cmd = []string{"sh", "-c", "apk add --no-cache redis && redis-cli -h redis ping"}
	}

	return dag.Container().
		From("alpine:3.21").
		WithServiceBinding("redis", redis).
		WithExec(cmd).
		Stdout(ctx)
}

func (m *Homerun) RunTestWithRedis(
	ctx context.Context,
	source *dagger.Directory,
	// +optional
	// +default="1.25.4"
	goVersion string,
	// +optional
	// +default="7.2.0-v18"
	redisVersion string,
	testPath string,
) (string, error) {
	// generate random redis password
	generatedRedisPassword, err := randomPassword(16)
	if err != nil {
		return "", fmt.Errorf("failed to generate redis password: %w", err)
	}

	// START REDIS SERVICE IN BACKGROUND
	redis := m.RedisService(redisVersion, generatedRedisPassword)

	// RUN TEST CONTAINER
	return dag.Container().
		From("golang:"+goVersion+"-alpine").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("gomod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("gobuild")).
		WithServiceBinding("redis", redis).
		WithEnvVariable("REDIS_ADDR", "redis").
		WithEnvVariable("REDIS_PORT", "6379").
		WithEnvVariable("REDIS_STREAM", "messages").
		WithEnvVariable("REDIS_PASSWORD", generatedRedisPassword).
		WithExec([]string{"go", "mod", "download"}).
		WithExec([]string{"go", "test", "-v", testPath}).
		Stdout(ctx)
}

// HELPER: GENERATE A RANDOM PASSWORD
func randomPassword(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
