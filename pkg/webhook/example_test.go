package webhook_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Vector/vector-leads-scraper/pkg/webhook"
)

func Example() {
	// Create a webhook client with custom configuration
	cfg := webhook.Config{
		MaxBatchSize:         100,
		BatchTimeout:         30 * time.Second,
		CompressionThreshold: 1024 * 1024, // 1MB
		RetryConfig: webhook.RetryConfig{
			MaxRetries:        3,
			InitialBackoff:    time.Second,
			MaxBackoff:        30 * time.Second,
			BackoffMultiplier: 2.0,
		},
		Endpoints: []string{
			"https://api.example.com/webhooks/endpoint1",
			"https://api.example.com/webhooks/endpoint2",
		},
	}

	client, err := webhook.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create webhook client: %v", err)
	}
	defer client.Shutdown(context.Background())

	// Create a record
	record := webhook.Record{
		ID:   "123",
		Data: map[string]interface{}{
			"name":    "Example Business",
			"address": "123 Main St",
			"phone":   "555-0123",
		},
		Timestamp: time.Now(),
	}

	// Send the record
	ctx := context.Background()
	if err := client.Send(ctx, record); err != nil {
		log.Printf("Failed to send record: %v", err)
	}

	// Records will be automatically batched and sent based on size or time
	// You can also manually flush the current batch
	if err := client.Flush(ctx); err != nil {
		log.Printf("Failed to flush records: %v", err)
	}
}

// customProcessor is an example implementation of BatchProcessor for testing
type customProcessor struct {
	records []webhook.Record
}

// ProcessBatch implements webhook.BatchProcessor
func (p *customProcessor) ProcessBatch(ctx context.Context, batch *webhook.Batch) error {
	p.records = append(p.records, batch.Records...)
	fmt.Printf("Processed %d records\n", len(batch.Records))
	return nil
}

func Example_withCustomProcessor() {
	// Create a mock HTTP server for demonstration
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create the webhook client
	cfg := webhook.DefaultConfig()
	cfg.Endpoints = []string{"http://localhost:8080/webhook"}

	client, err := webhook.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create webhook client: %v", err)
	}
	defer client.Shutdown(context.Background())

	// Send some test records
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		record := webhook.Record{
			ID:        fmt.Sprintf("test-%d", i),
			Data:      fmt.Sprintf("test data %d", i),
			Timestamp: time.Now(),
		}

		if err := client.Send(ctx, record); err != nil {
			log.Printf("Failed to send record: %v", err)
		}
	}

	// Force flush any remaining records
	if err := client.Flush(ctx); err != nil {
		log.Printf("Failed to flush records: %v", err)
	}
} 