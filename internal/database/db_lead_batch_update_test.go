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

// setupTestLeads creates a test job and leads for testing
func setupTestLeads(t *testing.T, numLeads int) (*lead_scraper_servicev1.ScrapingJob, []*lead_scraper_servicev1.Lead) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a test scraping job
	testJob := testutils.GenerateRandomizedScrapingJob()
	createdJob, err := conn.CreateScrapingJob(ctx, testJob)
	require.NoError(t, err)
	require.NotNil(t, createdJob)

	// Create test leads
	createdLeads := make([]*lead_scraper_servicev1.Lead, numLeads)
	for i := 0; i < numLeads; i++ {
		lead := testutils.GenerateRandomLead()
		created, err := conn.CreateLead(ctx, createdJob.Id, lead)
		require.NoError(t, err)
		require.NotNil(t, created)
		createdLeads[i] = created
	}

	return createdJob, createdLeads
}

// cleanupTestLeads cleans up test job and leads
func cleanupTestLeads(t *testing.T, job *lead_scraper_servicev1.ScrapingJob, leads []*lead_scraper_servicev1.Lead) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, lead := range leads {
		if lead != nil {
			err := conn.DeleteLead(ctx, lead.Id, DeletionTypeSoft)
			require.NoError(t, err)
		}
	}
	if job != nil {
		err := conn.DeleteScrapingJob(ctx, job.Id)
		require.NoError(t, err)
	}
}

// getUpdatedLeads returns a slice of updated leads for testing
func getUpdatedLeads(leads []*lead_scraper_servicev1.Lead) []*lead_scraper_servicev1.Lead {
	updatedLeads := make([]*lead_scraper_servicev1.Lead, len(leads))
	for i, lead := range leads {
		updatedLead := *lead
		updatedLead.Name = fmt.Sprintf("Updated Lead %d", i)
		updatedLead.Website = fmt.Sprintf("https://updated-lead-%d.com", i)
		updatedLead.Phone = fmt.Sprintf("+%d", 9876543210+i)
		updatedLead.Address = fmt.Sprintf("456 Updated St %d", i)
		updatedLead.City = "Updated City"
		updatedLead.State = "Updated State"
		updatedLead.Country = "Updated Country"
		updatedLead.Industry = "Updated Industry"
		updatedLead.PlaceId = fmt.Sprintf("ChIJ_updated%d", i)
		updatedLead.GoogleMapsUrl = "https://maps.google.com/?q=41.8781,-87.6298"
		updatedLead.Latitude = 41.8781
		updatedLead.Longitude = -87.6298
		updatedLead.GoogleRating = 4.8
		updatedLead.ReviewCount = 200
		updatedLeads[i] = &updatedLead
	}
	return updatedLeads
}

func TestBatchUpdateLeads(t *testing.T) {
	// Setup test data
	testJob, createdLeads := setupTestLeads(t, 10)
	defer cleanupTestLeads(t, testJob, createdLeads)

	tests := []struct {
		name      string
		leads     []*lead_scraper_servicev1.Lead
		wantError bool
		errType   error
		setup     func(t *testing.T) []*lead_scraper_servicev1.Lead
		validate  func(t *testing.T, leads []*lead_scraper_servicev1.Lead)
	}{
		{
			name: "[success scenario] - update all leads",
			setup: func(t *testing.T) []*lead_scraper_servicev1.Lead {
				return getUpdatedLeads(createdLeads)
			},
			wantError: false,
			validate: func(t *testing.T, leads []*lead_scraper_servicev1.Lead) {
				require.NotEmpty(t, leads)
				for i, lead := range leads {
					assert.NotNil(t, lead)
					assert.Equal(t, fmt.Sprintf("Updated Lead %d", i), lead.Name)
					assert.Equal(t, fmt.Sprintf("https://updated-lead-%d.com", i), lead.Website)
					assert.Equal(t, fmt.Sprintf("+%d", 9876543210+i), lead.Phone)
					assert.Equal(t, fmt.Sprintf("456 Updated St %d", i), lead.Address)
					assert.Equal(t, "Updated City", lead.City)
					assert.Equal(t, "Updated State", lead.State)
					assert.Equal(t, "Updated Country", lead.Country)
					assert.Equal(t, "Updated Industry", lead.Industry)
				}
			},
		},
		{
			name:      "[failure scenario] - nil leads",
			leads:     nil,
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name:      "[failure scenario] - empty leads slice",
			leads:     []*lead_scraper_servicev1.Lead{},
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name: "[failure scenario] - context timeout",
			setup: func(t *testing.T) []*lead_scraper_servicev1.Lead {
				return createdLeads
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var leads []*lead_scraper_servicev1.Lead
			if tt.setup != nil {
				leads = tt.setup(t)
			} else {
				leads = tt.leads
			}

			ctx := context.Background()
			if tt.name == "[failure scenario] - context timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()
				time.Sleep(2 * time.Millisecond)
			}

			success, err := conn.BatchUpdateLeads(ctx, leads)

			if tt.wantError {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.False(t, success)
				return
			}

			require.NoError(t, err)
			assert.True(t, success)

			// If validation is needed, fetch the updated leads and validate them
			if tt.validate != nil {
				// Fetch the updated leads to validate them
				updatedLeads := make([]*lead_scraper_servicev1.Lead, len(leads))
				for i, lead := range leads {
					updated, err := conn.GetLead(ctx, lead.Id)
					require.NoError(t, err)
					updatedLeads[i] = updated
				}
				tt.validate(t, updatedLeads)
			}
		})
	}
}

func TestBatchUpdateLeads_LargeBatch(t *testing.T) {
	// Setup test data
	testJob, createdLeads := setupTestLeads(t, 1000)
	defer cleanupTestLeads(t, testJob, createdLeads)

	// Prepare updates
	updatedLeads := getUpdatedLeads(createdLeads)

	// Perform batch update
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	success, err := conn.BatchUpdateLeads(ctx, updatedLeads)
	require.NoError(t, err)
	assert.True(t, success)

	// Verify all leads were updated correctly by fetching them
	for i, originalLead := range createdLeads {
		updated, err := conn.GetLead(ctx, originalLead.Id)
		require.NoError(t, err)
		require.NotNil(t, updated)

		// Use the original lead's index to construct expected values
		assert.Equal(t, fmt.Sprintf("Updated Lead %d", i), updated.Name)
		assert.Equal(t, fmt.Sprintf("https://updated-lead-%d.com", i), updated.Website)
		assert.Equal(t, fmt.Sprintf("+%d", 9876543210+i), updated.Phone)
		assert.Equal(t, fmt.Sprintf("456 Updated St %d", i), updated.Address)
		assert.Equal(t, "Updated City", updated.City)
		assert.Equal(t, "Updated State", updated.State)
		assert.Equal(t, "Updated Country", updated.Country)
		assert.Equal(t, "Updated Industry", updated.Industry)
		assert.Equal(t, fmt.Sprintf("ChIJ_updated%d", i), updated.PlaceId)
		assert.Equal(t, "https://maps.google.com/?q=41.8781,-87.6298", updated.GoogleMapsUrl)
		assert.Equal(t, float64(41.8781), updated.Latitude)
		assert.Equal(t, float64(-87.6298), updated.Longitude)
		assert.Equal(t, float32(4.8), updated.GoogleRating)
	}
} 