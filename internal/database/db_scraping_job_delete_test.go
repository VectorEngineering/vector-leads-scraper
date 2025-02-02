package database

import (
	"context"
	"sync"
	"testing"
	"time"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteScrapingJob(t *testing.T) {
	// Create a test job first
	testJob := &lead_scraper_servicev1.ScrapingJob{
		Status:      lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED,
		Priority:    1,
		PayloadType: "scraping_job",
		Payload:     []byte(`{"query": "test query"}`),
		Name:        "Test Job",
		Keywords:    []string{"keyword1", "keyword2"},
		Lang:        "en",
		Zoom:        15,
		Lat:         "40.7128",
		Lon:         "-74.0060",
		FastMode:    false,
		Radius:      10000,
		MaxTime:     3600,
	}

	created, err := conn.CreateScrapingJob(context.Background(), testJob)
	require.NoError(t, err)
	require.NotNil(t, created)

	tests := []struct {
		name      string
		jobID     uint64
		wantError bool
		errType   error
		setup     func(t *testing.T) uint64
		validate  func(t *testing.T, jobID uint64)
	}{
		{
			name:      "[success scenario] - valid id",
			jobID:     created.Id,
			wantError: false,
			validate: func(t *testing.T, jobID uint64) {
				// Verify the job was deleted
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_, err := conn.GetScrapingJob(ctx, jobID)
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrJobDoesNotExist)
			},
		},
		{
			name:      "[failure scenario] - zero id",
			jobID:     0,
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name:      "[failure scenario] - non-existent id",
			jobID:     999999,
			wantError: true,
			errType:   ErrJobDoesNotExist,
			validate: func(t *testing.T, jobID uint64) {
				// Double check that the job really doesn't exist
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_, err := conn.GetScrapingJob(ctx, jobID)
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrJobDoesNotExist)
			},
		},
		{
			name:      "[failure scenario] - already deleted job",
			wantError: true,
			errType:   ErrJobDoesNotExist,
			setup: func(t *testing.T) uint64 {
				// Create and delete a job
				job := &lead_scraper_servicev1.ScrapingJob{
					Status:      lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED,
					Priority:    1,
					PayloadType: "scraping_job",
					Name:        "Test Job",
					Zoom:        15,
					Lat:         "40.7128",
					Lon:         "-74.0060",
				}
				created, err := conn.CreateScrapingJob(context.Background(), job)
				require.NoError(t, err)
				require.NotNil(t, created)

				jobID := created.Id
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				err = conn.DeleteScrapingJob(ctx, jobID)
				cancel()
				require.NoError(t, err)

				return jobID
			},
		},
		{
			name:      "[failure scenario] - context timeout",
			jobID:     created.Id,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jobID uint64
			if tt.setup != nil {
				jobID = tt.setup(t)
			} else {
				jobID = tt.jobID
			}

			ctx := context.Background()
			if tt.name == "[failure scenario] - context timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()
				time.Sleep(2 * time.Millisecond)
			}

			err := conn.DeleteScrapingJob(ctx, jobID)

			if tt.wantError {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, jobID)
			}
		})
	}
}

func TestDeleteScrapingJob_ConcurrentDeletions(t *testing.T) {
	numJobs := 5
	var wg sync.WaitGroup
	errors := make(chan error, numJobs)
	jobIDs := make([]uint64, numJobs)

	// Create test jobs
	for i := 0; i < numJobs; i++ {
		job := &lead_scraper_servicev1.ScrapingJob{
			Status:      lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED,
			Priority:    int32(i + 1), // Different priority for each job
			PayloadType: "scraping_job",
			Payload:     []byte(`{"query": "test query"}`),
			Name:        "Test Job",
			Keywords:    []string{"keyword1", "keyword2"},
			Lang:        "en",
			Zoom:        15,
			Lat:         "40.7128",
			Lon:         "-74.0060",
			FastMode:    false,
			Radius:      10000,
			MaxTime:     3600,
		}
		created, err := conn.CreateScrapingJob(context.Background(), job)
		require.NoError(t, err)
		require.NotNil(t, created)
		jobIDs[i] = created.Id
	}

	// Delete jobs concurrently
	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		go func(jobID uint64) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			if err := conn.DeleteScrapingJob(ctx, jobID); err != nil {
				errors <- err
			}
		}(jobIDs[i])
	}

	wg.Wait()
	close(errors)

	// Check for errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}
	require.Empty(t, errs, "Expected no errors during concurrent deletions, got: %v", errs)

	// Verify all jobs were deleted
	for _, jobID := range jobIDs {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := conn.GetScrapingJob(ctx, jobID)
		cancel()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrJobDoesNotExist)
	}

	// Try to delete already deleted jobs - should fail with ErrJobDoesNotExist
	var deleteErrs []error
	for _, jobID := range jobIDs {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := conn.DeleteScrapingJob(ctx, jobID)
		cancel()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrJobDoesNotExist)
		if err != nil {
			deleteErrs = append(deleteErrs, err)
		}
	}
	require.NotEmpty(t, deleteErrs, "Expected errors when deleting already deleted jobs")
	for _, err := range deleteErrs {
		assert.ErrorIs(t, err, ErrJobDoesNotExist)
	}
}
