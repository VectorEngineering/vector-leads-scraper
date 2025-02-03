package database

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateScrapingWorkflow(t *testing.T) {
	validWorkflow := &lead_scraper_servicev1.ScrapingWorkflow{
		CronExpression:         "0 0 * * *",
		RetryCount:            0,
		MaxRetries:            5,
		AlertEmails:           "test@example.com",
		GeoFencingRadius:     float32(1000.0),
		GeoFencingLat:        40.7128,
		GeoFencingLon:        -74.0060,
		GeoFencingZoomMin:    1,
		GeoFencingZoomMax:    20,
		IncludeReviews:       true,
		IncludePhotos:        true,
		IncludeBusinessHours: true,
		MaxReviewsPerBusiness: 100,
		RespectRobotsTxt:     true,
		AcceptTermsOfService:  true,
		UserAgent:            "TestBot/1.0",
	}

	tests := []struct {
		name      string
		workflow  *lead_scraper_servicev1.ScrapingWorkflow
		wantError bool
		errType   error
		validate  func(t *testing.T, workflow *lead_scraper_servicev1.ScrapingWorkflow)
	}{
		{
			name:      "[success scenario] - valid workflow",
			workflow:  validWorkflow,
			wantError: false,
			validate: func(t *testing.T, workflow *lead_scraper_servicev1.ScrapingWorkflow) {
				assert.NotNil(t, workflow)
				assert.NotZero(t, workflow.Id)
				assert.Equal(t, validWorkflow.CronExpression, workflow.CronExpression)
				assert.Equal(t, validWorkflow.RetryCount, workflow.RetryCount)
				assert.Equal(t, validWorkflow.MaxRetries, workflow.MaxRetries)
				assert.Equal(t, validWorkflow.AlertEmails, workflow.AlertEmails)
				assert.Equal(t, validWorkflow.GeoFencingRadius, workflow.GeoFencingRadius)
				assert.Equal(t, validWorkflow.GeoFencingLat, workflow.GeoFencingLat)
				assert.Equal(t, validWorkflow.GeoFencingLon, workflow.GeoFencingLon)
				assert.Equal(t, validWorkflow.GeoFencingZoomMin, workflow.GeoFencingZoomMin)
				assert.Equal(t, validWorkflow.GeoFencingZoomMax, workflow.GeoFencingZoomMax)
				assert.Equal(t, validWorkflow.IncludeReviews, workflow.IncludeReviews)
				assert.Equal(t, validWorkflow.IncludePhotos, workflow.IncludePhotos)
				assert.Equal(t, validWorkflow.IncludeBusinessHours, workflow.IncludeBusinessHours)
				assert.Equal(t, validWorkflow.MaxReviewsPerBusiness, workflow.MaxReviewsPerBusiness)
				assert.Equal(t, validWorkflow.RespectRobotsTxt, workflow.RespectRobotsTxt)
				assert.Equal(t, validWorkflow.AcceptTermsOfService, workflow.AcceptTermsOfService)
				assert.Equal(t, validWorkflow.UserAgent, workflow.UserAgent)
			},
		},
		{
			name:      "[failure scenario] - nil workflow",
			workflow:  nil,
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name: "[failure scenario] - invalid cron expression",
			workflow: &lead_scraper_servicev1.ScrapingWorkflow{
				CronExpression: "99 99 99 99 99", // Obviously invalid cron
				MaxRetries:    5,
				GeoFencingZoomMin: 1,
				GeoFencingZoomMax: 20,
			},
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name: "[failure scenario] - invalid geo fencing parameters",
			workflow: &lead_scraper_servicev1.ScrapingWorkflow{
				CronExpression:    "0 0 * * *",
				MaxRetries:       5,
				GeoFencingRadius: -1,
				GeoFencingLat:    91.0,  // Invalid latitude (> 90)
				GeoFencingLon:    181.0, // Invalid longitude (> 180)
				GeoFencingZoomMin: 1,
				GeoFencingZoomMax: 20,
			},
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name:      "[failure scenario] - context timeout",
			workflow:  validWorkflow,
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

			result, err := conn.CreateScrapingWorkflow(ctx, tt.workflow)

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

			// Clean up created workflow
			if result != nil {
				err := conn.DeleteScrapingWorkflow(context.Background(), result.Id)
				require.NoError(t, err)
			}
		})
	}
}

func TestCreateScrapingWorkflow_ConcurrentCreation(t *testing.T) {
	numWorkflows := 5
	var wg sync.WaitGroup
	errors := make(chan error, numWorkflows)
	workflows := make(chan *lead_scraper_servicev1.ScrapingWorkflow, numWorkflows)

	// Create workflows concurrently
	for i := 0; i < numWorkflows; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			workflow := &lead_scraper_servicev1.ScrapingWorkflow{
				CronExpression:         "0 0 * * *",
				RetryCount:            0,
				MaxRetries:            5,
				AlertEmails:           "test@example.com",
				GeoFencingRadius:     float32(1000.0),
				GeoFencingLat:        40.7128,
				GeoFencingLon:        -74.0060,
				GeoFencingZoomMin:    1,
				GeoFencingZoomMax:    20,
				IncludeReviews:       true,
				IncludePhotos:        true,
				IncludeBusinessHours: true,
				MaxReviewsPerBusiness: 100,
				RespectRobotsTxt:     true,
				AcceptTermsOfService:  true,
				UserAgent:            fmt.Sprintf("TestBot/%d", index),
			}

			created, err := conn.CreateScrapingWorkflow(context.Background(), workflow)
			if err != nil {
				errors <- err
				return
			}
			workflows <- created
		}(i)
	}

	wg.Wait()
	close(errors)
	close(workflows)

	// Clean up created workflows
	createdWorkflows := make([]*lead_scraper_servicev1.ScrapingWorkflow, 0)
	for workflow := range workflows {
		createdWorkflows = append(createdWorkflows, workflow)
	}

	defer func() {
		for _, workflow := range createdWorkflows {
			if workflow != nil {
				err := conn.DeleteScrapingWorkflow(context.Background(), workflow.Id)
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

	// Verify all workflows were created successfully
	require.Equal(t, numWorkflows, len(createdWorkflows))
	for _, workflow := range createdWorkflows {
		require.NotNil(t, workflow)
		require.NotZero(t, workflow.Id)
		assert.Equal(t, float32(1000.0), workflow.GeoFencingRadius)
		assert.Equal(t, float64(40.7128), workflow.GeoFencingLat)
		assert.Equal(t, float64(-74.0060), workflow.GeoFencingLon)
		assert.Equal(t, int32(1), workflow.GeoFencingZoomMin)
		assert.Equal(t, int32(20), workflow.GeoFencingZoomMax)
		assert.True(t, workflow.IncludeReviews)
		assert.True(t, workflow.IncludePhotos)
		assert.True(t, workflow.IncludeBusinessHours)
		assert.Equal(t, int32(100), workflow.MaxReviewsPerBusiness)
		assert.True(t, workflow.RespectRobotsTxt)
		assert.True(t, workflow.AcceptTermsOfService)
	}
}
