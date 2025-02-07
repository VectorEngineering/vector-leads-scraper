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
	if _, err := conn.CreateScrapingWorkflow(ctx, tc.Workspace.Id, testutils.GenerateRandomScrapingWorkflow()); err != nil {
		t.Fatalf("failed to create scraping workflow: %v", err)
	}

	// Create test jobs
	jobs := make([]*lead_scraper_servicev1.ScrapingJob, 0, 5)
	for i := 0; i < 5; i++ {
		job, err := conn.CreateScrapingJob(ctx, tc.Workspace.Id, testutils.GenerateRandomizedScrapingJob())
		require.NoError(t, err)
		jobs = append(jobs, job)
	}

	tests := []struct {
		name    string
		jobs    []*lead_scraper_servicev1.ScrapingJob
		wantErr bool
	}{
		{
			name:    "success - update multiple jobs",
			jobs:    jobs,
			wantErr: false,
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
			if !tt.wantErr && len(tt.jobs) > 0 {
				// Update job statuses
				for _, job := range tt.jobs {
					job.Status = lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED
				}
			}

			updatedJobs, err := conn.BatchUpdateScrapingJobs(ctx, tc.Workspace.Id, tt.jobs)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, updatedJobs)
			assert.Equal(t, len(tt.jobs), len(updatedJobs))

			// Verify each job was updated correctly
			for i, job := range tt.jobs {
				assert.Equal(t, job.Id, updatedJobs[i].Id)
				assert.Equal(t, job.Status, updatedJobs[i].Status)
			}

			// Verify jobs were actually updated in the database
			for _, job := range tt.jobs {
				fetchedJob, err := conn.GetScrapingJob(ctx, job.Id)
				require.NoError(t, err)
				assert.Equal(t, job.Status, fetchedJob.Status)
			}
		})
	}
} 