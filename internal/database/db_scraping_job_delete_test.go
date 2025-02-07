package database

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteScrapingJob(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create a test job first
	testJob := testutils.GenerateRandomizedScrapingJob()

	created, err := conn.CreateScrapingJob(context.Background(), tc.Workspace.Id, testJob)
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
			name:      "[failure scenario] - already deleted job",
			wantError: true,
			errType:   ErrJobDoesNotExist,
			setup: func(t *testing.T) uint64 {
				// Create and delete a job
				job := testutils.GenerateRandomizedScrapingJob()
				created, err := conn.CreateScrapingJob(context.Background(), tc.Workspace.Id, job)
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
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	numJobs := 5
	var wg sync.WaitGroup
	errors := make(chan error, numJobs)
	jobIDs := make([]uint64, numJobs)

	// Create test jobs
	for i := 0; i < numJobs; i++ {
		job := testutils.GenerateRandomizedScrapingJob()
		created, err := conn.CreateScrapingJob(context.Background(), tc.Workspace.Id, job)
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

func TestBatchDeleteScrapingJobs(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create test jobs first
	numJobs := 5
	jobIDs := make([]uint64, numJobs)
	for i := 0; i < numJobs; i++ {
		testJob := testutils.GenerateRandomizedScrapingJob()
		created, err := conn.CreateScrapingJob(context.Background(), tc.Workspace.Id, testJob)
		require.NoError(t, err)
		require.NotNil(t, created)
		jobIDs[i] = created.Id
	}

	tests := []struct {
		name      string
		ids       []uint64
		wantError bool
		errType   error
		setup     func(t *testing.T) []uint64
		validate  func(t *testing.T, ids []uint64)
	}{
		{
			name:      "[success scenario] - valid ids",
			ids:       jobIDs,
			wantError: false,
			validate: func(t *testing.T, ids []uint64) {
				// Verify all jobs were deleted
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				for _, id := range ids {
					_, err := conn.GetScrapingJob(ctx, id)
					assert.Error(t, err)
					assert.ErrorIs(t, err, ErrJobDoesNotExist)
				}
			},
		},
		{
			name:      "[failure scenario] - empty ids slice",
			ids:       []uint64{},
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name:      "[failure scenario] - nil ids slice",
			ids:       nil,
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name:      "[failure scenario] - mix of existing and non-existing ids",
			wantError: true,
			errType:   ErrJobDoesNotExist,
			setup: func(t *testing.T) []uint64 {
				// Create one job and combine with non-existent ID
				job := testutils.GenerateRandomizedScrapingJob()
				created, err := conn.CreateScrapingJob(context.Background(), tc.Workspace.Id, job)
				require.NoError(t, err)
				require.NotNil(t, created)
				return []uint64{created.Id, 999999}
			},
		},
		{
			name:      "[failure scenario] - context timeout",
			ids:       jobIDs,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ids []uint64
			if tt.setup != nil {
				ids = tt.setup(t)
			} else {
				ids = tt.ids
			}

			ctx := context.Background()
			if tt.name == "[failure scenario] - context timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()
				time.Sleep(2 * time.Millisecond)
			}

			err := conn.BatchDeleteScrapingJobs(ctx, ids)

			if tt.wantError {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, ids)
			}
		})
	}
}

func TestBatchDeleteScrapingJobs_LargeBatch(t *testing.T) {	
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create a large batch of jobs
	numJobs := 100
	jobIDs := make([]uint64, numJobs)
	for i := 0; i < numJobs; i++ {
		job := testutils.GenerateRandomizedScrapingJob()
		created, err := conn.CreateScrapingJob(context.Background(), tc.Workspace.Id, job)
		require.NoError(t, err)
		require.NotNil(t, created)
		jobIDs[i] = created.Id
	}

	// Delete jobs in batch
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := conn.BatchDeleteScrapingJobs(ctx, jobIDs)
	require.NoError(t, err)

	// Verify all jobs were deleted
	for _, jobID := range jobIDs {
		_, err := conn.GetScrapingJob(ctx, jobID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrJobDoesNotExist)
	}
}
