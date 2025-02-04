package database

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

// validateCronExpression validates a cron expression
func validateCronExpression(cron string) error {
	// Split the cron expression into its components
	fields := strings.Fields(cron)
	if len(fields) != 5 {
		return fmt.Errorf("invalid cron expression: must have 5 fields")
	}

	// Define valid ranges for each field
	ranges := []struct {
		min, max int
		name     string
		pattern  string
	}{
		{0, 59, "minute", `^(\*|[0-9]|[1-5][0-9])$`},
		{0, 23, "hour", `^(\*|[0-9]|1[0-9]|2[0-3])$`},
		{1, 31, "day of month", `^(\*|[1-9]|[12][0-9]|3[01])$`},
		{1, 12, "month", `^(\*|[1-9]|1[0-2])$`},
		{0, 6, "day of week", `^(\*|[0-6])$`},
	}

	// Validate each field
	for i, field := range fields {
		if field == "*" {
			continue
		}

		// Check if the field matches the pattern
		matched, _ := regexp.MatchString(ranges[i].pattern, field)
		if !matched {
			return fmt.Errorf("invalid %s in cron expression", ranges[i].name)
		}

		// If it's a number, check if it's within range
		if num, err := strconv.Atoi(field); err == nil {
			if num < ranges[i].min || num > ranges[i].max {
				return fmt.Errorf("invalid %s in cron expression: must be between %d and %d", ranges[i].name, ranges[i].min, ranges[i].max)
			}
		}
	}

	return nil
}

// CreateScrapingWorkflow creates a new scraping workflow in the database
func (db *Db) CreateScrapingWorkflow(ctx context.Context, workflow *lead_scraper_servicev1.ScrapingWorkflow) (*lead_scraper_servicev1.ScrapingWorkflow, error) {
	if workflow == nil {
		return nil, ErrInvalidInput
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// Validate cron expression first
	if err := validateCronExpression(workflow.CronExpression); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// validate the workflow
	if err := workflow.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// convert to ORM model
	workflowORM, err := workflow.ToORM(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to ORM model: %w", err)
	}

	// create the workflow
	result := db.Client.Engine.WithContext(ctx).Create(&workflowORM)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create scraping workflow: %w", result.Error)
	}

	// convert back to protobuf
	pbResult, err := workflowORM.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
	}

	return &pbResult, nil
}

// BatchCreateScrapingWorkflows creates multiple scraping workflows in the database
func (db *Db) BatchCreateScrapingWorkflows(ctx context.Context, workspaceID uint64, workflows []*lead_scraper_servicev1.ScrapingWorkflow) ([]*lead_scraper_servicev1.ScrapingWorkflow, error) {
	var (
		sQop = db.QueryOperator.ScrapingWorkflowORM
	)

	if len(workflows) == 0 {
		return nil, ErrInvalidInput
	}

	if workspaceID == 0 {
		return nil, ErrInvalidInput
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// Validate all workflows first
	for _, workflow := range workflows {
		if workflow == nil {
			return nil, ErrInvalidInput
		}

		// Validate cron expression
		if err := validateCronExpression(workflow.CronExpression); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}

		// Validate the workflow
		if err := workflow.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}
	}

	// Convert all workflows to ORM models
	workflowORMs := make([]*lead_scraper_servicev1.ScrapingWorkflowORM, 0, len(workflows))
	for _, workflow := range workflows {
		workflowORM, err := workflow.ToORM(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to ORM model: %w", err)
		}
		workflowORMs = append(workflowORMs, &workflowORM)
	}

	// Create workflows in batches
	err := sQop.WithContext(ctx).Where(sQop.WorkspaceId.Eq(workspaceID)).CreateInBatches(workflowORMs, batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create scraping workflows: %w", err)
	}

	// Convert back to protobuf
	pbResults := make([]*lead_scraper_servicev1.ScrapingWorkflow, 0, len(workflowORMs))
	for _, workflowORM := range workflowORMs {
		pbResult, err := workflowORM.ToPB(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
		}
		pbResults = append(pbResults, &pbResult)
	}

	return pbResults, nil
}
