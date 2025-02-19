package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

// UpdateScrapingWorkflow updates an existing scraping workflow
func (db *Db) UpdateScrapingWorkflow(ctx context.Context, workflow *lead_scraper_servicev1.ScrapingWorkflow) (*lead_scraper_servicev1.ScrapingWorkflow, error) {
	if workflow == nil || workflow.Id == 0 {
		return nil, ErrInvalidInput
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// Get the query operator
	workflowQop := db.QueryOperator.ScrapingWorkflowORM

	// convert to ORM model
	workflowORM, err := workflow.ToORM(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to ORM model: %w", err)
	}

	// update the workflow
	if _, err := workflowQop.WithContext(ctx).Where(workflowQop.Id.Eq(workflow.Id)).Updates(&workflowORM); err != nil {
		return nil, fmt.Errorf("failed to update scraping workflow: %w", err)
	}

	// get the updated record
	updatedWorkflow, err := workflowQop.WithContext(ctx).Where(workflowQop.Id.Eq(workflow.Id)).First()
	if err != nil {
		return nil, fmt.Errorf("failed to get updated scraping workflow: %w", err)
	}

	pbResult, err := updatedWorkflow.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
	}

	return &pbResult, nil
}
