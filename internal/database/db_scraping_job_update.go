package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

// UpdateScrapingJob updates an existing scraping job
func (db *Db) UpdateScrapingJob(ctx context.Context, job *lead_scraper_servicev1.ScrapingJob) (*lead_scraper_servicev1.ScrapingJob, error) {
	if job == nil || job.Id == 0 {
		return nil, fmt.Errorf("invalid job: %w", ErrInvalidInput)
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// Get the query operator
	jobQop := db.QueryOperator.ScrapingJobORM

	// Check if job exists and get current status
	existingJob, err := jobQop.WithContext(ctx).Where(jobQop.Id.Eq(job.Id)).First()
	if err != nil {
		return nil, fmt.Errorf("job does not exist: %w", ErrJobDoesNotExist)
	}

	// Convert ORM to protobuf to get the current status enum
	existingPB, err := existingJob.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert existing job to protobuf: %w", err)
	}

	// Validate status transition
	if !isValidStatusTransition(existingPB.Status, job.Status) {
		return nil, fmt.Errorf("invalid status transition from %v to %v: %w", existingPB.Status, job.Status, ErrInvalidInput)
	}

	// convert to ORM model
	jobORM, err := job.ToORM(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to ORM model: %w", err)
	}

	// update the job
	if _, err := jobQop.WithContext(ctx).Where(jobQop.Id.Eq(job.Id)).Updates(&jobORM); err != nil {
		return nil, fmt.Errorf("failed to update scraping job: %w", err)
	}

	// get the updated record
	updatedJob, err := jobQop.WithContext(ctx).Where(jobQop.Id.Eq(job.Id)).First()
	if err != nil {
		return nil, fmt.Errorf("failed to get updated scraping job: %w", err)
	}

	pbResult, err := updatedJob.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
	}

	return &pbResult, nil
}

// isValidStatusTransition checks if a status transition is valid
func isValidStatusTransition(from, to lead_scraper_servicev1.BackgroundJobStatus) bool {
	// Allow same status transitions for idempotency
	if from == to {
		return true
	}

	// Define valid transitions
	switch from {
	case lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_UNSPECIFIED:
		// Allow any transition from unspecified
		return true
	case lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED:
		// From QUEUED, can only go to IN_PROGRESS or CANCELLED
		return to == lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_IN_PROGRESS ||
			to == lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_CANCELLED
	case lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_IN_PROGRESS:
		// From IN_PROGRESS, can go to any terminal state
		return to == lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED ||
			to == lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_FAILED ||
			to == lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_CANCELLED ||
			to == lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_TIMED_OUT
	case lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED,
		lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_FAILED,
		lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_CANCELLED,
		lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_TIMED_OUT:
		// From any terminal state, can only go back to QUEUED for retries
		return to == lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED
	default:
		return false
	}
}

// isTerminalStatus checks if a status is terminal (completed, failed, cancelled, timed out)
func isTerminalStatus(status lead_scraper_servicev1.BackgroundJobStatus) bool {
	switch status {
	case lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED,
		lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_FAILED,
		lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_CANCELLED,
		lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_TIMED_OUT:
		return true
	default:
		return false
	}
}
