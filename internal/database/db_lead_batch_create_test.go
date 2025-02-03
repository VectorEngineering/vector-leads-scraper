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

func TestBatchCreateLeads(t *testing.T) {
	// Create a test job first
	// Create a test scraping job first
	testJob := testutils.GenerateRandomizedScrapingJob()

	createdJob, err := conn.CreateScrapingJob(context.Background(), testJob)
	require.NoError(t, err)
	require.NotNil(t, createdJob)

	// Helper function to create test leads
	createTestLeads := func(count int) []*lead_scraper_servicev1.Lead {
		leads := make([]*lead_scraper_servicev1.Lead, count)
		for i := 0; i < count; i++ {
			leads[i] = &lead_scraper_servicev1.Lead{
				Name:         fmt.Sprintf("Test Lead %d", i),
				Website:      fmt.Sprintf("https://test-lead-%d.com", i),
				Phone:        fmt.Sprintf("+%d", 1234567890+i),
				Address:      fmt.Sprintf("123 Test St %d", i),
				City:         "Test City",
				State:        "Test State",
				Country:      "Test Country",
				Industry:     "Test Industry",
				PlaceId:      fmt.Sprintf("ChIJ_test%d", i),
				GoogleMapsUrl: "https://maps.google.com/?q=40.7128,-74.0060",
				Latitude:     40.7128,
				Longitude:    -74.0060,
				GoogleRating: 4.5,
				ReviewCount:  100,
			}
		}
		return leads
	}

	tests := []struct {
		name      string
		leads     []*lead_scraper_servicev1.Lead
		wantError bool
		errType   error
		setup     func(t *testing.T) []*lead_scraper_servicev1.Lead
		validate  func(t *testing.T, leads []*lead_scraper_servicev1.Lead)
	}{
		{
			name: "success - create small batch",
			setup: func(t *testing.T) []*lead_scraper_servicev1.Lead {
				return createTestLeads(10)
			},
			validate: func(t *testing.T, leads []*lead_scraper_servicev1.Lead) {
				// Verify each lead was created correctly
				for i, lead := range leads {
					// Fetch the lead from the database
					created, err := conn.GetLead(context.Background(), lead.Id)
					require.NoError(t, err)
					require.NotNil(t, created)

					// Verify all fields match
					assert.Equal(t, fmt.Sprintf("Test Lead %d", i), created.Name)
					assert.Equal(t, fmt.Sprintf("https://test-lead-%d.com", i), created.Website)
					assert.Equal(t, fmt.Sprintf("+%d", 1234567890+i), created.Phone)
					assert.Equal(t, fmt.Sprintf("123 Test St %d", i), created.Address)
					assert.Equal(t, "Test City", created.City)
					assert.Equal(t, "Test State", created.State)
					assert.Equal(t, "Test Country", created.Country)
					assert.Equal(t, "Test Industry", created.Industry)
					assert.Equal(t, fmt.Sprintf("ChIJ_test%d", i), created.PlaceId)
					assert.Equal(t, "https://maps.google.com/?q=40.7128,-74.0060", created.GoogleMapsUrl)
					assert.Equal(t, float64(40.7128), created.Latitude)
					assert.Equal(t, float64(-74.0060), created.Longitude)
					assert.Equal(t, float32(4.5), created.GoogleRating)
					assert.Equal(t, int32(100), created.ReviewCount)
				}
			},
		},
		{
			name: "success - create large batch",
			setup: func(t *testing.T) []*lead_scraper_servicev1.Lead {
				return createTestLeads(1000)
			},
			validate: func(t *testing.T, leads []*lead_scraper_servicev1.Lead) {
				// Verify a sample of leads
				for i := 0; i < len(leads); i += 100 {
					created, err := conn.GetLead(context.Background(), leads[i].Id)
					require.NoError(t, err)
					require.NotNil(t, created)
					assert.Equal(t, fmt.Sprintf("Test Lead %d", i), created.Name)
				}
			},
		},
		{
			name:      "failure - nil leads",
			leads:     nil,
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name:      "failure - empty leads slice",
			leads:     []*lead_scraper_servicev1.Lead{},
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name: "failure - invalid job ID",
			setup: func(t *testing.T) []*lead_scraper_servicev1.Lead {
				leads := createTestLeads(10) // Non-existent job ID
				return leads
			},
			wantError: true,
		},
		{
			name: "failure - context timeout",
			setup: func(t *testing.T) []*lead_scraper_servicev1.Lead {
				return createTestLeads(10)
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
			if tt.name == "failure - context timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()
				time.Sleep(2 * time.Millisecond)
			}

			result, err := conn.BatchCreateLeads(ctx, createdJob.Id, leads)

			if tt.wantError {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result)

			// Run validation if provided
			if tt.validate != nil {
				tt.validate(t, leads)
			}
		})
	}
} 