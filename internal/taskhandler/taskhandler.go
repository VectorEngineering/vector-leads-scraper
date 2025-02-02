// Package taskhandler provides task handling functionality for Redis-backed operations.
package taskhandler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Vector/vector-leads-scraper/pkg/redis"
	"github.com/Vector/vector-leads-scraper/pkg/redis/tasks"
	"github.com/Vector/vector-leads-scraper/runner"
	"github.com/hibiken/asynq"
)

// Handler manages task processing for Redis operations.
type Handler struct {
	mux        *asynq.ServeMux
	handlers   map[string]tasks.TaskHandler
	components *redis.Components
	wg         sync.WaitGroup
	done       chan struct{}
	logger     *log.Logger
}

// Options configures the task handler.
type Options struct {
	MaxRetries     int
	RetryInterval  time.Duration
	TaskTypes      []string
	HandlerOptions []tasks.HandlerOption
	Logger         *log.Logger
}

// New creates a new task handler with the provided options.
func New(cfg *runner.Config, opts *Options) (*Handler, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if opts == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}

	logger := opts.Logger
	if logger == nil {
		logger = log.New(log.Writer(), "[TaskHandler] ", log.LstdFlags)
	}

	// Initialize Redis components
	components, err := redis.InitFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis components: %w", err)
	}

	// Initialize task handlers
	handlers := make(map[string]tasks.TaskHandler)
	for _, taskType := range opts.TaskTypes {
		handler := tasks.NewHandler(
			tasks.WithMaxRetries(opts.MaxRetries),
			tasks.WithRetryInterval(opts.RetryInterval),
		)
		handlers[taskType] = handler
	}

	// Initialize task mux
	mux := asynq.NewServeMux()
	for taskType, handler := range handlers {
		h := handler // Create a new variable to avoid closure issues
		mux.HandleFunc(taskType, func(ctx context.Context, task *asynq.Task) error {
			return h.ProcessTask(ctx, task)
		})
	}

	return &Handler{
		mux:        mux,
		handlers:   handlers,
		components: components,
		done:       make(chan struct{}),
		logger:     logger,
	}, nil
}

// Run starts the task handler and begins processing tasks.
func (h *Handler) Run(ctx context.Context, workers int) error {
	h.logger.Printf("Starting task handler with %d workers", workers)

	// Start health check goroutine
	h.wg.Add(1)
	go h.monitorHealth(ctx)

	// Start the server in a goroutine
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		if err := h.components.Server.Start(ctx, h.mux); err != nil {
			h.logger.Printf("Error running Redis server: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	return h.Close(ctx)
}

// Private method for continuous monitoring
func (h *Handler) monitorHealth(ctx context.Context) {
	defer h.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.done:
			return
		case <-ticker.C:
			if !h.components.Client.IsHealthy(ctx) {
				h.logger.Println("Warning: Redis client connection is not healthy")
			}
			if !h.components.Server.IsHealthy(ctx) {
				h.logger.Println("Warning: Redis server is not healthy")
			}
		}
	}
}

// Public method for single health check
func (h *Handler) MonitorHealth(ctx context.Context) error {
	// Perform single health check without WaitGroup
	if !h.components.Client.IsHealthy(ctx) {
		return fmt.Errorf("redis health check failed")
	}
	return nil
}

// Close gracefully shuts down the task handler.
func (h *Handler) Close(ctx context.Context) error {
	h.logger.Println("Shutting down task handler...")

	// Signal all goroutines to stop
	close(h.done)

	// Wait for all goroutines to finish
	h.wg.Wait()

	// Close Redis components
	if err := h.components.Close(ctx); err != nil {
		h.logger.Printf("Error closing Redis components: %v", err)
	}

	h.logger.Println("Task handler shutdown complete")
	return nil
}

// EnqueueTask enqueues a new task for processing.
func (h *Handler) EnqueueTask(ctx context.Context, taskType string, payload []byte, opts ...asynq.Option) error {
	return h.components.Client.EnqueueTask(ctx, taskType, payload, opts...)
}

// GetMux returns the task mux for server registration.
func (h *Handler) GetMux() *asynq.ServeMux {
	return h.mux
}

// GetHandler returns the task handler for a specific task type.
func (h *Handler) GetHandler(taskType string) (tasks.TaskHandler, bool) {
	handler, ok := h.handlers[taskType]
	return handler, ok
}

// RegisterHandler registers a new task handler for a specific task type.
func (h *Handler) RegisterHandler(taskType string, handler tasks.TaskHandler) error {
	if _, exists := h.handlers[taskType]; exists {
		return fmt.Errorf("handler already registered for task type: %s", taskType)
	}

	h.handlers[taskType] = handler
	h.mux.HandleFunc(taskType, func(ctx context.Context, task *asynq.Task) error {
		return handler.ProcessTask(ctx, task)
	})

	return nil
}

// UnregisterHandler removes a task handler for a specific task type.
func (h *Handler) UnregisterHandler(taskType string) {
	delete(h.handlers, taskType)
}

// GetRedisClient returns the Redis client instance.
func (h *Handler) GetRedisClient() *redis.Client {
	if h.components == nil {
		return nil
	}
	return h.components.Client
} 