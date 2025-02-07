package database

import (
	"context"
	"testing"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchUpdateScrapingJobs(t *testing.T) {
	ctx := context.Background()
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create test workflow
	workflow := testutils.GenerateRandomScrapingWorkflow()
	createdWorkflow, err := conn.CreateScrapingWorkflow(ctx, tc.Workspace.Id, workflow)
	require.NoError(t, err)
	require.NotNil(t, createdWorkflow)

	// Clean up workflow after test
	defer func() {
		err := conn.DeleteScrapingWorkflow(ctx, createdWorkflow.Id)
		if err != nil {
			t.Logf("Failed to cleanup test workflow: %v", err)
		}
	}()

	// Create test jobs
	jobs := make([]*lead_scraper_servicev1.ScrapingJob, 0, 5)
	for i := 0; i < 5; i++ {
		job := testutils.GenerateRandomizedScrapingJob()
		job.Status = lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED
		
		createdJob, err := conn.CreateScrapingJob(ctx, tc.Workspace.Id, job)
		require.NoError(t, err)
		require.NotNil(t, createdJob)
		jobs = append(jobs, createdJob)
	}

	// Clean up jobs after test
	defer func() {
		for _, job := range jobs {
			err := conn.DeleteScrapingJob(ctx, job.Id)
			if err != nil {
				t.Logf("Failed to cleanup test job: %v", err)
			}
		}
	}()

	tests := []struct {
		name    string
		jobs    []*lead_scraper_servicev1.ScrapingJob
		wantErr bool
		setup   func(jobs []*lead_scraper_servicev1.ScrapingJob)
		verify  func(t *testing.T, updatedJobs []*lead_scraper_servicev1.ScrapingJob)
	}{
		{
			name:    "success - update multiple jobs",
			jobs:    jobs,
			wantErr: false,
			setup: func(jobs []*lead_scraper_servicev1.ScrapingJob) {
				for _, job := range jobs {
					job.Status = lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED
				}
			},
			verify: func(t *testing.T, updatedJobs []*lead_scraper_servicev1.ScrapingJob) {
				require.Len(t, updatedJobs, len(jobs))
				for i, job := range updatedJobs {
					assert.Equal(t, jobs[i].Id, job.Id)
					assert.Equal(t, lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED, job.Status)
					
					// // Verify job was actually updated in database
					// fetchedJob, err := conn.GetScrapingJob(ctx, job.Id)
					// require.NoError(t, err)
					// assert.Equal(t, lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED, fetchedJob.Status)
				}
			},
		},
		{
			name:    "error - empty jobs list",
			jobs:    []*lead_scraper_servicev1.ScrapingJob{},
			wantErr: true,
		},
		{
			name:    "error - nil jobs list",
			jobs:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(tt.jobs)
			}

			updatedJobs, err := conn.BatchUpdateScrapingJobs(ctx, tc.Workspace.Id, tt.jobs)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, updatedJobs)

			if tt.verify != nil {
				tt.verify(t, updatedJobs)
			}
		})
	}
} 