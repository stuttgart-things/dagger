package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisConnection(t *testing.T) {
	ctx := context.Background()

	// Get Redis connection details from environment
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	if redisAddr == "" {
		t.Fatal("REDIS_ADDR environment variable not set")
	}
	if redisPort == "" {
		t.Fatal("REDIS_PORT environment variable not set")
	}
	if redisPassword == "" {
		t.Fatal("REDIS_PASSWORD environment variable not set")
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisAddr, redisPort),
		Password: redisPassword,
		DB:       0,
	})
	defer client.Close()

	// Test connection with retry logic
	var pingErr error
	for i := 0; i < 10; i++ {
		_, pingErr = client.Ping(ctx).Result()
		if pingErr == nil {
			break
		}
		t.Logf("Retry %d: waiting for Redis to be ready...", i+1)
		time.Sleep(1 * time.Second)
	}

	if pingErr != nil {
		t.Fatalf("Failed to connect to Redis: %v", pingErr)
	}

	t.Log("Successfully connected to Redis")
}

func TestRedisBasicOperations(t *testing.T) {
	ctx := context.Background()

	// Get Redis connection details from environment
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisAddr, redisPort),
		Password: redisPassword,
		DB:       0,
	})
	defer client.Close()

	// Wait for Redis to be ready
	for i := 0; i < 10; i++ {
		if _, err := client.Ping(ctx).Result(); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Test SET operation
	testKey := "test:key"
	testValue := "test-value"

	err := client.Set(ctx, testKey, testValue, 0).Err()
	if err != nil {
		t.Fatalf("Failed to SET key: %v", err)
	}
	t.Logf("Successfully SET key: %s", testKey)

	// Test GET operation
	val, err := client.Get(ctx, testKey).Result()
	if err != nil {
		t.Fatalf("Failed to GET key: %v", err)
	}
	if val != testValue {
		t.Fatalf("Expected value %s, got %s", testValue, val)
	}
	t.Logf("Successfully GET key: %s with value: %s", testKey, val)

	// Test DELETE operation
	err = client.Del(ctx, testKey).Err()
	if err != nil {
		t.Fatalf("Failed to DELETE key: %v", err)
	}
	t.Logf("Successfully DELETE key: %s", testKey)

	// Verify key is deleted
	_, err = client.Get(ctx, testKey).Result()
	if err != redis.Nil {
		t.Fatalf("Expected key to be deleted, but got: %v", err)
	}
	t.Log("Verified key was deleted")
}

func TestRedisStream(t *testing.T) {
	ctx := context.Background()

	// Get Redis connection details from environment
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisStream := os.Getenv("REDIS_STREAM")

	if redisStream == "" {
		redisStream = "messages" // default value
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisAddr, redisPort),
		Password: redisPassword,
		DB:       0,
	})
	defer client.Close()

	// Wait for Redis to be ready
	for i := 0; i < 10; i++ {
		if _, err := client.Ping(ctx).Result(); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Test XADD operation (add message to stream)
	streamID, err := client.XAdd(ctx, &redis.XAddArgs{
		Stream: redisStream,
		Values: map[string]interface{}{
			"message": "test message",
			"timestamp": time.Now().Unix(),
		},
	}).Result()
	if err != nil {
		t.Fatalf("Failed to add message to stream: %v", err)
	}
	t.Logf("Successfully added message to stream %s with ID: %s", redisStream, streamID)

	// Test XREAD operation (read messages from stream)
	messages, err := client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{redisStream, "0"},
		Count:   10,
		Block:   0,
	}).Result()
	if err != nil {
		t.Fatalf("Failed to read messages from stream: %v", err)
	}

	if len(messages) == 0 {
		t.Fatal("Expected at least one message in stream")
	}

	t.Logf("Successfully read %d message(s) from stream", len(messages[0].Messages))

	// Clean up: delete the stream
	err = client.Del(ctx, redisStream).Err()
	if err != nil {
		t.Fatalf("Failed to delete stream: %v", err)
	}
	t.Logf("Successfully cleaned up stream: %s", redisStream)
}
