package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	// Get Redis connection details from environment
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisStream := os.Getenv("REDIS_STREAM")

	if redisAddr == "" {
		fmt.Println("ERROR: REDIS_ADDR environment variable not set")
		os.Exit(1)
	}
	if redisPort == "" {
		fmt.Println("ERROR: REDIS_PORT environment variable not set")
		os.Exit(1)
	}
	if redisPassword == "" {
		fmt.Println("ERROR: REDIS_PASSWORD environment variable not set")
		os.Exit(1)
	}
	if redisStream == "" {
		redisStream = "messages" // default value
	}

	fmt.Printf("Connecting to Redis at %s:%s...\n", redisAddr, redisPort)

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisAddr, redisPort),
		Password: redisPassword,
		DB:       0,
	})
	defer client.Close()

	// Test connection with retry logic
	fmt.Println("Testing Redis connection...")
	var pingErr error
	for i := 0; i < 10; i++ {
		_, pingErr = client.Ping(ctx).Result()
		if pingErr == nil {
			break
		}
		fmt.Printf("Retry %d: waiting for Redis to be ready...\n", i+1)
		time.Sleep(1 * time.Second)
	}

	if pingErr != nil {
		fmt.Printf("ERROR: Failed to connect to Redis: %v\n", pingErr)
		os.Exit(1)
	}

	fmt.Println("✓ Successfully connected to Redis")

	// Test basic operations
	fmt.Println("\nTesting basic SET/GET operations...")
	testKey := "test:key"
	testValue := "test-value"

	err := client.Set(ctx, testKey, testValue, 0).Err()
	if err != nil {
		fmt.Printf("ERROR: Failed to SET key: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Successfully SET key: %s\n", testKey)

	val, err := client.Get(ctx, testKey).Result()
	if err != nil {
		fmt.Printf("ERROR: Failed to GET key: %v\n", err)
		os.Exit(1)
	}
	if val != testValue {
		fmt.Printf("ERROR: Expected value %s, got %s\n", testValue, val)
		os.Exit(1)
	}
	fmt.Printf("✓ Successfully GET key: %s with value: %s\n", testKey, val)

	err = client.Del(ctx, testKey).Err()
	if err != nil {
		fmt.Printf("ERROR: Failed to DELETE key: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Successfully DELETE key: %s\n", testKey)

	// Test Redis Streams
	fmt.Printf("\nTesting Redis Streams with stream name: %s...\n", redisStream)
	streamID, err := client.XAdd(ctx, &redis.XAddArgs{
		Stream: redisStream,
		Values: map[string]interface{}{
			"message":   "Hello from Dagger test!",
			"timestamp": time.Now().Unix(),
			"test_run":  "homerun",
		},
	}).Result()
	if err != nil {
		fmt.Printf("ERROR: Failed to add message to stream: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Successfully added message to stream with ID: %s\n", streamID)

	messages, err := client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{redisStream, "0"},
		Count:   10,
		Block:   0,
	}).Result()
	if err != nil {
		fmt.Printf("ERROR: Failed to read messages from stream: %v\n", err)
		os.Exit(1)
	}

	if len(messages) == 0 {
		fmt.Println("ERROR: Expected at least one message in stream")
		os.Exit(1)
	}

	fmt.Printf("✓ Successfully read %d message(s) from stream\n", len(messages[0].Messages))
	for i, msg := range messages[0].Messages {
		fmt.Printf("  Message %d (ID: %s): %v\n", i+1, msg.ID, msg.Values)
	}

	// Clean up
	fmt.Println("\nCleaning up...")
	err = client.Del(ctx, redisStream).Err()
	if err != nil {
		fmt.Printf("WARNING: Failed to delete stream: %v\n", err)
	} else {
		fmt.Printf("✓ Successfully cleaned up stream: %s\n", redisStream)
	}

	fmt.Println("\n✅ All tests passed successfully!")
}
