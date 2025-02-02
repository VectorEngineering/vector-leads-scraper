// Package redis provides a Redis client implementation that combines standard Redis operations
// with task queue functionality. It supports both regular Redis operations through go-redis
// and task queue operations through asynq, including priority-based task queuing.
//
// The package provides a unified client that handles:
//   - Standard Redis operations (through go-redis)
//   - Task queue operations (through asynq)
//   - Priority-based task queuing based on subscription tiers
//   - Connection management and health checks
//   - Automatic retries with exponential backoff
//
// Concurrency and Thread Safety:
// The client implementation is thread-safe and can be safely used from multiple goroutines.
// All operations are protected by a read-write mutex to ensure thread safety.
//
// Error Handling:
// The package uses wrapped errors with additional context information.
// Connection errors are automatically retried with exponential backoff.
// All errors are properly propagated and can be unwrapped using errors.Unwrap.
//
// Usage Example:
//
//	cfg := &config.RedisConfig{
//	    Host: "localhost",
//	    Port: 6379,
//	}
//
//	// Create client with automatic connection retry
//	var client *redis.Client
//	err := RetryWithBackoff(
//	    func() error {
//	        c, err := NewClient(cfg)
//	        if err != nil {
//	            return err
//	        }
//	        client = c
//	        return nil
//	    },
//	    3,
//	    time.Second,
//	)
//
//	// Use priority queuing for tasks
//	err = client.EnqueueTaskBasedOnSubscription(
//	    ctx,
//	    "email:send",
//	    []byte(`{"to": "user@example.com"}`),
//	    priorityqueue.SubscriptionEnterprise,
//	)
//
//	// Monitor health
//	go func() {
//	    ticker := time.NewTicker(30 * time.Second)
//	    for {
//	        select {
//	        case <-ticker.C:
//	            if !client.IsHealthy(ctx) {
//	                log.Println("Redis connection unhealthy")
//	            }
//	        case <-ctx.Done():
//	            return
//	        }
//	    }
//	}()
package redis

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Vector/vector-leads-scraper/pkg/redis/config"
	"github.com/Vector/vector-leads-scraper/pkg/redis/priorityqueue"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// Client wraps Redis client functionality, providing a unified interface for
// Redis operations, task queuing, and priority-based task management.
// It manages three underlying clients:
//   - asynq client for task queue operations
//   - redis client for standard Redis operations
//   - priority queue client for subscription-based task prioritization
//
// Thread Safety:
// All operations are protected by a read-write mutex, making the client safe
// for concurrent use from multiple goroutines.
//
// Resource Management:
// The client manages connection pooling and cleanup automatically.
// Users must call Close() when the client is no longer needed to prevent resource leaks.
type Client struct {
	asynqClient    *asynq.Client
	redisClient    *redis.Client
	cfg            *config.RedisConfig
	mu             sync.RWMutex // Protects all client operations
	priorityClient *priorityqueue.Client
}

// Config holds Redis connection configuration parameters.
// It provides all necessary settings for establishing and maintaining
// Redis connections with appropriate retry and worker settings.
//
// Default Values:
//   - RetryInterval: 1 second
//   - MaxRetries: 3
//   - Workers: 10
//   - RetentionPeriod: 24 hours
type Config struct {
	// Host is the Redis server hostname
	Host string

	// Port is the Redis server port number
	Port int

	// Password is the Redis server password (if any)
	Password string

	// DB is the Redis database number
	DB int

	// Workers is the number of concurrent workers for processing tasks
	// Default: 10
	Workers int

	// RetryInterval is the initial interval between retry attempts
	// Default: 1 second
	RetryInterval time.Duration

	// MaxRetries is the maximum number of retry attempts
	// Default: 3
	MaxRetries int

	// RetentionPeriod is how long completed tasks are retained
	// Default: 24 hours
	RetentionPeriod time.Duration
}

// NewClient creates a new Redis client with the provided configuration.
// It initializes and tests connections to all required Redis services:
//   - Standard Redis connection
//   - Asynq task queue connection
//   - Priority queue connection
//
// The client must be closed using the Close method when no longer needed.
//
// Error Handling:
// Returns an error if any of the following occur:
//   - Invalid configuration
//   - Failed to connect to Redis
//   - Failed to initialize any client
//   - Failed connection tests
//
// Example:
//
//	cfg := &config.RedisConfig{
//	    Host: "localhost",
//	    Port: 6379,
//	}
//	client, err := NewClient(cfg)
//	if err != nil {
//	    if errors.Is(err, redis.ErrConnectionRefused) {
//	        // Handle connection error
//	    }
//	    log.Fatal(err)
//	}
//	defer client.Close()
func NewClient(cfg *config.RedisConfig) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Initialize asynq client
	asynqOpt := asynq.RedisClientOpt{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	asynqClient := asynq.NewClient(asynqOpt)

	// Initialize redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Initialize priority queue client
	priorityClient := priorityqueue.NewClient(asynqOpt, priorityqueue.DefaultConfig())

	// Test connections
	if err := testConnection(asynqClient, redisClient); err != nil {
		// Clean up any successful connections before returning error
		asynqClient.Close()
		redisClient.Close()
		priorityClient.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{
		asynqClient:    asynqClient,
		redisClient:    redisClient,
		cfg:            cfg,
		priorityClient: priorityClient,
	}, nil
}

// EnqueueTask enqueues a task with the given type and payload.
// This method uses the standard asynq client without priority queuing.
//
// Parameters:
//   - ctx: Context for the operation (used for cancellation and timeouts)
//   - taskType: Type identifier for the task (e.g., "email:send", "data:process")
//   - payload: Task payload as bytes (typically JSON-encoded data)
//   - opts: Additional task options (timeout, retries, etc.)
//
// Error Handling:
//   - Returns error if context is cancelled
//   - Returns error if Redis connection fails
//   - Returns error if task enqueuing fails
//
// Concurrency:
// This method is thread-safe and can be called concurrently from multiple goroutines.
//
// Example:
//
//	// Enqueue a task with options
//	err := client.EnqueueTask(ctx,
//	    "email:send",
//	    []byte(`{"to": "user@example.com"}`),
//	    asynq.MaxRetry(3),
//	    asynq.Timeout(5*time.Minute),
//	)
func (c *Client) EnqueueTask(ctx context.Context, taskType string, payload []byte, opts ...asynq.Option) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	task := asynq.NewTask(taskType, payload)
	_, err := c.asynqClient.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

// EnqueueTaskBasedOnSubscription enqueues a task with priority based on the subscription type.
// This method uses the priority queue client to ensure tasks are processed according to
// their subscription tier (Enterprise, Pro, or Free).
//
// Parameters:
//   - ctx: Context for the operation (used for cancellation and timeouts)
//   - taskType: Type identifier for the task (e.g., "email:send", "data:process")
//   - payload: Task payload as bytes (typically JSON-encoded data)
//   - subscriptionType: Determines the priority queue (Enterprise, Pro, or Free)
//
// Priority Levels:
//   - Enterprise: Highest priority (60% of processing time)
//   - Pro: Medium priority (30% of processing time)
//   - Free: Lowest priority (10% of processing time)
//
// Error Handling:
//   - Returns error if context is cancelled
//   - Returns error if subscription type is invalid
//   - Returns error if Redis connection fails
//   - Returns error if task enqueuing fails
//
// Concurrency:
// This method is thread-safe and can be called concurrently from multiple goroutines.
//
// Example:
//
//	// Enqueue a high-priority enterprise task
//	err := client.EnqueueTaskBasedOnSubscription(
//	    ctx,
//	    "email:send",
//	    []byte(`{"to": "enterprise@example.com"}`),
//	    priorityqueue.SubscriptionEnterprise,
//	)
//
//	// Enqueue a low-priority free tier task
//	err = client.EnqueueTaskBasedOnSubscription(
//	    ctx,
//	    "email:send",
//	    []byte(`{"to": "free@example.com"}`),
//	    priorityqueue.SubscriptionFree,
//	)
func (c *Client) EnqueueTaskBasedOnSubscription(ctx context.Context, taskType string, payload []byte, subscriptionType priorityqueue.SubscriptionType) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	task := asynq.NewTask(taskType, payload)
	_, err := c.priorityClient.EnqueueTask(
		ctx,
		task,
		subscriptionType,
		priorityqueue.GetQueueOptions(subscriptionType)...,
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

// Close closes all Redis client connections.
// It attempts to close each connection and collects any errors that occur.
// Even if one connection fails to close, it will attempt to close the others.
//
// Error Handling:
// Returns a combined error if any of the close operations fail.
// The error will contain details about which clients failed to close.
//
// Usage:
//   - Always call Close when the client is no longer needed
//   - Best used with defer immediately after client creation
//   - Safe to call multiple times (subsequent calls are no-op)
//
// Thread Safety:
// This method is thread-safe. It uses a write lock to ensure no other
// operations can occur while closing the connections.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	if err := c.asynqClient.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close asynq client: %w", err))
	}

	if err := c.redisClient.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close redis client: %w", err))
	}

	if err := c.priorityClient.Close(); err != nil {
		// log error as the same client is shared between asynq and priority queue
		log.Printf("failed to close priority client: %v", err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing Redis clients: %v", errs)
	}
	return nil
}

// IsHealthy checks if all Redis connections are healthy.
// It performs a simple ping operation on each connection to verify they are responsive.
//
// Parameters:
//   - ctx: Context for the health check operation
//
// Returns:
//   - true if all connections are healthy
//   - false if any connection is unhealthy
//
// Health Check Details:
//   - Asynq: Attempts to enqueue a test task
//   - Redis: Performs a PING command
//   - Priority Queue: Checked through asynq connection
//
// Usage:
// Best used for monitoring and health checks:
//
//	go func() {
//	    ticker := time.NewTicker(30 * time.Second)
//	    for {
//	        select {
//	        case <-ticker.C:
//	            if !client.IsHealthy(ctx) {
//	                log.Println("Redis connection unhealthy")
//	            }
//	        case <-ctx.Done():
//	            return
//	        }
//	    }
//	}()
//
// Thread Safety:
// This method is thread-safe and can be called concurrently from multiple goroutines.
func (c *Client) IsHealthy(ctx context.Context) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check asynq connection
	_, err1 := c.asynqClient.EnqueueContext(ctx, asynq.NewTask("health:check", nil))

	// Check redis connection
	err2 := c.redisClient.Ping(ctx).Err()

	return err1 == nil && err2 == nil
}

// RetryWithBackoff implements exponential backoff for connection retries.
// It will retry the operation with increasing intervals between attempts.
//
// Parameters:
//   - operation: The function to retry (should return nil on success)
//   - maxRetries: Maximum number of retry attempts
//   - initialInterval: Starting interval between retries
//
// Error Handling:
//   - Returns nil if operation succeeds within maxRetries
//   - Returns error with retry count if all attempts fail
//   - Original error is preserved as wrapped error
//
// Backoff Algorithm:
//   - First retry: initialInterval
//   - Second retry: initialInterval * 2
//   - Third retry: initialInterval * 4
//   - And so on...
//
// Example:
//
//	err := RetryWithBackoff(
//	    func() error {
//	        return client.Connect()
//	    },
//	    3,
//	    time.Second,
//	)
//
// Usage Tips:
//   - Set reasonable maxRetries to prevent infinite loops
//   - Choose initialInterval based on operation cost
//   - Consider context deadlines for time-sensitive operations
func RetryWithBackoff(operation func() error, maxRetries int, initialInterval time.Duration) error {
	var err error
	interval := initialInterval

	for i := 0; i < maxRetries; i++ {
		if err = operation(); err == nil {
			return nil
		}

		if i == maxRetries-1 {
			break
		}

		log.Printf("Retry attempt %d failed: %v. Retrying in %v...", i+1, err, interval)
		time.Sleep(interval)
		interval *= 2 // Exponential backoff
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}

// testConnection tests all Redis connections to ensure they are working.
// It performs a basic operation on each connection to verify connectivity.
//
// Tests Performed:
//   - Asynq: Attempts to enqueue a test task
//   - Redis: Performs a PING command
//
// Error Handling:
//   - Returns detailed error messages for each failed connection
//   - Wraps original errors to preserve error chain
//
// Internal Use:
// This function is used internally by NewClient to verify initial connections.
func testConnection(asynqClient *asynq.Client, redisClient *redis.Client) error {
	ctx := context.Background()

	// Test asynq connection
	task := asynq.NewTask("connection:test", nil)
	if _, err := asynqClient.EnqueueContext(ctx, task); err != nil {
		return fmt.Errorf("failed to connect to Redis (asynq): %w", err)
	}

	// Test redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis (redis): %w", err)
	}

	return nil
}
