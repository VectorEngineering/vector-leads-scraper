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

func TestListLeads(t *testing.T) {
	// Clean up any existing leads first
	result := conn.Client.Engine.Exec("DELETE FROM leads")
	require.NoError(t, result.Error)

	// Also clean up any existing scraping jobs
	result = conn.Client.Engine.Exec("DELETE FROM gmaps_jobs")
	require.NoError(t, result.Error)

	// Create a test scraping job first
	testJob := testutils.GenerateRandomizedScrapingJob()

	createdJob, err := conn.CreateScrapingJob(context.Background(), testJob)
	require.NoError(t, err)
	require.NotNil(t, createdJob)

	// Create multiple test leads with explicit ordering
	numLeads := 5
	leadIDs := make([]uint64, 0, numLeads)

	// Create leads with a small delay to ensure ordering
	for i := 0; i < numLeads; i++ {
		lead := testutils.GenerateRandomLead()
		// Ensure each lead has a unique name for ordering
		lead.Name = fmt.Sprintf("Test Lead %d", i)

		created, err := conn.CreateLead(context.Background(), createdJob.Id, lead)
		require.NoError(t, err)
		require.NotNil(t, created)
		leadIDs = append(leadIDs, created.Id)

		// Small delay to ensure consistent ordering
		time.Sleep(time.Millisecond)
	}

	// Verify we have exactly the number of leads we expect
	var count int64
	result = conn.Client.Engine.Table("leads").Count(&count)
	require.NoError(t, result.Error)
	require.LessOrEqual(t, int64(numLeads), count, "Should have exactly %d leads in the database", numLeads)

	// Clean up after all tests
	defer func() {
		// Clean up leads
		for _, id := range leadIDs {
			err := conn.DeleteLead(context.Background(), id, DeletionTypeSoft)
			require.NoError(t, err)
		}
		// Clean up job
		if createdJob != nil {
			err := conn.DeleteScrapingJob(context.Background(), createdJob.Id)
			require.NoError(t, err)
		}
		// Final cleanup
		conn.Client.Engine.Exec("DELETE FROM leads")
		conn.Client.Engine.Exec("DELETE FROM gmaps_jobs")
	}()

	tests := []struct {
		name      string
		limit     int
		offset    int
		wantError bool
		errType   error
		validate  func(t *testing.T, leads []*lead_scraper_servicev1.Lead)
	}{
		{
			name:      "get all leads - success",
			limit:     10,
			offset:    0,
			wantError: false,
			validate: func(t *testing.T, leads []*lead_scraper_servicev1.Lead) {
				assert.Equal(t, numLeads, len(leads), "Should return exactly %d leads", numLeads)
				// Verify each lead ID is in our created set
				for _, lead := range leads {
					assert.NotNil(t, lead)
					assert.Contains(t, leadIDs, lead.Id, "Lead ID should be in the created set")
				}
				// Verify ordering
				for i := 1; i < len(leads); i++ {
					assert.True(t, leads[i].Id > leads[i-1].Id, "Leads should be ordered by ID")
				}
			},
		},
		{
			name:      "pagination first page - success",
			limit:     3,
			offset:    0,
			wantError: false,
			validate: func(t *testing.T, leads []*lead_scraper_servicev1.Lead) {
				assert.Equal(t, 3, len(leads), "Should return exactly 3 leads")
				for _, lead := range leads {
					assert.NotNil(t, lead)
					assert.Contains(t, leadIDs, lead.Id, "Lead ID should be in the created set")
				}
			},
		},
		{
			name:      "pagination second page - success",
			limit:     3,
			offset:    3,
			wantError: false,
			validate: func(t *testing.T, leads []*lead_scraper_servicev1.Lead) {
				assert.Equal(t, 2, len(leads), "Should return exactly 2 leads")
				for _, lead := range leads {
					assert.NotNil(t, lead)
					assert.Contains(t, leadIDs, lead.Id, "Lead ID should be in the created set")
				}
			},
		},
		{
			name:      "empty result - success",
			limit:     10,
			offset:    numLeads + 1,
			wantError: false,
			validate: func(t *testing.T, leads []*lead_scraper_servicev1.Lead) {
				assert.Empty(t, leads)
			},
		},
		{
			name:      "invalid limit - failure",
			limit:     -1,
			offset:    0,
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name:      "invalid offset - failure",
			limit:     10,
			offset:    -1,
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name:      "context timeout - failure",
			limit:     10,
			offset:    0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.name == "context timeout - failure" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()
				time.Sleep(2 * time.Millisecond)
			}

			results, err := conn.ListLeads(ctx, tt.limit, tt.offset)

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

func TestListLeads_EmptyDatabase(t *testing.T) {
	// Clean up any existing leads first
	result := conn.Client.Engine.Exec("DELETE FROM leads")
	require.NoError(t, result.Error)

	results, err := conn.ListLeads(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.NotNil(t, results)
	assert.Empty(t, results)
}
