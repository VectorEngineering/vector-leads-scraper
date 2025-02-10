package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

// BatchUpdateScrapingJobs updates multiple scraping jobs in a single batch operation.
// It processes jobs in batches to avoid overwhelming the database.
// If an error occurs during batch processing, it will be logged and the operation
// will continue with the next batch.
func (db *Db) BatchUpdateScrapingJobs(ctx context.Context, workspaceID uint64, jobs []*lead_scraper_servicev1.ScrapingJob) ([]*lead_scraper_servicev1.ScrapingJob, error) {
	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// check that the workspace of interest exists
	workspaceQop := db.QueryOperator.WorkspaceORM

	// get the workspace by id
	workspace, err := workspaceQop.GetByID(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	if len(jobs) == 0 {
		return nil, ErrInvalidInput
	}

	batches := BreakIntoBatches[*lead_scraper_servicev1.ScrapingJob](jobs, batchSize)
	resultingJobs := make([]*lead_scraper_servicev1.ScrapingJobORM, 0, len(jobs))

	// Process each batch
	for _, batch := range batches {
		ormJobs, err := db.convertJobsToORM(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("failed to convert jobs to orm: %w", err)
		}

		// for each batch we append to the workspace of interest
		if err := workspaceQop.ScrapingJobs.Model(&workspace).Replace(ormJobs...); err != nil {
			return nil, fmt.Errorf("failed to append jobs to workspace: %w", err)
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

		resultingJobs = append(resultingJobs, ormJobs...)
	}

	// convert the orm jobs to jobs
	resultSet, err := db.convertJobsORMToJobs(ctx, resultingJobs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert jobs to orm: %w", err)
	}

	return resultSet, nil
}

// convert the jobs to orm jobs
func (db *Db) convertJobsToORM(ctx context.Context, jobs []*lead_scraper_servicev1.ScrapingJob) ([]*lead_scraper_servicev1.ScrapingJobORM, error) {
	ormJobs := make([]*lead_scraper_servicev1.ScrapingJobORM, 0, len(jobs))
	for _, job := range jobs {
		ormJob, err := job.ToORM(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to convert job to orm: %w", err)
		}
		ormJobs = append(ormJobs, &ormJob)
	}
	return ormJobs, nil
}

// convert the orm jobs to jobs
func (db *Db) convertJobsORMToJobs(ctx context.Context, ormJobs []*lead_scraper_servicev1.ScrapingJobORM) ([]*lead_scraper_servicev1.ScrapingJob, error) {
	jobs := make([]*lead_scraper_servicev1.ScrapingJob, 0, len(ormJobs))
	for _, ormJob := range ormJobs {
		job, err := ormJob.ToPB(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to convert job to pb: %w", err)
		}
		jobs = append(jobs, &job)
	}
	return jobs, nil
}
