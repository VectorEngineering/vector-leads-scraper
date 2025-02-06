package database

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateScrapingJob(t *testing.T) {
	validJob := testutils.GenerateRandomizedScrapingJob()

	tests := []struct {
		name      string
		job       *lead_scraper_servicev1.ScrapingJob
		wantError bool
		errType   error
		validate  func(t *testing.T, job *lead_scraper_servicev1.ScrapingJob)
	}{
		{
			name:      "[success scenario] - valid job",
			job:       validJob,
			wantError: false,
			validate: func(t *testing.T, job *lead_scraper_servicev1.ScrapingJob) {
				assert.NotNil(t, job)
				assert.NotZero(t, job.Id)
				assert.Equal(t, validJob.Status, job.Status)
				assert.Equal(t, validJob.Priority, job.Priority)
				assert.Equal(t, validJob.PayloadType, job.PayloadType)
				assert.Equal(t, validJob.Payload, job.Payload)
				assert.Equal(t, validJob.Name, job.Name)
				assert.Equal(t, validJob.Keywords, job.Keywords)
				assert.Equal(t, validJob.Lang, job.Lang)
				assert.Equal(t, validJob.Zoom, job.Zoom)
				assert.Equal(t, validJob.Lat, job.Lat)
				assert.Equal(t, validJob.Lon, job.Lon)
				assert.Equal(t, validJob.FastMode, job.FastMode)
				assert.Equal(t, validJob.Radius, job.Radius)
				assert.Equal(t, validJob.MaxTime, job.MaxTime)

				require.NotNil(t, job.CreatedAt)
				require.NotNil(t, job.UpdatedAt)
			},
		},
		{
			name:      "[failure scenario] - nil job",
			job:       nil,
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name:      "[failure scenario] - context timeout",
			job:       validJob,
			wantError: true,
		},
		{
			name: "[failure scenario] - invalid zoom value",
			job: &lead_scraper_servicev1.ScrapingJob{
				Status:      lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED,
				Priority:    1,
				PayloadType: "scraping_job",
				Name:        "Test Job",
				Zoom:        0,
				Lat:         "40.7128",
				Lon:         "-74.0060",
			},
			wantError: true,
			errType:   ErrInvalidInput,
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

			result, err := conn.CreateScrapingJob(ctx, tt.job)

			if tt.wantError {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			if tt.validate != nil {
				tt.validate(t, result)
			}

			// Clean up created job
			if result != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err := conn.DeleteScrapingJob(ctx, result.Id)
				require.NoError(t, err)
			}
		})
	}
}

func TestCreateScrapingJob_ConcurrentCreation(t *testing.T) {
	numJobs := 5
	var wg sync.WaitGroup
	errors := make(chan error, numJobs)
	jobs := make(chan *lead_scraper_servicev1.ScrapingJob, numJobs)

	// Create jobs concurrently
	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			job := testutils.GenerateRandomizedScrapingJob()
			job.Priority = int32(index + 1) // Different priority for each job

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			created, err := conn.CreateScrapingJob(ctx, job)
			if err != nil {
				errors <- err
				return
			}
			jobs <- created
		}(i)
	}

	wg.Wait()
	close(errors)
	close(jobs)

	// Clean up created jobs and collect them for validation
	createdJobs := make([]*lead_scraper_servicev1.ScrapingJob, 0)
	for job := range jobs {
		createdJobs = append(createdJobs, job)
	}

	defer func() {
		for _, job := range createdJobs {
			if job != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				err := conn.DeleteScrapingJob(ctx, job.Id)
				cancel()
				require.NoError(t, err)
			}
		}
	}()

	// Check for errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}
	require.Empty(t, errs, "Expected no errors during concurrent creation, got: %v", errs)

	// Verify all jobs were created successfully
	require.Equal(t, numJobs, len(createdJobs))

	// Track IDs to ensure uniqueness
	seenIDs := make(map[uint64]bool)

	for _, job := range createdJobs {
		require.NotNil(t, job)
		require.NotZero(t, job.Id)

		// Verify ID uniqueness
		_, exists := seenIDs[job.Id]
		assert.False(t, exists, "Duplicate job ID found: %d", job.Id)
		seenIDs[job.Id] = true

		// Validate timestamps
		require.NotNil(t, job.CreatedAt)
		require.NotNil(t, job.UpdatedAt)
	}
}

func TestBatchCreateScrapingJobs(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create a test workspace first
	createdWorkspace, err := conn.CreateWorkspace(context.Background(), &CreateWorkspaceInput{
		Workspace: testutils.GenerateRandomWorkspace(),
		AccountID: tc.Account.Id,
		TenantID:  tc.Tenant.Id,
		OrganizationID: tc.Organization.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createdWorkspace)

	// Clean up workspace after all tests
	defer func() {
		if createdWorkspace != nil {
			err := conn.DeleteWorkspace(context.Background(), createdWorkspace.Id)
			require.NoError(t, err)
		}
	}()

	tests := []struct {
		name        string
		workspaceID uint64
		jobs        []*lead_scraper_servicev1.ScrapingJob
		wantError   bool
		errType     error
		validate    func(t *testing.T, jobs []*lead_scraper_servicev1.ScrapingJob)
	}{
		{
			name:        "[success scenario] - multiple valid jobs",
			workspaceID: createdWorkspace.Id,
			jobs: func() []*lead_scraper_servicev1.ScrapingJob {
				numJobs := 5
				jobs := make([]*lead_scraper_servicev1.ScrapingJob, numJobs)
				for i := 0; i < numJobs; i++ {
					jobs[i] = testutils.GenerateRandomizedScrapingJob()
				}
				return jobs
			}(),
			wantError: false,
			validate: func(t *testing.T, jobs []*lead_scraper_servicev1.ScrapingJob) {
				assert.Equal(t, 5, len(jobs))
				for _, job := range jobs {
					assert.NotNil(t, job)
					assert.NotZero(t, job.Id)
				}
			},
		},
		{
			name:        "[failure scenario] - nil jobs slice",
			workspaceID: createdWorkspace.Id,
			jobs:        nil,
			wantError:   true,
			errType:     ErrInvalidInput,
		},
		{
			name:        "[failure scenario] - empty jobs slice",
			workspaceID: createdWorkspace.Id,
			jobs:        []*lead_scraper_servicev1.ScrapingJob{},
			wantError:   true,
			errType:     ErrInvalidInput,
		},
		{
			name:        "[failure scenario] - invalid workspace ID",
			workspaceID: 0,
			jobs: []*lead_scraper_servicev1.ScrapingJob{
				testutils.GenerateRandomizedScrapingJob(),
			},
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name:        "[failure scenario] - context timeout",
			workspaceID: createdWorkspace.Id,
			jobs: []*lead_scraper_servicev1.ScrapingJob{
				testutils.GenerateRandomizedScrapingJob(),
			},
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

			results, err := conn.BatchCreateScrapingJobs(ctx, tt.workspaceID, tt.jobs)

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

			// Clean up created jobs
			for _, job := range results {
				err := conn.DeleteScrapingJob(context.Background(), job.Id)
				require.NoError(t, err)
			}
		})
	}
}

func TestBatchCreateScrapingJobs_LargeBatch(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create a test workspace
	createdWorkspace, err := conn.CreateWorkspace(context.Background(), &CreateWorkspaceInput{
		Workspace: testutils.GenerateRandomWorkspace(),
		AccountID: tc.Account.Id,
		TenantID:  tc.Tenant.Id,
		OrganizationID: tc.Organization.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createdWorkspace)

	// Clean up workspace after test
	defer func() {
		err := conn.DeleteWorkspace(context.Background(), createdWorkspace.Id)
		require.NoError(t, err)
	}()

	// Create a large batch of jobs
	numJobs := 1000
	jobs := make([]*lead_scraper_servicev1.ScrapingJob, numJobs)
	for i := 0; i < numJobs; i++ {
		jobs[i] = testutils.GenerateRandomizedScrapingJob()
	}

	// Create jobs in batch
	results, err := conn.BatchCreateScrapingJobs(context.Background(), createdWorkspace.Id, jobs)
	require.NoError(t, err)
	require.NotNil(t, results)
	require.Equal(t, numJobs, len(results))

	// Verify all jobs were created with correct workspace ID
	for _, job := range results {
		assert.NotZero(t, job.Id)
	}

	// Clean up created jobs
	for _, job := range results {
		err := conn.DeleteScrapingJob(context.Background(), job.Id)
		require.NoError(t, err)
	}
}
