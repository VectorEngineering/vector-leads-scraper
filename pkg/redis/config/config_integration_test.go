package config

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRedisConfigIntegration_WithContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start Redis container
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:latest",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("Ready to accept connections"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v", err)
	}
	defer redisContainer.Terminate(ctx)

	// Get container host and port
	host, err := redisContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}

	// Set environment variables for Redis configuration
	os.Setenv("REDIS_HOST", host)
	os.Setenv("REDIS_PORT", port.Port())
	os.Setenv("REDIS_DB", "0")
	os.Setenv("REDIS_WORKERS", "5")
	os.Setenv("REDIS_RETRY_INTERVAL", "1s")
	os.Setenv("REDIS_MAX_RETRIES", "3")
	os.Setenv("REDIS_RETENTION_DAYS", "7")
	defer os.Clearenv()

	// Create Redis configuration
	cfg, err := NewRedisConfig()
	if err != nil {
		t.Fatalf("Failed to create Redis configuration: %v", err)
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.GetAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	defer client.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("Failed to ping Redis: %v", err)
	}

	// Test basic operations
	key := "test:key"
	value := "test:value"

	// Set value
	if err := client.Set(ctx, key, value, 0).Err(); err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Get value
	got, err := client.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if got != value {
		t.Errorf("Got %q, want %q", got, value)
	}

	// Test configuration values
	if cfg.Workers != 5 {
		t.Errorf("Workers = %d, want 5", cfg.Workers)
	}

	if cfg.RetryInterval != time.Second {
		t.Errorf("RetryInterval = %v, want 1s", cfg.RetryInterval)
	}

	if cfg.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", cfg.MaxRetries)
	}

	if cfg.RetentionPeriod != 7*24*time.Hour {
		t.Errorf("RetentionPeriod = %v, want 168h", cfg.RetentionPeriod)
	}

	// Test queue priorities
	if len(cfg.QueuePriorities) != len(DefaultQueuePriorities) {
		t.Errorf("QueuePriorities length = %d, want %d", len(cfg.QueuePriorities), len(DefaultQueuePriorities))
	}

	for queue, priority := range DefaultQueuePriorities {
		if cfg.QueuePriorities[queue] != priority {
			t.Errorf("QueuePriorities[%q] = %d, want %d", queue, cfg.QueuePriorities[queue], priority)
		}
	}
} 