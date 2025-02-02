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

// testScrapingJob returns a valid scraping job for testing
func testScrapingJob() *lead_scraper_servicev1.ScrapingJob {
	return &lead_scraper_servicev1.ScrapingJob{
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
}

func TestCreateScrapingJob(t *testing.T) {
	validJob := testScrapingJob()

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
				assert.Equal(t, job.CreatedAt.AsTime().Unix(), job.UpdatedAt.AsTime().Unix())
				assert.True(t, job.CreatedAt.AsTime().Before(time.Now()) || job.CreatedAt.AsTime().Equal(time.Now()))
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
			job := testScrapingJob()
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
		
		assert.Equal(t, lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED, job.Status)
		assert.Equal(t, "scraping_job", job.PayloadType)
		assert.Equal(t, "Test Job", job.Name)
		
		// Validate timestamps
		require.NotNil(t, job.CreatedAt)
		require.NotNil(t, job.UpdatedAt)
		assert.Equal(t, job.CreatedAt.AsTime().Unix(), job.UpdatedAt.AsTime().Unix())
		assert.True(t, job.CreatedAt.AsTime().Before(time.Now()) || job.CreatedAt.AsTime().Equal(time.Now()))
	}
}
