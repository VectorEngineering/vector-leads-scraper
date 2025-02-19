package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

// GetScrapingJob retrieves a scraping job by ID
func (db *Db) GetScrapingJob(ctx context.Context, id uint64) (*lead_scraper_servicev1.ScrapingJob, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// Get the query operator
	jobQop := db.QueryOperator.ScrapingJobORM

	// Get the job
	jobORM, err := jobQop.WithContext(ctx).Where(jobQop.Id.Eq(id)).First()
	if err != nil {
		if err.Error() == "record not found" {
			return nil, fmt.Errorf("%w: %v", ErrJobDoesNotExist, err)
		}
		return nil, fmt.Errorf("failed to get scraping job: %w", err)
	}

	pbResult, err := jobORM.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
	}

	return &pbResult, nil
}
