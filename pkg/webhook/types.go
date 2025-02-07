package webhook

import (
	"context"
	"time"
)

// Record represents a single business record to be sent via webhook
type Record struct {
	// ID is a unique identifier for the record
	ID string `json:"id"`

	// Data contains the actual business data
	Data interface{} `json:"data"`

	// Timestamp is when the record was created
	Timestamp time.Time `json:"timestamp"`
}

// Batch represents a collection of records to be sent together
type Batch struct {
	// Records is the collection of records in this batch
	Records []Record `json:"records"`

	// BatchID is a unique identifier for this batch
	BatchID string `json:"batchId"`

	// Timestamp is when the batch was created
	Timestamp time.Time `json:"timestamp"`
}

// BatchProcessor is responsible for processing batches of records
type BatchProcessor interface {
	// ProcessBatch processes a batch of records
	ProcessBatch(ctx context.Context, batch *Batch) error
}

// BatchExporter is responsible for exporting batches to webhook endpoints
type BatchExporter interface {
	BatchProcessor
	// Export sends a batch to all registered webhook endpoints
	Export(ctx context.Context, batch *Batch) error
	// Shutdown gracefully shuts down the exporter
	Shutdown(ctx context.Context) error
}

// Client is the main interface for the webhook client
type Client interface {
	// Send adds a record to be processed
	Send(ctx context.Context, record Record) error

	// Flush forces processing of any pending records
	Flush(ctx context.Context) error

	// Shutdown gracefully shuts down the client
	Shutdown(ctx context.Context) error
} 