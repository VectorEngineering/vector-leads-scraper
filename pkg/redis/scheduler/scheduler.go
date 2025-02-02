// Package scheduler provides functionality for scheduling periodic tasks using asynq.
// It implements a centralized scheduler that ensures tasks are enqueued at regular intervals
// without duplication, even in a distributed environment.
//
// Features:
//   - Cron-style scheduling
//   - Interval-based scheduling
//   - Time zone support
//   - Error handling with custom handlers
//   - Thread-safe operations
//   - Graceful shutdown
//
// Example usage:
//
//	scheduler := scheduler.New(redisOpt, &scheduler.Config{
//	    Location: time.UTC,
//	    ErrorHandler: func(task *asynq.Task, opts []asynq.Option, err error) {
//	        log.Printf("Failed to enqueue task: %v", err)
//	    },
//	})
//
//	// Schedule a task to run every minute
//	entryID, err := scheduler.Register(
//	    "* * * * *",
//	    asynq.NewTask("email:digest", nil),
//	    scheduler.WithQueue("periodic"),
//	)
//
//	// Schedule a task to run every 30 seconds
//	entryID, err = scheduler.Register(
//	    "@every 30s",
//	    asynq.NewTask("metrics:collect", nil),
//	)
//
//	// Run the scheduler
//	if err := scheduler.Run(ctx); err != nil {
//	    log.Fatal(err)
//	}
package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hibiken/asynq"
)

// Config holds configuration for the scheduler.
type Config struct {
	// Location specifies the time zone to use for scheduling.
	// If nil, UTC will be used.
	Location *time.Location

	// ErrorHandler is called when a task fails to be enqueued.
	// If nil, errors will be logged to stderr.
	ErrorHandler func(task *asynq.Task, opts []asynq.Option, err error)

	// ShutdownTimeout specifies how long to wait for tasks to complete
	// during shutdown. Default is 30 seconds.
	ShutdownTimeout time.Duration
}

// DefaultConfig returns the default scheduler configuration.
func DefaultConfig() *Config {
	return &Config{
		Location:        time.UTC,
		ShutdownTimeout: 30 * time.Second,
		ErrorHandler: func(task *asynq.Task, opts []asynq.Option, err error) {
			log.Printf("Failed to enqueue task: %v", err)
		},
	}
}

// Scheduler manages periodic task scheduling.
type Scheduler struct {
	client     *asynq.Client
	scheduler  *asynq.Scheduler
	config     *Config
	mu         sync.RWMutex
	entries    map[string]Entry
	isRunning  bool
	cancelFunc context.CancelFunc
}

// Entry represents a scheduled task entry.
type Entry struct {
	ID       string
	Spec     string
	Task     *asynq.Task
	Options  []asynq.Option
	NextRun  time.Time
	PrevRun  time.Time
	Disabled bool
}

// New creates a new scheduler with the given Redis options and configuration.
func New(redisOpt asynq.RedisClientOpt, config *Config) *Scheduler {
	if config == nil {
		config = DefaultConfig()
	}
	if config.Location == nil {
		config.Location = time.UTC
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = DefaultConfig().ErrorHandler
	}
	if config.ShutdownTimeout == 0 {
		config.ShutdownTimeout = DefaultConfig().ShutdownTimeout
	}

	return &Scheduler{
		client: asynq.NewClient(redisOpt),
		scheduler: asynq.NewScheduler(
			redisOpt,
			&asynq.SchedulerOpts{
				Location:           config.Location,
				EnqueueErrorHandler: config.ErrorHandler,
			},
		),
		config:  config,
		entries: make(map[string]Entry),
	}
}

// Register adds a new periodic task to the scheduler.
// The spec can be either a cron expression (e.g., "* * * * *") or
// an interval expression (e.g., "@every 30s").
func (s *Scheduler) Register(spec string, task *asynq.Task, opts ...asynq.Option) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return "", fmt.Errorf("cannot register new tasks while scheduler is running")
	}

	entryID, err := s.scheduler.Register(spec, task, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to register task: %w", err)
	}

	s.entries[entryID] = Entry{
		ID:      entryID,
		Spec:    spec,
		Task:    task,
		Options: opts,
	}

	return entryID, nil
}

// Unregister removes a periodic task from the scheduler.
func (s *Scheduler) Unregister(entryID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("cannot unregister tasks while scheduler is running")
	}

	if err := s.scheduler.Unregister(entryID); err != nil {
		return fmt.Errorf("failed to unregister task: %w", err)
	}

	delete(s.entries, entryID)
	return nil
}

// Run starts the scheduler and blocks until the context is cancelled.
func (s *Scheduler) Run(ctx context.Context) error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}

	runCtx, cancel := context.WithCancel(ctx)
	s.cancelFunc = cancel
	s.isRunning = true
	s.mu.Unlock()

	// Run the scheduler and wait for context cancellation
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.scheduler.Run()
	}()

	select {
	case <-runCtx.Done():
		return runCtx.Err()
	case err := <-errCh:
		return fmt.Errorf("scheduler error: %w", err)
	}
}

// Stop gracefully stops the scheduler.
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	// Wait for scheduler to shutdown
	shutdownComplete := make(chan struct{})
	var shutdownErr error
	go func() {
		s.scheduler.Shutdown()
		close(shutdownComplete)
	}()

	// Wait for shutdown or timeout
	select {
	case <-shutdownComplete:
		if shutdownErr != nil {
			return fmt.Errorf("error during shutdown: %w", shutdownErr)
		}
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout after %v", s.config.ShutdownTimeout)
	}

	// Close the client after scheduler shutdown
	if err := s.client.Close(); err != nil {
		return fmt.Errorf("error closing client: %w", err)
	}

	s.isRunning = false
	return nil
}

// GetEntries returns all registered task entries.
func (s *Scheduler) GetEntries() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]Entry, 0, len(s.entries))
	for _, entry := range s.entries {
		entries = append(entries, entry)
	}
	return entries
}

// WithQueue returns an option to specify the queue for a task.
func WithQueue(queue string) asynq.Option {
	return asynq.Queue(queue)
}

// WithTimeout returns an option to set task timeout.
func WithTimeout(d time.Duration) asynq.Option {
	return asynq.Timeout(d)
}

// WithRetry returns an option to set maximum retries.
func WithRetry(max int) asynq.Option {
	return asynq.MaxRetry(max)
}

// WithDeadline returns an option to set task deadline.
func WithDeadline(t time.Time) asynq.Option {
	return asynq.Deadline(t)
} 