// Package tasks provides Redis task handling functionality for asynchronous job processing.
// It implements a robust task handling system using asynq for managing various types
// of background jobs including web scraping and email extraction.
//
// Example usage of the task handling system:
//
//	// Create a handler with custom configuration
//	handler := tasks.NewHandler(
//		tasks.WithMaxRetries(5),
//		tasks.WithRetryInterval(10 * time.Second),
//		tasks.WithTaskTimeout(5 * time.Minute),
//		tasks.WithConcurrency(10),
//		tasks.WithProxies([]string{"http://proxy1.example.com"}),
//	)
//
//	// Create an asynq server
//	srv := asynq.NewServer(
//		asynq.RedisClientOpt{Addr: "localhost:6379"},
//		asynq.Config{
//			Concurrency: 10,
//			Queues: map[string]int{
//				"critical": 6,
//				"default": 3,
//				"low":     1,
//			},
//		},
//	)
//
//	// Create a mux and register task handlers
//	mux := asynq.NewServeMux()
//	mux.Handle(tasks.TypeEmailExtract.String(), handler)
//	mux.Handle(tasks.TypeScrapeGMaps.String(), handler)
//	mux.Handle(tasks.TypeHealthCheck.String(), handler)
//
//	// Run the server
//	if err := srv.Run(mux); err != nil {
//		log.Fatal(err)
//	}
//
// Example of enqueueing tasks:
//
//	// Create a client
//	client := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
//	defer client.Close()
//
//	// Create and enqueue an email extraction task
//	emailTask, _ := tasks.CreateEmailTask("example.com", 2, "")
//	info, err := client.Enqueue(emailTask, asynq.Queue("default"))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create and enqueue a Google Maps scraping task with more options
//	gmapsTask, _ := tasks.CreateScrapeGMapsTask("restaurant in new york")
//	info, err = client.Enqueue(gmapsTask,
//		asynq.Queue("critical"),
//		asynq.MaxRetry(3),
//		asynq.Timeout(10*time.Minute),
//		asynq.Deadline(time.Now().Add(24*time.Hour)),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/Vector/vector-leads-scraper/internal/database"
	"github.com/Vector/vector-leads-scraper/pkg/redis/scheduler"
	"github.com/hibiken/asynq"
)

// TaskHandler defines the interface for processing Redis tasks.
// Implementations of this interface should be able to handle
// different types of tasks and manage their execution lifecycle.
//
// Example implementation:
//
//	type CustomHandler struct {
//		logger *log.Logger
//		db     *sql.DB
//	}
//
//	func (h *CustomHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
//		switch task.Type() {
//		case "custom:task":
//			return h.processCustomTask(ctx, task)
//		default:
//			return fmt.Errorf("unsupported task type: %s", task.Type())
//		}
//	}
//
//	func (h *CustomHandler) processCustomTask(ctx context.Context, task *asynq.Task) error {
//		var payload struct {
//			UserID string `json:"user_id"`
//			Action string `json:"action"`
//		}
//		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
//			return fmt.Errorf("failed to unmarshal payload: %w", err)
//		}
//		// Process the task...
//		return nil
//	}
type TaskHandler interface {
	ProcessTask(ctx context.Context, task *asynq.Task) error
}

// Handler implements the TaskHandler interface and provides configuration
// options for task processing, including retry behavior, timeouts,
// and scraping-specific settings.
//
// Example usage with all configuration options:
//
//	handler := tasks.NewHandler(
//		// Configure retry behavior
//		tasks.WithMaxRetries(5),
//		tasks.WithRetryInterval(10 * time.Second),
//
//		// Set timeouts and concurrency
//		tasks.WithTaskTimeout(5 * time.Minute),
//		tasks.WithConcurrency(10),
//
//		// Configure data storage
//		tasks.WithDataFolder("/path/to/data"),
//
//		// Set up scraping options
//		tasks.WithProxies([]string{
//			"http://proxy1.example.com",
//			"http://proxy2.example.com",
//		}),
//		tasks.WithDisablePageReuse(true),
//	)
//
//	// Use the handler with asynq
//	mux := asynq.NewServeMux()
//	mux.Handle(tasks.TypeEmailExtract.String(), handler)
//
//	// Process tasks directly
//	ctx := context.Background()
//	task, _ := tasks.CreateEmailTask("example.com", 2, "")
//	if err := handler.ProcessTask(ctx, task); err != nil {
//		log.Fatal(err)
//	}
type Handler struct {
	maxRetries    int                         // Maximum number of retry attempts for failed tasks
	retryInterval time.Duration               // Time to wait between retry attempts
	taskTimeout   time.Duration               // Maximum time allowed for task execution
	dataFolder    string                      // Directory for storing task results and data
	concurrency   int                         // Number of concurrent scraping operations
	proxies       []string                    // List of proxy servers for scraping
	disableReuse  bool                        // Flag to disable page reuse in scraping
	scheduler     *scheduler.Scheduler        // Scheduler for periodic tasks
	db            database.DatabaseOperations // Database for task operations
}

// HandlerOption defines a function type for configuring Handler instances.
// It follows the functional options pattern for flexible configuration.
type HandlerOption func(*Handler)

// WithMaxRetries creates an option to set the maximum number of retry attempts
// for failed tasks.
//
// Parameters:
//   - retries: The maximum number of retries
//
// Returns:
//   - HandlerOption: A function that sets the max retries configuration
func WithMaxRetries(retries int) HandlerOption {
	return func(h *Handler) {
		h.maxRetries = retries
	}
}

// WithRetryInterval creates an option to set the duration between retry attempts
// for failed tasks.
//
// Parameters:
//   - interval: The time duration to wait between retries
//
// Returns:
//   - HandlerOption: A function that sets the retry interval configuration
func WithRetryInterval(interval time.Duration) HandlerOption {
	return func(h *Handler) {
		h.retryInterval = interval
	}
}

// WithTaskTimeout creates an option to set the maximum duration allowed
// for task execution before timing out.
//
// Parameters:
//   - timeout: The maximum duration for task execution
//
// Returns:
//   - HandlerOption: A function that sets the task timeout configuration
func WithTaskTimeout(timeout time.Duration) HandlerOption {
	return func(h *Handler) {
		h.taskTimeout = timeout
	}
}

// WithDataFolder creates an option to set the directory path
// where task results and data will be stored.
//
// Parameters:
//   - folder: The path to the data storage directory
//
// Returns:
//   - HandlerOption: A function that sets the data folder configuration
func WithDataFolder(folder string) HandlerOption {
	return func(h *Handler) {
		h.dataFolder = folder
	}
}

// WithConcurrency creates an option to set the number of concurrent
// scraping operations allowed.
//
// Parameters:
//   - n: The number of concurrent operations
//
// Returns:
//   - HandlerOption: A function that sets the concurrency configuration
func WithConcurrency(n int) HandlerOption {
	return func(h *Handler) {
		h.concurrency = n
	}
}

// WithProxies creates an option to set the list of proxy servers
// to be used for scraping operations.
//
// Parameters:
//   - proxies: A slice of proxy server URLs
//
// Returns:
//   - HandlerOption: A function that sets the proxies configuration
func WithProxies(proxies []string) HandlerOption {
	return func(h *Handler) {
		h.proxies = proxies
	}
}

// WithDisablePageReuse creates an option to control whether pages
// should be reused during scraping operations.
//
// Parameters:
//   - disable: Boolean flag to enable/disable page reuse
//
// Returns:
//   - HandlerOption: A function that sets the page reuse configuration
func WithDisablePageReuse(disable bool) HandlerOption {
	return func(h *Handler) {
		h.disableReuse = disable
	}
}

func WithDatabase(db database.DatabaseOperations) HandlerOption {
	return func(h *Handler) {
		h.db = db
	}
}

// NewHandler creates a new task handler with the specified options.
// It initializes a Handler with default values and applies any provided
// options to customize the configuration.
//
// Parameters:
//   - scheduler: The scheduler instance for managing periodic tasks
//   - opts: Variable number of HandlerOption functions to configure the handler
//
// Returns:
//   - *Handler: A new Handler instance with the specified configuration
func NewHandler(scheduler *scheduler.Scheduler, opts ...HandlerOption) *Handler {
	h := &Handler{
		maxRetries:    3,
		retryInterval: 5 * time.Second,
		taskTimeout:   30 * time.Second,
		concurrency:   2,
		scheduler:     scheduler,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// ProcessTask processes a task based on its type. It applies the configured
// timeout and routes the task to the appropriate handler based on its type.
//
// Supported task types:
//   - TypeScrapeGMaps: Google Maps scraping tasks
//   - TypeEmailExtract: Email extraction tasks
//   - TypeHealthCheck: Health check tasks
//   - TypeConnectionTest: Connection test tasks
//
// Parameters:
//   - ctx: Context for the operation
//   - task: The task to be processed
//
// Returns:
//   - error: Any error encountered during task processing
func (h *Handler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	ctx, cancel := context.WithTimeout(ctx, h.taskTimeout)
	defer cancel()

	switch task.Type() {
	case TypeScrapeGMaps.String():
		return h.processScrapeTask(ctx, task)
	case TypeEmailExtract.String():
		return h.processEmailTask(ctx, task)
	case TypeHealthCheck.String():
		return nil // Health check task always succeeds
	case TypeConnectionTest.String():
		return nil // Connection test task always succeeds
	case TypeLeadProcess.String():
		return h.processLeadProcessTask(ctx, task)
	case TypeLeadValidate.String():
		return h.processLeadValidateTask(ctx, task)
	case TypeLeadEnrich.String():
		return h.processLeadEnrichTask(ctx, task)
	case TypeReportGenerate.String():
		return h.processReportGenerateTask(ctx, task)
	case TypeDataExport.String():
		return h.processDataExportTask(ctx, task)
	case TypeDataImport.String():
		return h.processDataImportTask(ctx, task)
	case TypeDataCleanup.String():
		return h.processDataCleanupTask(ctx, task)
	case TypeWorkflowExecution.String():
		return h.processWorkflowExecutionTask(ctx, task)
	default:
		return fmt.Errorf("unknown task type: %s", task.Type())
	}
}
