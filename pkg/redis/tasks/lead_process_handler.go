// Package tasks provides Redis task handling functionality for asynchronous job processing.
package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// LeadProcessPayload represents the payload for lead processing tasks
type LeadProcessPayload struct {
	LeadID    string            `json:"lead_id"`
	Source    string            `json:"source"`
	BatchSize int               `json:"batch_size"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

func (p *LeadProcessPayload) Validate() error {
	if p.LeadID == "" {
		return ErrInvalidPayload{Field: "lead_id", Message: "lead ID is required"}
	}
	if p.Source == "" {
		return ErrInvalidPayload{Field: "source", Message: "source is required"}
	}
	if p.BatchSize <= 0 {
		p.BatchSize = 1000 // Set default batch size
	}
	return nil
}

// processLeadProcessTask handles the processing of lead data
// It processes leads in batches as specified in the payload
func (h *Handler) processLeadProcessTask(ctx context.Context, task *asynq.Task) error {
	// Check if context is cancelled
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	var payload LeadProcessPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal lead process payload: %w", err)
	}

	if err := payload.Validate(); err != nil {
		return fmt.Errorf("invalid lead process payload: %w", err)
	}

	// TODO: Implement lead processing logic
	// This would typically involve:
	// 1. Retrieving the lead data from the source
	// 2. Processing the data in batches
	// 3. Applying transformations and validations
	// 4. Storing the processed data
	// 5. Handling any errors during processing

	return nil
}

// CreateLeadProcessTask creates a new lead processing task
func CreateLeadProcessTask(leadID string, source string, batchSize int, metadata map[string]string) ([]byte, error) {
	payload := &LeadProcessPayload{
		LeadID:    leadID,
		Source:    source,
		BatchSize: batchSize,
		Metadata:  metadata,
	}
	return NewTask(TypeLeadProcess.String(), payload)
}
