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

func TestGetLead(t *testing.T) {
	createdJob, createdLead := setupTestLeadAndJob(t)

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
		id        uint64
		wantError bool
		errType   error
		setup     func(t *testing.T) uint64
		validate  func(t *testing.T, lead *lead_scraper_servicev1.Lead)
	}{
		{
			name:      "valid id - success",
			id:        createdLead.Id,
			wantError: false,
			validate: func(t *testing.T, lead *lead_scraper_servicev1.Lead) {
				assert.NotNil(t, lead)
				assert.Equal(t, createdLead.Id, lead.Id)
				assert.Equal(t, createdLead.Name, lead.Name)
				assert.Equal(t, createdLead.Website, lead.Website)
				assert.Equal(t, createdLead.Phone, lead.Phone)
				assert.Equal(t, createdLead.Address, lead.Address)
				assert.Equal(t, createdLead.City, lead.City)
				assert.Equal(t, createdLead.State, lead.State)
				assert.Equal(t, createdLead.Country, lead.Country)
				assert.Equal(t, createdLead.Industry, lead.Industry)
				assert.Equal(t, createdLead.PlaceId, lead.PlaceId)
				assert.Equal(t, createdLead.GoogleMapsUrl, lead.GoogleMapsUrl)
				assert.Equal(t, createdLead.Latitude, lead.Latitude)
				assert.Equal(t, createdLead.Longitude, lead.Longitude)
				assert.Equal(t, createdLead.GoogleRating, lead.GoogleRating)
				assert.Equal(t, createdLead.ReviewCount, lead.ReviewCount)
			},
		},
		{
			name:      "invalid id - failure",
			id:        0,
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name:      "non-existent id - failure",
			id:        999999,
			wantError: true,
			errType:   ErrJobDoesNotExist,
		},
		{
			name:      "context timeout - failure",
			id:        createdLead.Id,
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

			result, err := conn.GetLead(ctx, tt.id)

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
		})
	}
}

func TestGetLead_ConcurrentReads(t *testing.T) {
	createdJob, createdLead := setupTestLeadAndJob(t)

	// Clean up after test
	defer func() {
		err := conn.DeleteLead(context.Background(), createdLead.Id, DeletionTypeSoft)
		require.NoError(t, err)
		err = conn.DeleteScrapingJob(context.Background(), createdJob.Id)
		require.NoError(t, err)
	}()

	numReads := 10
	var wg sync.WaitGroup
	errors := make(chan error, numReads)
	results := make(chan *lead_scraper_servicev1.Lead, numReads)

	// Perform concurrent reads
	for i := 0; i < numReads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lead, err := conn.GetLead(context.Background(), createdLead.Id)
			if err != nil {
				errors <- err
				return
			}
			results <- lead
		}()
	}

	wg.Wait()
	close(errors)
	close(results)

	// Check for errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}
	require.Empty(t, errs, "Expected no errors during concurrent reads, got: %v", errs)

	// Verify all reads returned the same data
	for lead := range results {
		assert.Equal(t, createdLead.Id, lead.Id)
		assert.Equal(t, createdLead.Name, lead.Name)
		assert.Equal(t, createdLead.Website, lead.Website)
		assert.Equal(t, createdLead.Phone, lead.Phone)
		assert.Equal(t, createdLead.Address, lead.Address)
	}
}
