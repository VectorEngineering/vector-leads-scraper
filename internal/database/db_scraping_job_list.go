package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

// ListScrapingJobs retrieves a list of scraping jobs with pagination
func (db *Db) ListScrapingJobs(ctx context.Context, limit, offset uint64) ([]*lead_scraper_servicev1.ScrapingJob, error) {
	// Validate input parameters
	if limit <= 0 || int64(limit) < 0 {
		return nil, fmt.Errorf("invalid limit: %w", ErrInvalidInput)
	}
	if int64(offset) < 0 {
		return nil, fmt.Errorf("invalid offset: %w", ErrInvalidInput)
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	var jobsORM []lead_scraper_servicev1.ScrapingJobORM
	result := db.Client.Engine.WithContext(ctx).
		Order("created_at desc").
		Limit(int(limit)).
		Offset(int(offset)).
		Find(&jobsORM)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list scraping jobs: %w", result.Error)
	}

	pbResults := make([]*lead_scraper_servicev1.ScrapingJob, len(jobsORM))
	for i, jobORM := range jobsORM {
		pbResult, err := jobORM.ToPB(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
		}
		pbResults[i] = &pbResult
	}

	return pbResults, nil
}
