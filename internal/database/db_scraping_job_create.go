package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

// CreateScrapingJob creates a new scraping job in the database
func (db *Db) CreateScrapingJob(ctx context.Context, workspaceID uint64, job *lead_scraper_servicev1.ScrapingJob) (*lead_scraper_servicev1.ScrapingJob, error) {
	// ensure the db operation executes within the specified timeout
	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if job == nil {
		return nil, ErrInvalidInput
	}

	// validate the job
	if err := job.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// check that the workspace exists
	workspaceQop := db.QueryOperator.WorkspaceORM
	workspace, err := workspaceQop.GetByID(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// convert to ORM model
	jobORM, err := job.ToORM(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// append the job to the workspace
	if err := workspaceQop.ScrapingJobs.Model(&workspace).Append(&jobORM); err != nil {
		return nil, fmt.Errorf("failed to append job to workspace: %w", err)
	}

	// update the workspace
	res, err := workspaceQop.Where(workspaceQop.Id.Eq(workspaceID)).Updates(&workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}

	if res.RowsAffected == 0 || res.Error != nil {
		if res.Error != nil {
			return nil, fmt.Errorf("failed to update workspace: %w", res.Error)
		}

		return nil, fmt.Errorf("failed to update workspace: %w", ErrWorkspaceDoesNotExist)
	}

	// convert back to protobuf
	pbResult, err := jobORM.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
	}

	return &pbResult, nil
}

func (db *Db) BatchCreateScrapingJobs(ctx context.Context, workspaceID uint64, jobs []*lead_scraper_servicev1.ScrapingJob) ([]*lead_scraper_servicev1.ScrapingJob, error) {
	var (
		sQop = db.QueryOperator.ScrapingJobORM
	)

	if len(jobs) == 0 {
		return nil, ErrInvalidInput
	}

	if workspaceID == 0 {
		return nil, ErrInvalidInput
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// convert to ORM model
	jobORMs, err := db.convertScrapingJobsToORM(ctx, jobs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to ORM model: %w", err)
	}

	// insert the jobs in batches
	if err := sQop.WithContext(ctx).Where(sQop.WorkspaceId.Eq(workspaceID)).CreateInBatches(jobORMs, batchSize); err != nil {
		return nil, fmt.Errorf("failed to insert jobs: %w", err)
	}

	// convert back to protobuf
	pbResults := make([]*lead_scraper_servicev1.ScrapingJob, 0, len(jobORMs))
	for _, jobORM := range jobORMs {
		pbResult, err := jobORM.ToPB(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
		}
		pbResults = append(pbResults, &pbResult)
	}

	return pbResults, nil
}

func (db *Db) convertScrapingJobsToORM(ctx context.Context, jobs []*lead_scraper_servicev1.ScrapingJob) ([]*lead_scraper_servicev1.ScrapingJobORM, error) {
	jobORMs := make([]*lead_scraper_servicev1.ScrapingJobORM, 0, len(jobs))
	for _, job := range jobs {
		jobORM, err := job.ToORM(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to ORM model: %w", err)
		}
		jobORMs = append(jobORMs, &jobORM)
	}

	return jobORMs, nil
}
