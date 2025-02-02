package redis

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Vector/vector-leads-scraper/pkg/redis/config"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// Client wraps Redis client functionality
type Client struct {
	asynqClient *asynq.Client
	redisClient *redis.Client
	cfg         *config.RedisConfig
	mu          sync.RWMutex
}

// Config holds Redis connection configuration parameters
type Config struct {
	Host            string
	Port            int
	Password        string
	DB              int
	Workers         int
	RetryInterval   time.Duration
	MaxRetries      int
	RetentionPeriod time.Duration
}

// NewClient creates a new Redis client with the provided configuration
func NewClient(cfg *config.RedisConfig) (*Client, error) {
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

	// Test connections
	if err := testConnection(asynqClient, redisClient); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{
		asynqClient: asynqClient,
		redisClient: redisClient,
		cfg:         cfg,
	}, nil
}

// EnqueueTask enqueues a task with the given type and payload
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

// Close closes both Redis client connections
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

	if len(errs) > 0 {
		return fmt.Errorf("errors closing Redis clients: %v", errs)
	}
	return nil
}

// IsHealthy checks if both Redis connections are healthy
func (c *Client) IsHealthy(ctx context.Context) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check asynq connection
	_, err1 := c.asynqClient.EnqueueContext(ctx, asynq.NewTask("health:check", nil))
	
	// Check redis connection
	err2 := c.redisClient.Ping(ctx).Err()

	return err1 == nil && err2 == nil
}

// RetryWithBackoff implements exponential backoff for connection retries
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

// testConnection tests both Redis connections
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
