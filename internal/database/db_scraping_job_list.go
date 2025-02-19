package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/go-playground/validator/v10"
	"go.uber.org/multierr"
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

	// Get the query operator
	jobQop := db.QueryOperator.ScrapingJobORM

	// Get the jobs
	jobsORM, err := jobQop.WithContext(ctx).
		Order(jobQop.CreatedAt.Desc()).
		Limit(int(limit)).
		Offset(int(offset)).
		Find()
	if err != nil {
		return nil, fmt.Errorf("failed to list scraping jobs: %w", err)
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

type ListScrapingJobsByWorkspaceInput struct {
	WorkspaceID uint64
	WorkflowID  uint64
	Limit       uint64 `validate:"required,gt=0"`
	Offset      uint64 `validate:"required,gte=0"`
}

func (input *ListScrapingJobsByWorkspaceInput) Validate() error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(input); err != nil {
		return multierr.Append(ErrInvalidInput, err)
	}
	return nil
}

func (db *Db) ListScrapingJobsByParams(ctx context.Context, input *ListScrapingJobsByWorkspaceInput) ([]*lead_scraper_servicev1.ScrapingJob, error) {
	// Validate input parameters
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// Get the query operator
	jobQop := db.QueryOperator.ScrapingJobORM

	// build the query based on the presence of the workflow id and workspace id
	query := jobQop.WithContext(ctx)
	if input.WorkflowID != 0 {
		query = query.Where(jobQop.ScrapingWorkflowId.Eq(input.WorkflowID))
	}
	if input.WorkspaceID != 0 {
		query = query.Where(jobQop.WorkspaceId.Eq(input.WorkspaceID))
	}

	// Get the jobs
	jobsORM, err := query.
		Order(jobQop.CreatedAt.Desc()).
		Limit(int(input.Limit)).
		Offset(int(input.Offset)).
		Find()
	if err != nil {
		return nil, fmt.Errorf("failed to list scraping jobs: %w", err)
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

// ListScrapingJobsByStatus retrieves a list of scraping jobs with a specific status
func (db *Db) ListScrapingJobsByStatus(ctx context.Context, status lead_scraper_servicev1.BackgroundJobStatus, limit, offset uint64) ([]*lead_scraper_servicev1.ScrapingJob, error) {
	// Validate input parameters
	if limit <= 0 || int64(limit) < 0 {
		return nil, fmt.Errorf("invalid limit: %w", ErrInvalidInput)
	}
	if int64(offset) < 0 {
		return nil, fmt.Errorf("invalid offset: %w", ErrInvalidInput)
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// Get the query operator
	jobQop := db.QueryOperator.ScrapingJobORM

	// Get the jobs with status filter
	jobsORM, err := jobQop.WithContext(ctx).
		Where(jobQop.Status.Eq(status.String())).
		Order(jobQop.CreatedAt.Desc()).
		Limit(int(limit)).
		Offset(int(offset)).
		Find()
	if err != nil {
		return nil, fmt.Errorf("failed to list scraping jobs: %w", err)
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
