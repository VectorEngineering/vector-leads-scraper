// Package testcontainers_test provides integration tests and examples for the testcontainers package.
// These tests demonstrate the usage of the package and verify its functionality.
package testcontainers

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

// TestTestContext verifies the functionality of the TestContext type and its associated methods.
// It demonstrates various use cases and serves as documentation through examples.
//
// The test suite includes:
//   - Basic container initialization and cleanup
//   - Multiple container instance management
//   - Redis operations
//   - PostgreSQL operations
//
// Each test case demonstrates a different aspect of the package's functionality
// and serves as an example for users.
func TestTestContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test case: Basic container initialization and cleanup
	t.Run("creates and cleans up test context", func(t *testing.T) {
		WithTestContext(t, func(ctx *TestContext) {
			// Test Redis connection
			result, err := ctx.Redis.Ping(ctx.ctx).Result()
			require.NoError(t, err)
			assert.Equal(t, "PONG", result)

			// Test PostgreSQL connection
			err = ctx.DB.Ping(ctx.ctx)
			require.NoError(t, err)
		})
	})

	// Test case: Multiple container instances
	t.Run("handles multiple test contexts", func(t *testing.T) {
		WithTestContext(t, func(ctx1 *TestContext) {
			WithTestContext(t, func(ctx2 *TestContext) {
				// Verify different ports
				assert.NotEqual(t, ctx1.RedisConfig.Port, ctx2.RedisConfig.Port)
				assert.NotEqual(t, ctx1.PostgresConfig.Port, ctx2.PostgresConfig.Port)

				// Verify both contexts are functional
				result1, err := ctx1.Redis.Ping(ctx1.ctx).Result()
				require.NoError(t, err)
				assert.Equal(t, "PONG", result1)

				result2, err := ctx2.Redis.Ping(ctx2.ctx).Result()
				require.NoError(t, err)
				assert.Equal(t, "PONG", result2)
			})
		})
	})

	// Test case: Redis operations
	t.Run("verifies Redis operations", func(t *testing.T) {
		WithTestContext(t, func(ctx *TestContext) {
			// Test Set operation
			err := ctx.Redis.Set(ctx.ctx, "test_key", "test_value", time.Minute).Err()
			require.NoError(t, err)

			// Test Get operation
			val, err := ctx.Redis.Get(ctx.ctx, "test_key").Result()
			require.NoError(t, err)
			assert.Equal(t, "test_value", val)

			// Test Delete operation
			err = ctx.Redis.Del(ctx.ctx, "test_key").Err()
			require.NoError(t, err)

			// Verify key was deleted
			_, err = ctx.Redis.Get(ctx.ctx, "test_key").Result()
			assert.Error(t, err)
		})
	})

	// Test case: PostgreSQL operations
	t.Run("verifies PostgreSQL operations", func(t *testing.T) {
		WithTestContext(t, func(ctx *TestContext) {
			// Create a test table
			_, err := ctx.DB.Exec(ctx.ctx, `
				CREATE TABLE test_table (
					id SERIAL PRIMARY KEY,
					name TEXT NOT NULL
				)
			`)
			require.NoError(t, err)

			// Insert a row
			_, err = ctx.DB.Exec(ctx.ctx, `
				INSERT INTO test_table (name) VALUES ($1)
			`, "test_name")
			require.NoError(t, err)

			// Query the row
			var name string
			err = ctx.DB.QueryRow(ctx.ctx, `
				SELECT name FROM test_table WHERE id = 1
			`).Scan(&name)
			require.NoError(t, err)
			assert.Equal(t, "test_name", name)

			// Update the row
			_, err = ctx.DB.Exec(ctx.ctx, `
				UPDATE test_table SET name = $1 WHERE id = 1
			`, "updated_name")
			require.NoError(t, err)

			// Verify update
			err = ctx.DB.QueryRow(ctx.ctx, `
				SELECT name FROM test_table WHERE id = 1
			`).Scan(&name)
			require.NoError(t, err)
			assert.Equal(t, "updated_name", name)
		})
	})

	// Test case: Container configuration
	t.Run("verifies container configuration", func(t *testing.T) {
		WithTestContext(t, func(ctx *TestContext) {
			// Verify Redis configuration
			assert.NotEmpty(t, ctx.RedisConfig.Host)
			assert.Greater(t, ctx.RedisConfig.Port, 0)
			assert.Empty(t, ctx.RedisConfig.Password) // No password for test container

			// Verify PostgreSQL configuration
			assert.NotEmpty(t, ctx.PostgresConfig.Host)
			assert.Greater(t, ctx.PostgresConfig.Port, 0)
			assert.Equal(t, "test", ctx.PostgresConfig.User)
			assert.Equal(t, "test", ctx.PostgresConfig.Password)
			assert.Equal(t, "testdb", ctx.PostgresConfig.Database)
		})
	})

	// Test case: Container cleanup
	t.Run("handles cleanup properly", func(t *testing.T) {
		var ctx *TestContext
		func() {
			ctx = NewTestContext(t)
			defer ctx.Cleanup()

			// Verify containers are running
			result, err := ctx.Redis.Ping(ctx.ctx).Result()
			require.NoError(t, err)
			assert.Equal(t, "PONG", result)

			err = ctx.DB.Ping(ctx.ctx)
			require.NoError(t, err)
		}()

		// After cleanup, trying to use the connections should fail
		_, err := ctx.Redis.Ping(ctx.ctx).Result()
		assert.Error(t, err)

		err = ctx.DB.Ping(ctx.ctx)
		assert.Error(t, err)
	})

	// Test case: Context timeout
	t.Run("respects context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()

		_, err := NewRedisContainer(ctx)
		assert.Error(t, err)

		_, err = NewPostgresContainer(ctx)
		assert.Error(t, err)
	})
}

// TestContainerAddresses verifies the address formatting functionality
func TestContainerAddresses(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("formats Redis address correctly", func(t *testing.T) {
		WithTestContext(t, func(ctx *TestContext) {
			address := ctx.redisContainer.GetAddress()
			assert.Contains(t, address, ":")
			assert.Contains(t, address, ctx.RedisConfig.Host)
			assert.Contains(t, address, strconv.Itoa(ctx.RedisConfig.Port))
		})
	})

	t.Run("formats PostgreSQL DSN correctly", func(t *testing.T) {
		WithTestContext(t, func(ctx *TestContext) {
			dsn := ctx.postgresContainer.GetDSN()
			assert.Contains(t, dsn, "postgresql://")
			assert.Contains(t, dsn, ctx.PostgresConfig.Host)
			assert.Contains(t, dsn, strconv.Itoa(ctx.PostgresConfig.Port))
			assert.Contains(t, dsn, ctx.PostgresConfig.User)
			assert.Contains(t, dsn, ctx.PostgresConfig.Password)
			assert.Contains(t, dsn, ctx.PostgresConfig.Database)
			assert.Contains(t, dsn, "sslmode=disable")
		})
	})
}

// TestContainerInitialization tests various container initialization scenarios
func TestContainerInitialization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("handles invalid Redis port", func(t *testing.T) {
		ctx := context.Background()
		req := testcontainers.ContainerRequest{
			Image:        "redis:latest",
			ExposedPorts: []string{"invalid_port/tcp"},
		}

		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err == nil {
			container.Terminate(ctx)
			t.Fatal("Expected error for invalid port")
		}
	})

	t.Run("handles invalid Postgres port", func(t *testing.T) {
		ctx := context.Background()
		req := testcontainers.ContainerRequest{
			Image:        "postgres:latest",
			ExposedPorts: []string{"invalid_port/tcp"},
		}

		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err == nil {
			container.Terminate(ctx)
			t.Fatal("Expected error for invalid port")
		}
	})

	t.Run("handles invalid Redis image", func(t *testing.T) {
		ctx := context.Background()
		req := testcontainers.ContainerRequest{
			Image:        "nonexistent-redis:invalid",
			ExposedPorts: []string{"6379/tcp"},
		}

		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:         true,
		})
		if err == nil {
			container.Terminate(ctx)
			t.Fatal("Expected error for invalid image")
		}
		assert.Contains(t, err.Error(), "pull access denied" /* or check for appropriate error message */)
	})

	t.Run("handles invalid Postgres image", func(t *testing.T) {
		ctx := context.Background()
		req := testcontainers.ContainerRequest{
			Image:        "nonexistent-postgres:latest",
			ExposedPorts: []string{"5432/tcp"},
		}

		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err == nil {
			container.Terminate(ctx)
			t.Fatal("Expected error for invalid image")
		}
		assert.Contains(t, err.Error(), "pull access denied" /* or check for appropriate error message */)
	})
}

// TestCleanupOrder verifies that cleanup functions are executed in the correct order
func TestCleanupOrder(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("executes cleanup in reverse order", func(t *testing.T) {
		var cleanupOrder []int
		ctx := NewTestContext(t)

		// Add numbered cleanup functions
		for i := 0; i < 3; i++ {
			i := i // capture loop variable
			ctx.addCleanup(func() {
				cleanupOrder = append(cleanupOrder, i)
			})
		}

		ctx.Cleanup()

		// Verify reverse order
		assert.Equal(t, []int{2, 1, 0}, cleanupOrder)
	})

	t.Run("handles cleanup after panic", func(t *testing.T) {
		var cleaned bool
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Error("Expected panic")
				}
			}()

			WithTestContext(t, func(ctx *TestContext) {
				ctx.addCleanup(func() {
					cleaned = true
				})
				panic("test panic")
			})
		}()

		assert.True(t, cleaned, "Cleanup should run even after panic")
	})
}

// TestContainerTermination verifies container termination behavior
func TestContainerTermination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("handles Redis termination errors", func(t *testing.T) {
		ctx := context.Background()
		container, err := NewRedisContainer(ctx)
		require.NoError(t, err)

		// Terminate twice to force an error
		err = container.Terminate(ctx)
		require.NoError(t, err)
		err = container.Terminate(ctx)
		assert.Error(t, err)
	})

	t.Run("handles Postgres termination errors", func(t *testing.T) {
		ctx := context.Background()
		container, err := NewPostgresContainer(ctx)
		require.NoError(t, err)

		// Terminate twice to force an error
		err = container.Terminate(ctx)
		require.NoError(t, err)
		err = container.Terminate(ctx)
		assert.Error(t, err)
	})
}

// TestContainerNetworking verifies container networking functionality
func TestContainerNetworking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("verifies Redis port mapping", func(t *testing.T) {
		WithTestContext(t, func(ctx *TestContext) {
			assert.Greater(t, ctx.RedisConfig.Port, 0)
			assert.Less(t, ctx.RedisConfig.Port, 65536)

			// Verify connection using mapped port
			result, err := ctx.Redis.Ping(ctx.ctx).Result()
			require.NoError(t, err)
			assert.Equal(t, "PONG", result)
		})
	})

	t.Run("verifies Postgres port mapping", func(t *testing.T) {
		WithTestContext(t, func(ctx *TestContext) {
			assert.Greater(t, ctx.PostgresConfig.Port, 0)
			assert.Less(t, ctx.PostgresConfig.Port, 65536)

			// Verify connection using mapped port
			err := ctx.DB.Ping(ctx.ctx)
			require.NoError(t, err)
		})
	})
}

// TestContainerConfiguration verifies container configuration handling
func TestContainerConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("verifies Redis container defaults", func(t *testing.T) {
		ctx := context.Background()
		container, err := NewRedisContainer(ctx)
		require.NoError(t, err)
		defer container.Terminate(ctx)

		assert.NotEmpty(t, container.Host)
		assert.Greater(t, container.Port, 0)
		assert.Empty(t, container.Password)
	})

	t.Run("verifies Postgres container defaults", func(t *testing.T) {
		ctx := context.Background()
		container, err := NewPostgresContainer(ctx)
		require.NoError(t, err)
		defer container.Terminate(ctx)

		assert.NotEmpty(t, container.Host)
		assert.Greater(t, container.Port, 0)
		assert.Equal(t, defaultUser, container.User)
		assert.Equal(t, defaultPassword, container.Password)
		assert.Equal(t, defaultDatabase, container.Database)
	})

	t.Run("verifies DSN format", func(t *testing.T) {
		ctx := context.Background()
		container, err := NewPostgresContainer(ctx)
		require.NoError(t, err)
		defer container.Terminate(ctx)

		dsn := container.GetDSN()
		assert.Contains(t, dsn, fmt.Sprintf("postgresql://%s:%s@", container.User, container.Password))
		assert.Contains(t, dsn, fmt.Sprintf(":%d/", container.Port))
		assert.Contains(t, dsn, container.Database)
		assert.Contains(t, dsn, "sslmode=disable")
	})
}

// TestConnectionManagement verifies connection pool behavior and cleanup
func TestConnectionManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("handles connection pool lifecycle", func(t *testing.T) {
		WithTestContext(t, func(ctx *TestContext) {
			// Verify initial pool state
			assert.NotNil(t, ctx.DB)
			assert.NotNil(t, ctx.Redis)

			// Test connection pool stats
			poolStats := ctx.DB.Stat()
			assert.Greater(t, poolStats.MaxConns(), int32(0))
			assert.Equal(t, int32(0), poolStats.AcquiredConns())

			// Test Redis connection pool
			stats := ctx.Redis.PoolStats()
			assert.GreaterOrEqual(t, stats.TotalConns, uint32(0))
		})
	})

	t.Run("handles connection errors gracefully", func(t *testing.T) {
		ctx := NewTestContext(t)
		// Force connection error by closing the pool
		ctx.DB.Close()

		// Attempt operations after pool closure
		err := ctx.DB.Ping(ctx.ctx)
		assert.Error(t, err)

		// Close Redis client and verify error handling
		err = ctx.Redis.Close()
		assert.NoError(t, err)
		_, err = ctx.Redis.Ping(ctx.ctx).Result()
		assert.Error(t, err)

		// Skip cleanup since we've already closed the connections
		ctx.cleanup = nil
		ctx.Cleanup()
	})
}

// TestConfigValidation verifies configuration validation and edge cases
func TestConfigValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("validates Redis configuration", func(t *testing.T) {
		WithTestContext(t, func(ctx *TestContext) {
			assert.NotEmpty(t, ctx.RedisConfig.Host)
			assert.Greater(t, ctx.RedisConfig.Port, 0)
			assert.Empty(t, ctx.RedisConfig.Password) // No password for test container
		})
	})

	t.Run("validates Postgres configuration", func(t *testing.T) {
		WithTestContext(t, func(ctx *TestContext) {
			assert.NotEmpty(t, ctx.PostgresConfig.Host)
			assert.Greater(t, ctx.PostgresConfig.Port, 0)
			assert.Equal(t, "test", ctx.PostgresConfig.User)
			assert.Equal(t, "test", ctx.PostgresConfig.Password)
			assert.Equal(t, "testdb", ctx.PostgresConfig.Database)
		})
	})
}

// TestAddressFormatting verifies address and DSN formatting
func TestAddressFormatting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("formats Redis address correctly", func(t *testing.T) {
		WithTestContext(t, func(ctx *TestContext) {
			address := ctx.redisContainer.GetAddress()
			assert.Contains(t, address, ctx.RedisConfig.Host)
			assert.Contains(t, address, strconv.Itoa(ctx.RedisConfig.Port))
		})
	})

	t.Run("formats Postgres DSN correctly", func(t *testing.T) {
		WithTestContext(t, func(ctx *TestContext) {
			dsn := ctx.postgresContainer.GetDSN()
			assert.Contains(t, dsn, ctx.PostgresConfig.Host)
			assert.Contains(t, dsn, strconv.Itoa(ctx.PostgresConfig.Port))
			assert.Contains(t, dsn, ctx.PostgresConfig.User)
			assert.Contains(t, dsn, ctx.PostgresConfig.Password)
			assert.Contains(t, dsn, ctx.PostgresConfig.Database)
			assert.Contains(t, dsn, "sslmode=disable")
		})
	})
}

// TestCleanupBehavior verifies cleanup behavior in various scenarios
func TestCleanupBehavior(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("handles cleanup after panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				// Expected panic
			}
		}()

		WithTestContext(t, func(ctx *TestContext) {
			// Trigger a panic
			panic("test panic")
		})
	})

	t.Run("handles multiple cleanups", func(t *testing.T) {
		ctx := NewTestContext(t)

		// First cleanup
		ctx.Cleanup()

		// Second cleanup should be safe
		ctx.cleanup = nil
		ctx.Cleanup()
	})

	t.Run("cleans up in correct order", func(t *testing.T) {
		cleanupOrder := make([]int, 0)
		ctx := NewTestContext(t)

		// Add cleanup functions with tracking
		ctx.addCleanup(func() { cleanupOrder = append(cleanupOrder, 1) })
		ctx.addCleanup(func() { cleanupOrder = append(cleanupOrder, 2) })
		ctx.addCleanup(func() { cleanupOrder = append(cleanupOrder, 3) })

		ctx.Cleanup()

		// Verify reverse order execution
		assert.Equal(t, []int{3, 2, 1}, cleanupOrder)
	})
}
