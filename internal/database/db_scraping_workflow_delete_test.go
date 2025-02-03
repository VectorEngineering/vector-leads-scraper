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

func TestDeleteScrapingWorkflow(t *testing.T) {
	// Create a test workflow first
	testWorkflow := testutils.GenerateRandomScrapingWorkflow()

	created, err := conn.CreateScrapingWorkflow(context.Background(), testWorkflow)
	require.NoError(t, err)
	require.NotNil(t, created)

	tests := []struct {
		name      string
		id        uint64
		wantError bool
		errType   error
		setup     func(t *testing.T) uint64
		validate  func(t *testing.T, id uint64)
	}{
		{
			name:      "[success scenario] - valid id",
			id:        created.Id,
			wantError: false,
			validate: func(t *testing.T, id uint64) {
				// Verify the workflow was deleted
				_, err := conn.GetScrapingWorkflow(context.Background(), id)
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrWorkflowDoesNotExist)
			},
		},
		{
			name:      "[failure scenario] - invalid id",
			id:        0,
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name:      "[failure scenario] - non-existent id",
			id:        999999,
			wantError: true,
			errType:   ErrWorkflowDoesNotExist,
		},
		{
			name:      "[failure scenario] - already deleted workflow",
			wantError: true,
			errType:   ErrWorkflowDoesNotExist,
			setup: func(t *testing.T) uint64 {
				// Create and delete a workflow
				workflow := &lead_scraper_servicev1.ScrapingWorkflow{
					CronExpression: "0 0 * * *",
					MaxRetries:    5,
					GeoFencingZoomMin: 1,
					GeoFencingZoomMax: 20,
				}
				created, err := conn.CreateScrapingWorkflow(context.Background(), workflow)
				require.NoError(t, err)
				require.NotNil(t, created)

				err = conn.DeleteScrapingWorkflow(context.Background(), created.Id)
				require.NoError(t, err)

				return created.Id
			},
		},
		{
			name:      "[failure scenario] - context timeout",
			id:        created.Id,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id uint64
			if tt.setup != nil {
				id = tt.setup(t)
			} else {
				id = tt.id
			}

			ctx := context.Background()
			if tt.name == "[failure scenario] - context timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()
				time.Sleep(5 * time.Millisecond)
			}

			err := conn.DeleteScrapingWorkflow(ctx, id)

			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, id)
			}
		})
	}
}

func TestDeleteScrapingWorkflow_ConcurrentDeletions(t *testing.T) {
	numWorkflows := 5
	var wg sync.WaitGroup
	errors := make(chan error, numWorkflows)
	workflowIDs := make([]uint64, numWorkflows)

	// Create test workflows
	for i := 0; i < numWorkflows; i++ {
		workflow := &lead_scraper_servicev1.ScrapingWorkflow{
			CronExpression:         "0 0 * * *",
			RetryCount:            0,
			MaxRetries:            5,
			AlertEmails:           "test@example.com",
			GeoFencingRadius:     1000.0,
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
		created, err := conn.CreateScrapingWorkflow(context.Background(), workflow)
		require.NoError(t, err)
		require.NotNil(t, created)
		workflowIDs[i] = created.Id
	}

	// Delete workflows concurrently
	for i := 0; i < numWorkflows; i++ {
		wg.Add(1)
		go func(id uint64) {
			defer wg.Done()
			if err := conn.DeleteScrapingWorkflow(context.Background(), id); err != nil {
				errors <- err
			}
		}(workflowIDs[i])
	}

	wg.Wait()
	close(errors)

	// Check for errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}
	require.Empty(t, errs, "Expected no errors during concurrent deletions, got: %v", errs)

	// Verify all workflows were deleted
	for _, id := range workflowIDs {
		_, err := conn.GetScrapingWorkflow(context.Background(), id)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrWorkflowDoesNotExist)
	}
}

func TestBatchDeleteScrapingWorkflows(t *testing.T) {
	// Create test workflows first
	numWorkflows := 5
	workflowIDs := make([]uint64, numWorkflows)
	for i := 0; i < numWorkflows; i++ {
		testWorkflow := testutils.GenerateRandomScrapingWorkflow()
		created, err := conn.CreateScrapingWorkflow(context.Background(), testWorkflow)
		require.NoError(t, err)
		require.NotNil(t, created)
		workflowIDs[i] = created.Id
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
			ids:       workflowIDs,
			wantError: false,
			validate: func(t *testing.T, ids []uint64) {
				// Verify all workflows were deleted
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				for _, id := range ids {
					_, err := conn.GetScrapingWorkflow(ctx, id)
					assert.Error(t, err)
					assert.ErrorIs(t, err, ErrWorkflowDoesNotExist)
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
			errType:   ErrWorkflowDoesNotExist,
			setup: func(t *testing.T) []uint64 {
				// Create one workflow and combine with non-existent ID
				workflow := testutils.GenerateRandomScrapingWorkflow()
				created, err := conn.CreateScrapingWorkflow(context.Background(), workflow)
				require.NoError(t, err)
				require.NotNil(t, created)
				return []uint64{created.Id, 999999}
			},
		},
		{
			name:      "[failure scenario] - context timeout",
			ids:       workflowIDs,
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

			err := conn.BatchDeleteScrapingWorkflows(ctx, ids)

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

func TestBatchDeleteScrapingWorkflows_LargeBatch(t *testing.T) {
	// Create a large batch of workflows
	numWorkflows := 100
	workflowIDs := make([]uint64, numWorkflows)
	for i := 0; i < numWorkflows; i++ {
		workflow := testutils.GenerateRandomScrapingWorkflow()
		created, err := conn.CreateScrapingWorkflow(context.Background(), workflow)
		require.NoError(t, err)
		require.NotNil(t, created)
		workflowIDs[i] = created.Id
	}

	// Delete workflows in batch
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := conn.BatchDeleteScrapingWorkflows(ctx, workflowIDs)
	require.NoError(t, err)

	// Verify all workflows were deleted
	for _, workflowID := range workflowIDs {
		_, err := conn.GetScrapingWorkflow(ctx, workflowID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrWorkflowDoesNotExist)
	}
}
