package database

import (
	"context"
	"fmt"
)

// DeleteScrapingJob deletes a scraping job by ID
func (db *Db) DeleteScrapingJob(ctx context.Context, id uint64) error {
	var (
		sQop = db.QueryOperator.ScrapingJobORM
	)

	if id == 0 {
		return ErrInvalidInput
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// Check if the job exists first
	exists, err := sQop.WithContext(ctx).Where(sQop.Id.Eq(id)).Count()
	if err != nil {
		return fmt.Errorf("failed to check if job exists: %w", err)
	}
	if exists == 0 {
		return ErrJobDoesNotExist
	}

	// Delete the job
	result, err := sQop.WithContext(ctx).Where(sQop.Id.Eq(id)).Delete()
	if err != nil {
		return fmt.Errorf("failed to delete scraping job: %w", err)
	}

	// Check if any rows were affected
	if result.RowsAffected == 0 {
		return ErrJobDoesNotExist
	}

	return nil
} 

func (db *Db) BatchDeleteScrapingJobs(ctx context.Context, ids []uint64) error {
	var (
		sQop = db.QueryOperator.ScrapingJobORM
	)

	if len(ids) == 0 {
		return ErrInvalidInput
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// check if the jobs exist
	exists, err := sQop.WithContext(ctx).Where(sQop.Id.In(ids...)).Count()
	if err != nil {
		return fmt.Errorf("failed to check if jobs exist: %w", err)
	}

	if exists == 0 {
		return ErrJobDoesNotExist
	}

	// delete the jobs
	result, err := sQop.WithContext(ctx).Where(sQop.Id.In(ids...)).Delete()
	if err != nil {
		return fmt.Errorf("failed to delete scraping jobs: %w", err)
	}

	if result.RowsAffected != int64(len(ids)) {
		return ErrJobDoesNotExist
	}

	return nil
}
