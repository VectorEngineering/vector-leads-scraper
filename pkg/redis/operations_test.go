package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Vector/vector-leads-scraper/testcontainers"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestRedisOperations(t *testing.T) {
	testcontainers.WithTestContext(t, func(ctx *testcontainers.TestContext) {
		// Create a shared Redis client for cleanup between tests
		redisClient := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", ctx.RedisConfig.Host, ctx.RedisConfig.Port),
			Password: ctx.RedisConfig.Password,
			DB:       0,
		})
		defer redisClient.Close()

		// Helper function to clean up Redis between tests
		cleanup := func() {
			redisClient.FlushAll(context.Background())
		}

		t.Run("Set and Get operations", func(t *testing.T) {
			cleanup()
			client := &Client{
				redisClient: redisClient,
			}

			testCtx := context.Background()
			user := testUser{ID: "123", Name: "John Doe"}

			// Test Set
			err := client.Set(testCtx, "user:123", user, time.Hour)
			require.NoError(t, err)

			// Test Get
			var retrieved testUser
			err = client.Get(testCtx, "user:123", &retrieved)
			require.NoError(t, err)
			assert.Equal(t, user, retrieved)

			// Verify TTL
			ttl, err := client.GetTTL(testCtx, "user:123")
			require.NoError(t, err)
			assert.True(t, ttl > 0)
		})

		t.Run("Delete operation", func(t *testing.T) {
			cleanup()
			client := &Client{
				redisClient: redisClient,
			}

			testCtx := context.Background()
			user := testUser{ID: "123", Name: "John Doe"}

			// Set up test data
			err := client.Set(testCtx, "user:123", user, 0)
			require.NoError(t, err)

			// Test Delete
			err = client.Delete(testCtx, "user:123")
			require.NoError(t, err)

			// Verify key was deleted
			var retrieved testUser
			err = client.Get(testCtx, "user:123", &retrieved)
			assert.Error(t, err)
		})

		t.Run("Update operation", func(t *testing.T) {
			cleanup()
			client := &Client{
				redisClient: redisClient,
			}

			testCtx := context.Background()
			user := testUser{ID: "123", Name: "John Doe"}

			// Set initial value with expiration
			err := client.Set(testCtx, "user:123", user, time.Hour)
			require.NoError(t, err)

			// Update value
			updatedUser := testUser{ID: "123", Name: "Jane Doe"}
			err = client.Update(testCtx, "user:123", updatedUser)
			require.NoError(t, err)

			// Verify value was updated
			var retrieved testUser
			err = client.Get(testCtx, "user:123", &retrieved)
			require.NoError(t, err)
			assert.Equal(t, updatedUser, retrieved)

			// Verify TTL was preserved
			ttl, err := client.GetTTL(testCtx, "user:123")
			require.NoError(t, err)
			assert.True(t, ttl > 0)
		})

		t.Run("SetWithExpiration operation", func(t *testing.T) {
			cleanup()
			client := &Client{
				redisClient: redisClient,
			}

			testCtx := context.Background()
			user := testUser{ID: "123", Name: "John Doe"}

			// Ensure key doesn't exist
			exists, err := redisClient.Exists(testCtx, "user:123").Result()
			require.NoError(t, err)
			require.Equal(t, int64(0), exists)

			// Test SetWithExpiration
			ok, err := client.SetWithExpiration(testCtx, "user:123", user, time.Hour)
			require.NoError(t, err)
			assert.True(t, ok)

			// Verify expiration was set
			ttl, err := client.GetTTL(testCtx, "user:123")
			require.NoError(t, err)
			assert.True(t, ttl > 0)

			// Try setting again (should fail)
			ok, err = client.SetWithExpiration(testCtx, "user:123", user, time.Hour)
			require.NoError(t, err)
			assert.False(t, ok)
		})

		t.Run("GetTTL operation", func(t *testing.T) {
			cleanup()
			client := &Client{
				redisClient: redisClient,
			}

			testCtx := context.Background()
			user := testUser{ID: "123", Name: "John Doe"}

			// Set value with expiration
			err := client.Set(testCtx, "user:123", user, time.Hour)
			require.NoError(t, err)

			// Test GetTTL
			ttl, err := client.GetTTL(testCtx, "user:123")
			require.NoError(t, err)
			assert.True(t, ttl > 0)
		})

		t.Run("UpdateExpiration operation", func(t *testing.T) {
			cleanup()
			client := &Client{
				redisClient: redisClient,
			}

			testCtx := context.Background()
			user := testUser{ID: "123", Name: "John Doe"}

			// Set value without expiration
			err := client.Set(testCtx, "user:123", user, 0)
			require.NoError(t, err)

			// Update expiration
			err = client.UpdateExpiration(testCtx, "user:123", time.Hour)
			require.NoError(t, err)

			// Verify TTL was set
			ttl, err := client.GetTTL(testCtx, "user:123")
			require.NoError(t, err)
			assert.True(t, ttl > 0)
		})

		t.Run("GetWithTTL operation", func(t *testing.T) {
			cleanup()
			client := &Client{
				redisClient: redisClient,
			}

			testCtx := context.Background()
			user := testUser{ID: "123", Name: "John Doe"}

			// Set value with expiration
			err := client.Set(testCtx, "user:123", user, time.Hour)
			require.NoError(t, err)

			// Test GetWithTTL
			var retrieved testUser
			ttl, err := client.GetWithTTL(testCtx, "user:123", &retrieved)
			require.NoError(t, err)
			assert.Equal(t, user, retrieved)
			assert.True(t, ttl > 0)
		})

		t.Run("Error cases", func(t *testing.T) {
			cleanup()
			client := &Client{
				redisClient: redisClient,
			}

			testCtx := context.Background()

			// Test Get non-existent key
			var user testUser
			err := client.Get(testCtx, "nonexistent", &user)
			assert.Error(t, err)

			// Test Update non-existent key
			err = client.Update(testCtx, "nonexistent", user)
			assert.Error(t, err)

			// Test UpdateExpiration non-existent key
			err = client.UpdateExpiration(testCtx, "nonexistent", time.Hour)
			assert.Error(t, err)

			// Test Set with invalid value
			err = client.Set(testCtx, "invalid", make(chan int), 0)
			assert.Error(t, err)
		})
	})
}
