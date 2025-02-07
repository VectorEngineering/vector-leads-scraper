// Package taskhandler provides task handling functionality for Redis-backed operations.
// It implements a robust task processing system that manages the lifecycle of asynchronous
// tasks using Redis as the backend store. The package supports task scheduling,
// monitoring, and graceful shutdown capabilities.
//
// Example usage of the task handler:
//
//	// Create configuration
//	cfg := &runner.Config{
//		RedisAddr: "localhost:6379",
//		Queues: map[string]int{
//			"critical": 6,
//			"default": 3,
//			"low":     1,
//		},
//	}
//
//	// Configure task handler options
//	opts := &taskhandler.Options{
//		MaxRetries:    3,
//		RetryInterval: 10 * time.Second,
//		TaskTypes: []string{
//			tasks.TypeEmailExtract.String(),
//			tasks.TypeScrapeGMaps.String(),
//		},
//		HandlerOptions: []tasks.HandlerOption{
//			tasks.WithConcurrency(10),
//			tasks.WithTaskTimeout(5 * time.Minute),
//		},
//		Logger: log.New(os.Stdout, "[TaskHandler] ", log.LstdFlags),
//	}
//
//	// Create and start the handler
//	handler, err := taskhandler.New(cfg, opts)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Start processing with context
//	ctx := context.Background()
//	if err := handler.Run(ctx, 10); err != nil {
//		log.Fatal(err)
//	}
//
// Example of task operations:
//
//	// Enqueue a new task
//	payload := &tasks.EmailPayload{
//		URL:      "https://example.com",
//		MaxDepth: 2,
//	}
//	data, _ := json.Marshal(payload)
//	err := handler.EnqueueTask(ctx, tasks.TypeEmailExtract,
//		data,
//		asynq.Queue("default"),
//		asynq.MaxRetry(3),
//	)
//
//	// Process task operations
//	taskId := "task123"
//	queue := "default"
//
//	// Run a task
//	err = handler.ProcessTask(ctx, taskId, queue, taskhandler.RunTask)
//
//	// Cancel a running task
//	err = handler.ProcessTask(ctx, taskId, queue, taskhandler.CancelTaskProcessing)
//
//	// Archive a completed task
//	err = handler.ProcessTask(ctx, taskId, queue, taskhandler.ArchiveTask)
//
//	// Get task details
//	info, err := handler.GetTaskDetails(ctx, taskId, queue)
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Printf("Task Status: %s", info.Status)
//	log.Printf("Progress: %d/%d", info.Progress, info.MaxProgress)
package taskhandler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Vector/vector-leads-scraper/internal/taskhandler/tasks"
	"github.com/Vector/vector-leads-scraper/pkg/redis"
	"github.com/Vector/vector-leads-scraper/runner"
	"github.com/hibiken/asynq"
)

// Handler manages task processing for Redis operations.
// It provides functionality for task enqueueing, processing, monitoring,
// and lifecycle management.
type Handler struct {
	mux        *asynq.ServeMux              // ServeMux for routing tasks to handlers
	handlers   map[string]tasks.TaskHandler // Map of task type to handlers
	components *redis.Components            // Redis components for task operations
	wg         sync.WaitGroup               // WaitGroup for graceful shutdown
	done       chan struct{}                // Channel for shutdown signaling
	logger     *log.Logger                  // Logger for handler operations
}

// Options configures the task handler.
// It provides configuration options for retry behavior, task types,
// and handler-specific settings.
type Options struct {
	MaxRetries     int                   // Maximum number of retry attempts
	RetryInterval  time.Duration         // Time to wait between retries
	TaskTypes      []string              // List of supported task types
	HandlerOptions []tasks.HandlerOption // Options for task handlers
	Logger         *log.Logger           // Custom logger for the handler
}

// TaskProcessingOperation represents different operations that can be
// performed on a task.
type TaskProcessingOperation string

const (
	// CancelTaskProcessing cancels a running task
	CancelTaskProcessing TaskProcessingOperation = "cancel"
	// ArchiveTask archives a completed task
	ArchiveTask TaskProcessingOperation = "archive"
	// RunTask starts task execution
	RunTask TaskProcessingOperation = "run"
	// DeleteTask removes a task from the queue
	DeleteTask TaskProcessingOperation = "delete"
)

// String returns the string representation of the operation.
func (o TaskProcessingOperation) String() string {
	return string(o)
}

// New creates a new task handler with the provided options.
// It initializes Redis components and sets up task handlers based on
// the provided configuration.
//
// Example:
//
//	handler, err := taskhandler.New(
//		&runner.Config{
//			RedisAddr: "localhost:6379",
//		},
//		&taskhandler.Options{
//			MaxRetries:    3,
//			RetryInterval: 10 * time.Second,
//			TaskTypes: []string{
//				tasks.TypeEmailExtract.String(),
//			},
//		},
//	)
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

	// Create the handler instance with empty handlers map
	h := &Handler{
		mux:        asynq.NewServeMux(),
		handlers:   make(map[string]tasks.TaskHandler),
		components: components,
		done:       make(chan struct{}),
		logger:     logger,
	}

	// Register task handlers
	for _, taskType := range opts.TaskTypes {
		if err := h.RegisterHandler(taskType, tasks.NewHandler(
			tasks.WithMaxRetries(opts.MaxRetries),
			tasks.WithRetryInterval(opts.RetryInterval),
		)); err != nil {
			// Clean up on error
			h.Close(context.Background())
			return nil, fmt.Errorf("failed to register handler for task type %s: %w", taskType, err)
		}
	}

	return h, nil
}

// Run starts the task handler and begins processing tasks.
// It initializes health monitoring and starts the server with the
// specified number of worker goroutines.
//
// Example:
//
//	ctx := context.Background()
//	if err := handler.Run(ctx, 10); err != nil {
//		log.Fatal(err)
//	}
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

// GetMux returns the task mux for server registration.
// The returned ServeMux can be used to register additional handlers
// or to configure the asynq server.
//
// Example:
//
//	mux := handler.GetMux()
//	mux.HandleFunc("custom:task", func(ctx context.Context, task *asynq.Task) error {
//		// Custom task handling logic
//		return nil
//	})
func (h *Handler) GetMux() *asynq.ServeMux {
	return h.mux
}

// GetHandler returns the task handler for a specific task type.
// It returns the handler and a boolean indicating whether the handler exists.
//
// Example:
//
//	if handler, exists := h.GetHandler(tasks.TypeEmailExtract.String()); exists {
//		// Use the handler
//		err := handler.ProcessTask(ctx, task)
//	} else {
//		log.Printf("No handler found for task type: %s", taskType)
//	}
func (h *Handler) GetHandler(taskType string) (tasks.TaskHandler, bool) {
	handler, ok := h.handlers[taskType]
	return handler, ok
}

// RegisterHandler registers a new task handler for a specific task type.
// It returns an error if a handler is already registered for the task type.
//
// Example:
//
//	customHandler := &CustomHandler{
//		logger: log.New(os.Stdout, "[Custom] ", log.LstdFlags),
//	}
//	err := handler.RegisterHandler("custom:task", customHandler)
//	if err != nil {
//		log.Printf("Failed to register handler: %v", err)
//	}
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
// This method is safe to call even if no handler exists for the task type.
//
// Example:
//
//	handler.UnregisterHandler("custom:task")
//	log.Printf("Unregistered handler for custom:task")
func (h *Handler) UnregisterHandler(taskType string) {
	delete(h.handlers, taskType)
}

// GetRedisClient returns the Redis client instance.
// Returns nil if the Redis components are not initialized.
//
// Example:
//
//	if client := handler.GetRedisClient(); client != nil {
//		// Use the Redis client
//		err := client.Ping(ctx)
//		if err != nil {
//			log.Printf("Redis connection error: %v", err)
//		}
//	}
func (h *Handler) GetRedisClient() *redis.Client {
	if h.components == nil {
		return nil
	}
	return h.components.Client
}

// monitorHealth continuously monitors the health of Redis connections.
// It runs as a goroutine and checks the health status every 30 seconds.
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

// MonitorHealth performs a single health check of the Redis connections.
// This is a public method that can be used for on-demand health checks.
//
// Example:
//
//	if err := handler.MonitorHealth(ctx); err != nil {
//		log.Printf("Health check failed: %v", err)
//		// Take appropriate action (e.g., restart connections)
//	}
func (h *Handler) MonitorHealth(ctx context.Context) error {
	if !h.components.Client.IsHealthy(ctx) {
		return fmt.Errorf("redis health check failed")
	}
	return nil
}

// Close gracefully shuts down the task handler.
// It ensures all goroutines are stopped and Redis connections are closed properly.
//
// Example:
//
//	ctx := context.Background()
//	if err := handler.Close(ctx); err != nil {
//		log.Printf("Error during shutdown: %v", err)
//	}
func (h *Handler) Close(ctx context.Context) error {
	h.logger.Println("Shutting down task handler...")

	// Use sync.Once to ensure the channel is closed only once
	var once sync.Once
	once.Do(func() {
		// Signal all goroutines to stop
		close(h.done)
	})

	// Wait for all goroutines to finish
	h.wg.Wait()

	// Close Redis components
	if err := h.components.Close(ctx); err != nil {
		h.logger.Printf("Error closing Redis components: %v", err)
		return fmt.Errorf("error closing Redis components: %w", err)
	}

	h.logger.Println("Task handler shutdown complete")
	return nil
}

// EnqueueTask enqueues a new task for processing.
// It validates the task type and adds the task to the specified queue
// with the provided options.
//
// Example:
//
//	payload := &tasks.EmailPayload{
//		URL: "https://example.com",
//	}
//	data, _ := json.Marshal(payload)
//	err := handler.EnqueueTask(ctx,
//		tasks.TypeEmailExtract,
//		data,
//		asynq.Queue("default"),
//		asynq.MaxRetry(3),
//	)
func (h *Handler) EnqueueTask(ctx context.Context, taskType tasks.TaskType, payload []byte, opts ...asynq.Option) error {
	// validate taskType
	if taskType == "" {
		return fmt.Errorf("taskType cannot be empty")
	}

	return h.components.Client.EnqueueTask(ctx, taskType.String(), payload, opts...)
}

// ProcessTask performs the specified operation on a task.
// It supports running, canceling, archiving, and deleting tasks.
//
// Example:
//
//	// Run a task
//	err := handler.ProcessTask(ctx, "task123", "default", taskhandler.RunTask)
//
//	// Cancel a running task
//	err = handler.ProcessTask(ctx, "task123", "default", taskhandler.CancelTaskProcessing)
//
//	// Archive a completed task
//	err = handler.ProcessTask(ctx, "task123", "default", taskhandler.ArchiveTask)
func (h *Handler) ProcessTask(ctx context.Context, taskId string, queue string, operation TaskProcessingOperation) error {
	// validate taskId and queue
	if taskId == "" {
		return fmt.Errorf("taskId cannot be empty")
	}
	if queue == "" {
		return fmt.Errorf("queue cannot be empty")
	}
	switch operation {
	case RunTask:
		return h.components.Client.RunTask(ctx, taskId, queue)
	case CancelTaskProcessing:
		return h.components.Client.CancelProcessing(ctx, taskId)
	case ArchiveTask:
		return h.components.Client.ArchiveAndStoreTask(ctx, queue, taskId)
	case DeleteTask:
		return h.components.Client.DeleteTask(ctx, queue, taskId)
	default:
		return fmt.Errorf("invalid operation: %s", operation)
	}
}

// GetTaskDetails retrieves detailed information about a specific task.
// It returns task status, progress, and configuration information.
//
// Example:
//
//	info, err := handler.GetTaskDetails(ctx, "task123", "default")
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Printf("Task Status: %s", info.Status)
//	log.Printf("Progress: %d/%d", info.Progress, info.MaxProgress)
func (h *Handler) GetTaskDetails(ctx context.Context, taskId string, queue string) (*asynq.TaskInfo, error) {
	// validate taskId and queue
	if taskId == "" {
		return nil, fmt.Errorf("taskId cannot be empty")
	}
	if queue == "" {
		return nil, fmt.Errorf("queue cannot be empty")
	}

	return h.components.Client.GetTaskInfo(ctx, taskId, queue)
}
