package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

// ListScrapingWorkflows retrieves a list of scraping workflows with pagination
func (db *Db) ListScrapingWorkflows(ctx context.Context, limit, offset int) ([]*lead_scraper_servicev1.ScrapingWorkflow, error) {
	// Validate input parameters
	if limit < 0 {
		return nil, fmt.Errorf("invalid limit: %w", ErrInvalidInput)
	}
	if offset < 0 {
		return nil, fmt.Errorf("invalid offset: %w", ErrInvalidInput)
	}

	// Set default limit if not specified
	if limit == 0 {
		limit = 10 // default limit
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// Get the query operator
	workflowQop := db.QueryOperator.ScrapingWorkflowORM

	// Get the workflows
	workflowsORM, err := workflowQop.WithContext(ctx).
		Order(workflowQop.Id.Asc()).
		Limit(limit).
		Offset(offset).
		Find()
	if err != nil {
		return nil, fmt.Errorf("failed to list scraping workflows: %w", err)
	}

	workflows := make([]*lead_scraper_servicev1.ScrapingWorkflow, 0, len(workflowsORM))
	for _, workflowORM := range workflowsORM {
		workflow, err := workflowORM.ToPB(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
		}
		workflows = append(workflows, &workflow)
	}

	return workflows, nil
}
