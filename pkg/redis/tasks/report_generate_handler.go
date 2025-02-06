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
	Format     string            `json:"format"`
	Filters    map[string]string `json:"filters,omitempty"`
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

// processReportGenerateTask handles the generation of reports
func (h *Handler) processReportGenerateTask(ctx context.Context, task *asynq.Task) error {
	// Check if context is cancelled
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	var payload ReportGeneratePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal report generate payload: %w", err)
	}

	if err := payload.Validate(); err != nil {
		return fmt.Errorf("invalid report generate payload: %w", err)
	}

	// TODO: Implement report generation logic
	// This would typically involve:
	// 1. Gathering data based on filters
	// 2. Processing and aggregating data
	// 3. Generating report in specified format
	// 4. Storing or sending the report
	// 5. Handling any errors during generation

	return nil
}

// CreateReportGenerateTask creates a new report generation task
func CreateReportGenerateTask(reportType string, filters map[string]string, format string) ([]byte, error) {
	if filters == nil {
		filters = make(map[string]string)
	}
	payload := &ReportGeneratePayload{
		ReportType: reportType,
		Format:     format,
		Filters:    filters,
	}
	return NewTask(TypeReportGenerate.String(), payload)
}
