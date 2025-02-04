// Package tasks provides Redis task handling functionality for asynchronous job processing.
package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// ReportGeneratePayload represents the payload for report generation tasks
type ReportGeneratePayload struct {
	ReportType string            `json:"report_type"`
	Filters    map[string]string `json:"filters"`
	Format     string            `json:"format"`
}

func (p *ReportGeneratePayload) Validate() error {
	if p.ReportType == "" {
		return ErrInvalidPayload{Field: "report_type", Message: "report type is required"}
	}
	if p.Format == "" {
		return ErrInvalidPayload{Field: "format", Message: "format is required"}
	}
	return nil
}

// processReportGenerateTask handles the generation of reports from lead data
func (h *Handler) processReportGenerateTask(ctx context.Context, task *asynq.Task) error {
	var payload ReportGeneratePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal report generate payload: %w", err)
	}

	if err := payload.Validate(); err != nil {
		return fmt.Errorf("invalid report generate payload: %w", err)
	}

	// TODO: Implement report generation logic
	// This would typically involve:
	// 1. Querying data based on filters
	// 2. Generating report in specified format
	// 3. Storing report output
	// 4. Notifying completion

	return nil
}

// CreateReportGenerateTask creates a new report generation task
func CreateReportGenerateTask(reportType string, filters map[string]string, format string) ([]byte, error) {
	payload := &ReportGeneratePayload{
		ReportType: reportType,
		Filters:    filters,
		Format:     format,
	}
	return NewTask(TypeReportGenerate.String(), payload)
}
