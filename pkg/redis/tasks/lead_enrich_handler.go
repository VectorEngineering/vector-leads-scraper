// Package tasks provides Redis task handling functionality for asynchronous job processing.
package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// LeadEnrichPayload represents the payload for lead enrichment tasks
type LeadEnrichPayload struct {
	LeadID     string   `json:"lead_id"`
	EnrichType []string `json:"enrich_type"`
}

func (p *LeadEnrichPayload) Validate() error {
	if p.LeadID == "" {
		return ErrInvalidPayload{Field: "lead_id", Message: "lead ID is required"}
	}
	if len(p.EnrichType) == 0 {
		return ErrInvalidPayload{Field: "enrich_type", Message: "at least one enrich type is required"}
	}
	return nil
}

// processLeadEnrichTask handles the enrichment of lead data with additional information
func (h *Handler) processLeadEnrichTask(ctx context.Context, task *asynq.Task) error {
	var payload LeadEnrichPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal lead enrich payload: %w", err)
	}

	if err := payload.Validate(); err != nil {
		return fmt.Errorf("invalid lead enrich payload: %w", err)
	}

	// TODO: Implement lead enrichment logic
	// This would typically involve:
	// 1. Fetching additional data from external sources
	// 2. Updating lead information
	// 3. Handling rate limits and API quotas
	// 4. Logging enrichment results

	return nil
}

// CreateLeadEnrichTask creates a new lead enrichment task
func CreateLeadEnrichTask(leadID string, enrichTypes []string) ([]byte, error) {
	payload := &LeadEnrichPayload{
		LeadID:     leadID,
		EnrichType: enrichTypes,
	}
	return NewTask(TypeLeadEnrich.String(), payload)
}
