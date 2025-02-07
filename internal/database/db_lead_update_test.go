package database

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpdateLead(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create a test scraping job first
	testJob := testutils.GenerateRandomizedScrapingJob()

	createdJob, err := conn.CreateScrapingJob(context.Background(), tc.Workspace.Id, testJob)
	require.NoError(t, err)
	require.NotNil(t, createdJob)

	// Create a test lead
	testLead := testutils.GenerateRandomLead()

	createdLead, err := conn.CreateLead(context.Background(), createdJob.Id, testLead)
	require.NoError(t, err)
	require.NotNil(t, createdLead)

	// Clean up after all tests
	defer func() {
		if createdLead != nil {
			err := conn.DeleteLead(context.Background(), createdLead.Id, DeletionTypeSoft)
			require.NoError(t, err)
		}
		if createdJob != nil {
			err := conn.DeleteScrapingJob(context.Background(), createdJob.Id)
			require.NoError(t, err)
		}
	}()

	tests := []struct {
		name      string
		lead      *lead_scraper_servicev1.Lead
		wantError bool
		errType   error
		setup     func(t *testing.T) *lead_scraper_servicev1.Lead
		validate  func(t *testing.T, lead *lead_scraper_servicev1.Lead)
	}{
		{
			name: "[success scenario] - valid update",
			setup: func(t *testing.T) *lead_scraper_servicev1.Lead {
				updatedLead := proto.Clone(createdLead).(*lead_scraper_servicev1.Lead)
				updatedLead.Name = "Updated Lead"
				updatedLead.Website = "https://updated-lead.com"
				updatedLead.Phone = "+9876543210"
				updatedLead.Address = "456 Updated St"
				updatedLead.City = "Updated City"
				updatedLead.State = "Updated State"
				updatedLead.Country = "Updated Country"
				updatedLead.Industry = "Updated Industry"
				updatedLead.PlaceId = "ChIJ_updated123"
				updatedLead.GoogleMapsUrl = "https://maps.google.com/?q=41.8781,-87.6298"
				updatedLead.Latitude = 41.8781
				updatedLead.Longitude = -87.6298
				updatedLead.GoogleRating = 4.8
				updatedLead.ReviewCount = 200
				return updatedLead
			},
			wantError: false,
			validate: func(t *testing.T, lead *lead_scraper_servicev1.Lead) {
				assert.NotNil(t, lead)
				assert.Equal(t, createdLead.Id, lead.Id)
				assert.Equal(t, "Updated Lead", lead.Name)
				assert.Equal(t, "https://updated-lead.com", lead.Website)
				assert.Equal(t, "+9876543210", lead.Phone)
				assert.Equal(t, "456 Updated St", lead.Address)
				assert.Equal(t, "Updated City", lead.City)
				assert.Equal(t, "Updated State", lead.State)
				assert.Equal(t, "Updated Country", lead.Country)
				assert.Equal(t, "Updated Industry", lead.Industry)
				assert.Equal(t, "ChIJ_updated123", lead.PlaceId)
				assert.Equal(t, "https://maps.google.com/?q=41.8781,-87.6298", lead.GoogleMapsUrl)
				assert.Equal(t, float64(41.8781), lead.Latitude)
				assert.Equal(t, float64(-87.6298), lead.Longitude)
				assert.Equal(t, float32(4.8), lead.GoogleRating)
				assert.Equal(t, int32(200), lead.ReviewCount)
			},
		},
		{
			name:      "[failure scenario] - nil lead",
			lead:      nil,
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name: "[failure scenario] - zero id",
			setup: func(t *testing.T) *lead_scraper_servicev1.Lead {
				invalidLead := proto.Clone(createdLead).(*lead_scraper_servicev1.Lead)
				invalidLead.Id = 0
				return invalidLead
			},
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name: "[failure scenario] - non-existent id",
			setup: func(t *testing.T) *lead_scraper_servicev1.Lead {
				invalidLead := proto.Clone(createdLead).(*lead_scraper_servicev1.Lead)
				invalidLead.Id = 999999
				return invalidLead
			},
			wantError: true,
			errType:   ErrJobDoesNotExist,
		},
		{
			name: "[failure scenario] - context timeout",
			setup: func(t *testing.T) *lead_scraper_servicev1.Lead {
				return createdLead
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lead *lead_scraper_servicev1.Lead
			if tt.setup != nil {
				lead = tt.setup(t)
			} else {
				lead = tt.lead
			}

			ctx := context.Background()
			if tt.name == "[failure scenario] - context timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()
				time.Sleep(2 * time.Millisecond)
			}

			result, err := conn.UpdateLead(ctx, lead)

			if tt.wantError {
				require.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestUpdateLead_ConcurrentUpdates(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create a test scraping job first
	testJob := testutils.GenerateRandomizedScrapingJob()
	createdJob, err := conn.CreateScrapingJob(context.Background(), tc.Workspace.Id, testJob)
	require.NoError(t, err)
	require.NotNil(t, createdJob)

	// Create test leads
	numLeads := 5
	createdLeads := make([]*lead_scraper_servicev1.Lead, numLeads)
	for i := 0; i < numLeads; i++ {
		lead := testutils.GenerateRandomLead()
		created, err := conn.CreateLead(context.Background(), createdJob.Id, lead)
		require.NoError(t, err)
		require.NotNil(t, created)
		createdLeads[i] = created
	}

	// Clean up after test
	defer func() {
		for _, lead := range createdLeads {
			if lead != nil {
				err := conn.DeleteLead(context.Background(), lead.Id, DeletionTypeSoft)
				require.NoError(t, err)
			}
		}
		err := conn.DeleteScrapingJob(context.Background(), createdJob.Id)
		require.NoError(t, err)
	}()

	var wg sync.WaitGroup
	errors := make(chan error, numLeads)

	// Update leads concurrently
	for i := 0; i < numLeads; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			lead := createdLeads[index]
			lead.Name = fmt.Sprintf("Updated Lead %d", index)
			lead.Website = fmt.Sprintf("https://updated-lead-%d.com", index)

			_, err := conn.UpdateLead(context.Background(), lead)
			if err != nil {
				errors <- err
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}
	require.Empty(t, errs, "Expected no errors during concurrent updates, got: %v", errs)

	// Verify all leads were updated successfully
	for i, lead := range createdLeads {
		updated, err := conn.GetLead(context.Background(), lead.Id)
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, fmt.Sprintf("Updated Lead %d", i), updated.Name)
		assert.Equal(t, fmt.Sprintf("https://updated-lead-%d.com", i), updated.Website)
	}
}
