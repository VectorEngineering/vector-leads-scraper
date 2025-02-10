package webhook

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/multierr"
)

// client implements the Client interface
type client struct {
	config    Config
	exporter  BatchExporter
	processor BatchProcessor

	// Batch management
	currentBatch *Batch
	batchMu      sync.Mutex
	batchTimer   *time.Timer
	done         chan struct{}

	// Shutdown management
	shutdownOnce sync.Once
	shutdownErr  error
}

// NewClient creates a new webhook client
func NewClient(cfg Config) (Client, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	c := &client{
		config:   cfg,
		exporter: newHTTPExporter(cfg),
		done:     make(chan struct{}),
	}

	c.processor = c.exporter // Use exporter as the processor by default
	c.currentBatch = newBatch()
	c.batchTimer = time.NewTimer(cfg.BatchTimeout)

	go c.processingLoop()

	return c, nil
}

// validateConfig validates the configuration
func validateConfig(cfg Config) error {
	if cfg.MaxBatchSize < 1 {
		return fmt.Errorf("MaxBatchSize must be greater than 0")
	}
	if cfg.BatchTimeout < time.Second {
		return fmt.Errorf("BatchTimeout must be at least 1 second")
	}
	if len(cfg.Endpoints) == 0 {
		return fmt.Errorf("at least one endpoint must be configured")
	}
	if cfg.RetryConfig.MaxRetries < 0 {
		return fmt.Errorf("MaxRetries must be non-negative")
	}
	if cfg.RetryConfig.InitialBackoff < time.Millisecond {
		return fmt.Errorf("InitialBackoff must be at least 1ms")
	}
	if cfg.RetryConfig.MaxBackoff < cfg.RetryConfig.InitialBackoff {
		return fmt.Errorf("MaxBackoff must be greater than or equal to InitialBackoff")
	}
	if cfg.RetryConfig.BackoffMultiplier < 1 {
		return fmt.Errorf("BackoffMultiplier must be greater than or equal to 1")
	}
	return nil
}

// newBatch creates a new batch with a unique ID
func newBatch() *Batch {
	return &Batch{
		Records:   make([]Record, 0),
		BatchID:   uuid.New().String(),
		Timestamp: time.Now(),
	}
}

// Send implements Client.Send
func (c *client) Send(ctx context.Context, record Record) error {
	select {
	case <-c.done:
		return fmt.Errorf("client is shut down")
	default:
	}

	c.batchMu.Lock()
	defer c.batchMu.Unlock()

	c.currentBatch.Records = append(c.currentBatch.Records, record)

	if len(c.currentBatch.Records) >= c.config.MaxBatchSize {
		return c.flushLocked(ctx)
	}

	return nil
}

// Flush implements Client.Flush
func (c *client) Flush(ctx context.Context) error {
	c.batchMu.Lock()
	defer c.batchMu.Unlock()

	return c.flushLocked(ctx)
}

// flushLocked flushes the current batch while holding the lock
func (c *client) flushLocked(ctx context.Context) error {
	if len(c.currentBatch.Records) == 0 {
		return nil
	}

	batch := c.currentBatch
	c.currentBatch = newBatch()

	// Reset the timer
	if !c.batchTimer.Stop() {
		select {
		case <-c.batchTimer.C:
		default:
		}
	}
	c.batchTimer.Reset(c.config.BatchTimeout)

	// Process the batch
	if err := c.processor.ProcessBatch(ctx, batch); err != nil {
		return fmt.Errorf("failed to process batch: %w", err)
	}

	return nil
}

// processingLoop runs the main processing loop
func (c *client) processingLoop() {
	for {
		select {
		case <-c.done:
			return
		case <-c.batchTimer.C:
			c.batchMu.Lock()
			// Ignore error here as we can't do much about it
			_ = c.flushLocked(context.Background())
			c.batchMu.Unlock()
		}
	}
}

// Shutdown implements Client.Shutdown
func (c *client) Shutdown(ctx context.Context) error {
	c.shutdownOnce.Do(func() {
		// Signal processing loop to stop
		close(c.done)

		// Stop the timer
		c.batchTimer.Stop()

		// Flush any remaining records
		c.batchMu.Lock()
		if err := c.flushLocked(ctx); err != nil {
			c.shutdownErr = multierr.Append(c.shutdownErr, fmt.Errorf("failed to flush during shutdown: %w", err))
		}
		c.batchMu.Unlock()

		// Shutdown the exporter
		if err := c.exporter.Shutdown(ctx); err != nil {
			c.shutdownErr = multierr.Append(c.shutdownErr, fmt.Errorf("failed to shutdown exporter: %w", err))
		}
	})

	return c.shutdownErr
}
