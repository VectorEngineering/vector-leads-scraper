package database

import (
	"context"
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
			leads[i] = testutils.GenerateRandomLead()
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
				for _, lead := range leads {
					// Fetch the lead from the database
					created, err := conn.GetLead(context.Background(), lead.Id)
					require.NoError(t, err)
					require.NotNil(t, created)
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
				tt.validate(t, result)
			}
		})
	}
}
