package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

// GetScrapingWorkflow retrieves a scraping workflow by ID
func (db *Db) GetScrapingWorkflow(ctx context.Context, id uint64) (*lead_scraper_servicev1.ScrapingWorkflow, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// Get the query operator
	workflowQop := db.QueryOperator.ScrapingWorkflowORM

	// Get the workflow
	workflowORM, err := workflowQop.WithContext(ctx).Where(workflowQop.Id.Eq(id)).First()
	if err != nil {
		if err.Error() == "record not found" {
			return nil, fmt.Errorf("%w: %v", ErrWorkflowDoesNotExist, err)
		}
		return nil, fmt.Errorf("failed to get scraping workflow: %w", err)
	}

	pbResult, err := workflowORM.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
	}

	return &pbResult, nil
}
