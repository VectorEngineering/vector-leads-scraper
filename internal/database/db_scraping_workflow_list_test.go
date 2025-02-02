package database

import (
	"context"
	"testing"
	"time"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListScrapingWorkflows(t *testing.T) {
	// Clean up any existing workflows first
	existingWorkflows, err := conn.ListScrapingWorkflows(context.Background(), 100, 0)
	require.NoError(t, err)
	for _, w := range existingWorkflows {
		err := conn.DeleteScrapingWorkflow(context.Background(), w.Id)
		require.NoError(t, err)
	}

	// Create multiple test workflows
	workflows := []*lead_scraper_servicev1.ScrapingWorkflow{
		{
			CronExpression:       "0 0 * * *",
			RetryCount:           0,
			MaxRetries:           5,
			AlertEmails:          "test1@example.com",
			OrgId:               "test-org-1",
			TenantId:            "test-tenant-1",
			GeoFencingRadius:    1000.0,
			GeoFencingLat:       40.7128,
			GeoFencingLon:       -74.0060,
			GeoFencingZoomMin:    1,
			GeoFencingZoomMax:    20,
			NotificationWebhookUrl: "https://example.com/webhook1",
		},
		{
			CronExpression:       "0 12 * * *",
			RetryCount:           1,
			MaxRetries:           3,
			AlertEmails:          "test2@example.com",
			OrgId:               "test-org-2",
			TenantId:            "test-tenant-2",
			GeoFencingRadius:    2000.0,
			GeoFencingLat:       41.8781,
			GeoFencingLon:       -87.6298,
			GeoFencingZoomMin:    1,
			GeoFencingZoomMax:    20,
			NotificationWebhookUrl: "https://example.com/webhook2",
		},
		{
			CronExpression:       "0 0 1 * *",
			RetryCount:           2,
			MaxRetries:           4,
			AlertEmails:          "test3@example.com",
			OrgId:               "test-org-3",
			TenantId:            "test-tenant-3",
			GeoFencingRadius:    3000.0,
			GeoFencingLat:       34.0522,
			GeoFencingLon:       -118.2437,
			GeoFencingZoomMin:    1,
			GeoFencingZoomMax:    20,
			NotificationWebhookUrl: "https://example.com/webhook3",
		},
	}

	createdWorkflows := make([]*lead_scraper_servicev1.ScrapingWorkflow, 0, len(workflows))
	for _, w := range workflows {
		created, err := conn.CreateScrapingWorkflow(context.Background(), w)
		require.NoError(t, err, "Failed to create test workflow")
		require.NotNil(t, created, "Created workflow is nil")
		createdWorkflows = append(createdWorkflows, created)
		// Add a small delay to ensure ordering by creation time
		time.Sleep(10 * time.Millisecond)
	}

	// Clean up after all tests
	defer func() {
		for _, w := range createdWorkflows {
			if w != nil {
				err := conn.DeleteScrapingWorkflow(context.Background(), w.Id)
				if err != nil {
					t.Logf("Failed to delete test workflow: %v", err)
				}
			}
		}
	}()

	tests := []struct {
		name      string
		limit     int
		offset    int
		wantCount int
		wantError bool
		errType   error
	}{
		{
			name:      "[success scenario] - get all workflows",
			limit:     10,
			offset:    0,
			wantCount: len(createdWorkflows),
			wantError: false,
		},
		{
			name:      "[success scenario] - pagination first page",
			limit:     2,
			offset:    0,
			wantCount: 2,
			wantError: false,
		},
		{
			name:      "[success scenario] - pagination second page",
			limit:     2,
			offset:    2,
			wantCount: 1,
			wantError: false,
		},
		{
			name:      "[success scenario] - empty result",
			limit:     10,
			offset:    100,
			wantCount: 0,
			wantError: false,
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
				time.Sleep(5 * time.Millisecond)
			}

			results, err := conn.ListScrapingWorkflows(ctx, tt.limit, tt.offset)

			if tt.wantError {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, results)
			assert.Len(t, results, tt.wantCount)

			if len(results) > 0 {
				// Verify the first result has all required fields
				first := results[0]
				assert.NotZero(t, first.Id)
				assert.NotEmpty(t, first.CronExpression)
				assert.NotEmpty(t, first.AlertEmails)
				assert.NotEmpty(t, first.OrgId)
				assert.NotEmpty(t, first.TenantId)

				// For pagination test cases, verify we get the expected workflows in order
				if tt.name == "[success scenario] - pagination first page" {
					assert.Equal(t, createdWorkflows[0].Id, results[0].Id)
					assert.Equal(t, createdWorkflows[1].Id, results[1].Id)
				} else if tt.name == "[success scenario] - pagination second page" {
					assert.Equal(t, createdWorkflows[2].Id, results[0].Id)
				}
			}
		})
	}
}

func TestListScrapingWorkflows_EmptyDatabase(t *testing.T) {
	// Clean up any existing workflows first
	existingWorkflows, err := conn.ListScrapingWorkflows(context.Background(), 100, 0)
	require.NoError(t, err)
	for _, w := range existingWorkflows {
		err := conn.DeleteScrapingWorkflow(context.Background(), w.Id)
		require.NoError(t, err)
	}

	results, err := conn.ListScrapingWorkflows(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Empty(t, results)
}
