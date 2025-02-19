package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListScrapingJobs(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()
	table := lead_scraper_servicev1.ScrapingJobORM{}.TableName()

	// Clean up any existing jobs first
	result := conn.Client.Engine.Exec(fmt.Sprintf("DELETE FROM %s", table))
	require.NoError(t, result.Error)

	// Create multiple test jobs
	numJobs := 5
	jobIDs := make([]uint64, numJobs)

	for i := 0; i < numJobs; i++ {
		job := testutils.GenerateRandomizedScrapingJob()
		created, err := conn.CreateScrapingJob(context.Background(), tc.Workspace.Id, job)
		require.NoError(t, err)
		require.NotNil(t, created)
		jobIDs[i] = created.Id

		// Add a small delay to ensure consistent ordering by creation time
		time.Sleep(10 * time.Millisecond)
	}

	// Clean up after all tests
	defer func() {
		for _, id := range jobIDs {
			err := conn.DeleteScrapingJob(context.Background(), id)
			require.NoError(t, err)
		}
	}()

	tests := []struct {
		name      string
		limit     int
		offset    int
		wantError bool
		errType   error
		validate  func(t *testing.T, jobs []*lead_scraper_servicev1.ScrapingJob)
	}{
		{
			name:      "[success scenario] - get all jobs",
			limit:     10,
			offset:    0,
			wantError: false,
			validate: func(t *testing.T, jobs []*lead_scraper_servicev1.ScrapingJob) {
				assert.Len(t, jobs, numJobs)
				// Verify jobs are returned in order of creation (newest first)
				for _, job := range jobs {
					assert.NotNil(t, job)
					assert.NotZero(t, job.Id)
				}
			},
		},
		{
			name:      "[success scenario] - pagination first page",
			limit:     3,
			offset:    0,
			wantError: false,
			validate: func(t *testing.T, jobs []*lead_scraper_servicev1.ScrapingJob) {
				assert.Len(t, jobs, 3)
				for _, job := range jobs {
					assert.NotNil(t, job)
					assert.NotZero(t, job.Id)
				}
			},
		},
		{
			name:      "[success scenario] - pagination second page",
			limit:     3,
			offset:    3,
			wantError: false,
			validate: func(t *testing.T, jobs []*lead_scraper_servicev1.ScrapingJob) {
				assert.Len(t, jobs, 2) // Only 2 remaining jobs
				for _, job := range jobs {
					assert.NotNil(t, job)
					assert.NotZero(t, job.Id)
				}
			},
		},
		{
			name:      "[success scenario] - empty result",
			limit:     10,
			offset:    numJobs + 1,
			wantError: false,
			validate: func(t *testing.T, jobs []*lead_scraper_servicev1.ScrapingJob) {
				assert.Empty(t, jobs)
			},
		},
		{
			name:      "[failure scenario] - invalid limit",
			limit:     -1,
			offset:    0,
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name:      "[failure scenario] - invalid offset",
			limit:     10,
			offset:    -1,
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name:      "[failure scenario] - context timeout",
			limit:     10,
			offset:    0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.name == "[failure scenario] - context timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()
				time.Sleep(2 * time.Millisecond)
			}

			results, err := conn.ListScrapingJobs(ctx, uint64(tt.limit), uint64(tt.offset))

			if tt.wantError {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, results)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, results)

			if tt.validate != nil {
				tt.validate(t, results)
			}
		})
	}
}

func TestListScrapingJobs_EmptyDatabase(t *testing.T) {
	table := lead_scraper_servicev1.ScrapingJobORM{}.TableName()
	// Clean up any existing jobs first
	result := conn.Client.Engine.Exec(fmt.Sprintf("DELETE FROM %s", table))
	require.NoError(t, result.Error)

	results, err := conn.ListScrapingJobs(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestDb_ListScrapingJobsByStatus(t *testing.T) {
	table := lead_scraper_servicev1.ScrapingJobORM{}.TableName()
	// Initialize test context
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Clean up any existing jobs first
	result := conn.Client.Engine.Exec(fmt.Sprintf("DELETE FROM %s", table))
	require.NoError(t, result.Error)

	// Create test jobs with different statuses
	jobStatuses := []lead_scraper_servicev1.BackgroundJobStatus{
		lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED,
		lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_IN_PROGRESS,
		lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED,
		lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_FAILED,
	}

	// Create test jobs
	var createdJobs []*lead_scraper_servicev1.ScrapingJob
	for _, status := range jobStatuses {
		job := testutils.GenerateRandomizedScrapingJob()
		job.Status = status
		created, err := conn.CreateScrapingJob(context.Background(), tc.Workspace.Id, job)
		require.NoError(t, err)
		require.NotNil(t, created)
		createdJobs = append(createdJobs, created)
	}

	// Cleanup after test
	defer func() {
		for _, job := range createdJobs {
			err := conn.DeleteScrapingJob(context.Background(), job.Id)
			require.NoError(t, err)
		}
	}()

	tests := []struct {
		name     string
		status   lead_scraper_servicev1.BackgroundJobStatus
		limit    uint64
		offset   uint64
		wantErr  bool
		errType  error
		validate func(t *testing.T, got []*lead_scraper_servicev1.ScrapingJob)
	}{
		{
			name:   "[success] list queued jobs",
			status: lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED,
			limit:  10,
			offset: 0,
			validate: func(t *testing.T, got []*lead_scraper_servicev1.ScrapingJob) {
				require.Len(t, got, 1)
				assert.Equal(t, lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED, got[0].Status)
			},
		},
		{
			name:   "[success] list in progress jobs",
			status: lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_IN_PROGRESS,
			limit:  10,
			offset: 0,
			validate: func(t *testing.T, got []*lead_scraper_servicev1.ScrapingJob) {
				require.Len(t, got, 1)
				assert.Equal(t, lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_IN_PROGRESS, got[0].Status)
			},
		},
		{
			name:   "[success] list with offset",
			status: lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED,
			limit:  1,
			offset: 0,
			validate: func(t *testing.T, got []*lead_scraper_servicev1.ScrapingJob) {
				require.Len(t, got, 1)
				assert.Equal(t, lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED, got[0].Status)
			},
		},
		{
			name:   "[success] empty result for non-existent status",
			status: lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_UNSPECIFIED,
			limit:  10,
			offset: 0,
			validate: func(t *testing.T, got []*lead_scraper_servicev1.ScrapingJob) {
				assert.Len(t, got, 0)
			},
		},
		{
			name:    "[failure] invalid limit",
			status:  lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED,
			limit:   0,
			offset:  0,
			wantErr: true,
			errType: ErrInvalidInput,
		},
		{
			name:    "[failure] invalid offset",
			status:  lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED,
			limit:   10,
			offset:  ^uint64(0), // Max uint64 value to force negative int64
			wantErr: true,
			errType: ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conn.ListScrapingJobsByStatus(context.Background(), tt.status, tt.limit, tt.offset)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}
