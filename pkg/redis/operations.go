// Package redis provides Redis client functionality with support for both
// task queue operations (via asynq) and key-value operations (via go-redis).
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Example usage:
//
//	type User struct {
//	    ID   string `json:"id"`
//	    Name string `json:"name"`
//	}
//
//	func ExampleUsage() {
//	    // Initialize client
//	    client, err := NewClient(cfg)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer client.Close()
//
//	    ctx := context.Background()
//
//	    // Store a user with 1 hour expiration
//	    user := User{ID: "123", Name: "John Doe"}
//	    if err := client.Set(ctx, "user:123", user, time.Hour); err != nil {
//	        log.Printf("Failed to store user: %v", err)
//	    }
//
//	    // Retrieve the user
//	    var retrievedUser User
//	    if err := client.Get(ctx, "user:123", &retrievedUser); err != nil {
//	        log.Printf("Failed to get user: %v", err)
//	    }
//
//	    // Update user with TTL preservation
//	    retrievedUser.Name = "Jane Doe"
//	    if err := client.Update(ctx, "user:123", retrievedUser); err != nil {
//	        log.Printf("Failed to update user: %v", err)
//	    }
//
//	    // Get user and TTL in a single operation
//	    var userWithTTL User
//	    ttl, err := client.GetWithTTL(ctx, "user:123", &userWithTTL)
//	    if err != nil {
//	        log.Printf("Failed to get user with TTL: %v", err)
//	    }
//	    log.Printf("User %s expires in %v", userWithTTL.Name, ttl)
//	}

// Set stores a value with an optional expiration. The value must be JSON-serializable.
// If expiration is 0, the key will not expire.
//
// Example:
//
//	user := User{ID: "123", Name: "John Doe"}
//	err := client.Set(ctx, "user:123", user, time.Hour)
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.redisClient.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}

	return nil
}

// Get retrieves a value by key and unmarshals it into the provided interface.
// Returns an error if the key doesn't exist or if unmarshaling fails.
//
// Example:
//
//	var user User
//	err := client.Get(ctx, "user:123", &user)
func (c *Client) Get(ctx context.Context, key string, value interface{}) error {
	data, err := c.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get value: %w", err)
	}

	if err := json.Unmarshal(data, value); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete removes a value by key. Returns nil if the key was deleted or didn't exist.
//
// Example:
//
//	err := client.Delete(ctx, "user:123")
func (c *Client) Delete(ctx context.Context, key string) error {
	if err := c.redisClient.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}

	return nil
}

// Update updates an existing value by key while preserving its current TTL.
// Returns an error if the key doesn't exist.
//
// Example:
//
//	user.Name = "Jane Doe"
//	err := client.Update(ctx, "user:123", user)
func (c *Client) Update(ctx context.Context, key string, value interface{}) error {
	// Check if key exists
	exists, err := c.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to check key existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("key not found: %s", key)
	}

	// Get current TTL
	ttl, err := c.redisClient.TTL(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get TTL: %w", err)
	}

	// Update value with existing TTL
	return c.Set(ctx, key, value, ttl)
}

// SetWithExpiration sets a value with expiration only if the key doesn't exist.
// Returns true if the key was set, false if it already existed.
//
// Example:
//
//	user := User{ID: "123", Name: "John Doe"}
//	ok, err := client.SetWithExpiration(ctx, "user:123", user, time.Hour)
//	if err == nil && ok {
//	    fmt.Println("User was set successfully")
//	}
func (c *Client) SetWithExpiration(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	// Use SetNX to only set if key doesn't exist
	ok, err := c.redisClient.SetNX(ctx, key, data, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set value with expiration: %w", err)
	}

	return ok, nil
}

// GetTTL returns the remaining time-to-live of a key.
// Returns -2 if the key doesn't exist, -1 if the key has no expiration.
//
// Example:
//
//	ttl, err := client.GetTTL(ctx, "user:123")
//	if err == nil {
//	    fmt.Printf("Key expires in %v\n", ttl)
//	}
func (c *Client) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.redisClient.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}
	return ttl, nil
}

// UpdateExpiration updates the expiration of an existing key.
// Returns an error if the key doesn't exist.
//
// Example:
//
//	err := client.UpdateExpiration(ctx, "user:123", time.Hour)
func (c *Client) UpdateExpiration(ctx context.Context, key string, expiration time.Duration) error {
	ok, err := c.redisClient.Expire(ctx, key, expiration).Result()
	if err != nil {
		return fmt.Errorf("failed to update expiration: %w", err)
	}
	if !ok {
		return fmt.Errorf("key not found: %s", key)
	}
	return nil
}

// GetWithTTL gets a value and its TTL in a single atomic operation.
// Returns the TTL and unmarshals the value into the provided interface.
//
// Example:
//
//	var user User
//	ttl, err := client.GetWithTTL(ctx, "user:123", &user)
//	if err == nil {
//	    fmt.Printf("User %s expires in %v\n", user.Name, ttl)
//	}
func (c *Client) GetWithTTL(ctx context.Context, key string, value interface{}) (time.Duration, error) {
	// Start a pipeline to execute both operations
	pipe := c.redisClient.Pipeline()
	getCmd := pipe.Get(ctx, key)
	ttlCmd := pipe.TTL(ctx, key)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	// Get value
	data, err := getCmd.Bytes()
	if err != nil {
		if err == redis.Nil {
			return 0, fmt.Errorf("key not found: %s", key)
		}
		return 0, fmt.Errorf("failed to get value: %w", err)
	}

	// Unmarshal value
	if err := json.Unmarshal(data, value); err != nil {
		return 0, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	// Get TTL
	ttl, err := ttlCmd.Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}

	return ttl, nil
}
