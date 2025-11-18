// A generated module for Homerun functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

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
		From("redis:alpine").
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
		From("alpine:latest").
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
		WithExec([]string{"go", "run", testPath}).
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
