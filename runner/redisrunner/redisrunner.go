// Package redisrunner provides Redis-backed task processing functionality.
package redisrunner

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Vector/vector-leads-scraper/internal/taskhandler"
	"github.com/Vector/vector-leads-scraper/pkg/redis/tasks"
	"github.com/Vector/vector-leads-scraper/runner"
	"github.com/hibiken/asynq"
)

// RedisRunner manages Redis-backed task processing.
type RedisRunner struct {
	cfg     *runner.Config
	handler *taskhandler.Handler
	logger  *log.Logger
}

// New creates a new Redis runner with the provided configuration.
func New(cfg *runner.Config) (*RedisRunner, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	logger := log.New(log.Writer(), "[RedisRunner] ", log.LstdFlags)

	// Initialize task handler with default options
	handlerOpts := &taskhandler.Options{
		MaxRetries:    3,
		RetryInterval: 5 * time.Second,
		TaskTypes:     []string{tasks.TypeScrapeGMaps},
		Logger:        logger,
	}

	handler, err := taskhandler.New(cfg, handlerOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create task handler: %w", err)
	}

	return &RedisRunner{
		cfg:     cfg,
		handler: handler,
		logger:  logger,
	}, nil
}

// Run starts the Redis runner and begins processing tasks.
func (r *RedisRunner) Run(ctx context.Context) error {
	r.logger.Println("Starting Redis runner...")
	return r.handler.Run(ctx, r.cfg.RedisWorkers)
}

// Close gracefully shuts down the Redis runner.
func (r *RedisRunner) Close(ctx context.Context) error {
	r.logger.Println("Shutting down Redis runner...")
	return r.handler.Close(ctx)
}

// EnqueueTask enqueues a new task for processing.
func (r *RedisRunner) EnqueueTask(ctx context.Context, taskType string, payload []byte, opts ...asynq.Option) error {
	return r.handler.EnqueueTask(ctx, taskType, payload, opts...)
}

// GetHandler returns the task handler for a specific task type.
func (r *RedisRunner) GetHandler(taskType string) (tasks.TaskHandler, bool) {
	return r.handler.GetHandler(taskType)
}

// RegisterHandler registers a new task handler for a specific task type.
func (r *RedisRunner) RegisterHandler(taskType string, handler tasks.TaskHandler) error {
	return r.handler.RegisterHandler(taskType, handler)
}

// UnregisterHandler removes a task handler for a specific task type.
func (r *RedisRunner) UnregisterHandler(taskType string) {
	r.handler.UnregisterHandler(taskType)
}
