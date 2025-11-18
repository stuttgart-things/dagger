# Homerun Redis Tests

This directory contains tests for the `RunTestWithRedis` Dagger function.

## Test Files

- **main.go** - Standalone test program that can be run with `go run` to test Redis connectivity and basic operations
- **redis_test.go** - Go test suite with comprehensive Redis tests
- **go.mod** - Go module definition with Redis client dependency

## Running Tests with Dagger

From the homerun module directory, you can run the tests using the `RunTestWithRedis` function:

```bash
# Navigate to the homerun module
cd /home/sthings/projects/dagger/homerun

# Run the test with default versions
dagger call run-test-with-redis --source=/home/sthings/projects/dagger/tests/homerun --test-path=.

# Run with specific Go and Redis versions
dagger call run-test-with-redis \
  --source=/home/sthings/projects/dagger/tests/homerun \
  --test-path=. \
  --go-version=1.25.4 \
  --redis-version=7.2.0-v18
```

## What the Tests Cover

### Connection Test
- Validates Redis connectivity with environment variables
- Tests authentication with password
- Includes retry logic for connection stability

### Basic Operations Test
- SET operation
- GET operation
- DELETE operation
- Verification of deleted keys

### Redis Streams Test
- XADD - Adding messages to a stream
- XREAD - Reading messages from a stream
- Stream cleanup

## Environment Variables

The tests expect the following environment variables (automatically set by the Dagger function):

- `REDIS_ADDR` - Redis host address
- `REDIS_PORT` - Redis port (default: 6379)
- `REDIS_PASSWORD` - Redis authentication password
- `REDIS_STREAM` - Redis stream name (default: messages)

## Local Testing

To run the tests locally without Dagger, you need a Redis instance:

```bash
# Start Redis with Docker
docker run -d -p 6379:6379 --name redis-test redis/redis-stack-server:7.2.0-v18

# Set environment variables
export REDIS_ADDR=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=yourpassword
export REDIS_STREAM=messages

# Run the standalone test
go run main.go

# Or run the test suite
go test -v
```
