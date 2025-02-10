// Package tasks provides Redis task handling functionality for asynchronous job processing.
package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// GeoFencing represents geographical boundaries and zoom levels for workflow execution
type GeoFencing struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	ZoomMin   int32   `json:"zoom_min"`
	ZoomMax   int32   `json:"zoom_max"`
}

func (g *GeoFencing) Validate() error {
	if g == nil {
		return nil // GeoFencing is optional
	}
	if g.Latitude < -90 || g.Latitude > 90 {
		return ErrInvalidPayload{Field: "latitude", Message: "must be between -90 and 90"}
	}
	if g.Longitude < -180 || g.Longitude > 180 {
		return ErrInvalidPayload{Field: "longitude", Message: "must be between -180 and 180"}
	}
	if g.ZoomMin < 0 {
		return ErrInvalidPayload{Field: "zoom_min", Message: "cannot be negative"}
	}
	if g.ZoomMax < g.ZoomMin {
		return ErrInvalidPayload{Field: "zoom_max", Message: "cannot be less than zoom_min"}
	}
	return nil
}

// WorkflowExecutionPayload represents the payload for workflow execution tasks
type WorkflowExecutionPayload struct {
	WorkflowID     uint64      `json:"workflow_id"`
	WorkspaceID    uint64      `json:"workspace_id"`
	WorkflowName   string      `json:"workflow_name"`
	CronExpression string      `json:"cron_expression"`
	GeoFencing     *GeoFencing `json:"geo_fencing,omitempty"`
	ExecutionTime  time.Time   `json:"execution_time"`
}

func (p *WorkflowExecutionPayload) Validate() error {
	if p.WorkflowID == 0 {
		return ErrInvalidPayload{Field: "workflow_id", Message: "workflow ID is required"}
	}
	if p.WorkflowName == "" {
		return ErrInvalidPayload{Field: "workflow_name", Message: "workflow name is required"}
	}
	if p.CronExpression == "" {
		return ErrInvalidPayload{Field: "cron_expression", Message: "cron expression is required"}
	}
	if p.ExecutionTime.IsZero() {
		return ErrInvalidPayload{Field: "execution_time", Message: "execution time is required"}
	}
	if err := p.GeoFencing.Validate(); err != nil {
		return err
	}
	return nil
}

// CreateWorkflowExecutionTask creates a new workflow execution task
func CreateWorkflowExecutionTask(workflowID uint64, workspaceID uint64, workflowName, cronExpression string, geoFencing *GeoFencing, executionTime time.Time) (*asynq.Task, error) {
	payload := &WorkflowExecutionPayload{
		WorkflowID:     workflowID,
		WorkspaceID:    workspaceID,
		WorkflowName:   workflowName,
		CronExpression: cronExpression,
		GeoFencing:     geoFencing,
		ExecutionTime:  executionTime,
	}

	t, err := NewTask(TypeWorkflowExecution.String(), payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeWorkflowExecution.String(), t), nil
}

// processWorkflowExecutionTask handles the execution of workflow tasks
func (h *Handler) processWorkflowExecutionTask(ctx context.Context, task *asynq.Task) error {
	// Check if context is cancelled
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	var payload WorkflowExecutionPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal workflow execution payload: %w", err)
	}

	if err := payload.Validate(); err != nil {
		return fmt.Errorf("invalid workflow execution payload: %w", err)
	}

	// // Register the task with the scheduler using the workflow's cron expression
	// entryID, err := h.scheduler.Register(
	// 	payload.CronExpression,
	// 	workflowTask,
	// 	opts...,
	// )
	// if err != nil {
	// 	return fmt.Errorf("failed to register workflow task: %w", err)
	// }

	// log.Printf("Registered workflow task with entry ID: %s for workflow ID: %d", entryID, payload.WorkflowID)

	// // query the workflow record from the database
	// workflow, err := h.db.GetScrapingWorkflow(ctx, payload.WorkflowID)
	// if err != nil {
	// 	return fmt.Errorf("failed to get workflow: %w", err)
	// }

	// // we need to update the workflow record with the entryID
	// workflow.ScheduledEntryId = entryID
	// workflow.Status = lead_scraper_servicev1.WorkflowStatus_WORKFLOW_STATUS_ACTIVE

	// // update the workflow record in the database
	// if _, err := h.db.UpdateScrapingWorkflow(ctx, workflow); err != nil {
	// 	return fmt.Errorf("failed to update workflow: %w", err)
	// }

	// // TODO: Store the entryID in the workflow record for future reference
	// // This would allow us to:
	// // 1. Track which workflows are currently scheduled
	// // 2. Unregister the task when the workflow is paused or deleted
	// // 3. Update the schedule if the workflow configuration changes

	return nil
}
