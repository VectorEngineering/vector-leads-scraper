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

func setupTestLeadAndJob(t *testing.T) (*lead_scraper_servicev1.ScrapingJob, *lead_scraper_servicev1.Lead) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create a test scraping job
	testJob := testutils.GenerateRandomizedScrapingJob()
	createdJob, err := conn.CreateScrapingJob(context.Background(), tc.Workspace.Id, testJob)
	require.NoError(t, err)
	require.NotNil(t, createdJob)

	// Create a test lead
	testLead := testutils.GenerateRandomLead()
	createdLead, err := conn.CreateLead(context.Background(), createdJob.Id, testLead)
	require.NoError(t, err)
	require.NotNil(t, createdLead)

	return createdJob, createdLead
}

func TestDeleteLead(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	createdJob, createdLead := setupTestLeadAndJob(t)

	// Clean up job after all tests
	defer func() {
		if createdJob != nil {
			err := conn.DeleteScrapingJob(context.Background(), createdJob.Id)
			require.NoError(t, err)
		}
	}()

	tests := []struct {
		name         string
		id           uint64
		deletionType DeletionType
		wantError    bool
		errType      error
		setup        func(t *testing.T) uint64
		validate     func(t *testing.T, id uint64)
	}{
		{
			name:         "soft delete - success",
			id:           createdLead.Id,
			deletionType: DeletionTypeSoft,
			wantError:    false,
			validate: func(t *testing.T, id uint64) {
				// Verify the lead was soft deleted
				_, err := conn.GetLead(context.Background(), id)
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrJobDoesNotExist)
			},
		},
		{
			name: "hard delete - success",
			setup: func(t *testing.T) uint64 {
				// Create a new lead for hard delete
				lead := testutils.GenerateRandomLead()
				created, err := conn.CreateLead(context.Background(), createdJob.Id, lead)
				require.NoError(t, err)
				require.NotNil(t, created)
				return created.Id
			},
			deletionType: DeletionTypeHard,
			wantError:    false,
			validate: func(t *testing.T, id uint64) {
				// Verify the lead was hard deleted
				_, err := conn.GetLead(context.Background(), id)
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrJobDoesNotExist)
			},
		},
		{
			name:         "invalid id - failure",
			id:           0,
			deletionType: DeletionTypeSoft,
			wantError:    true,
			errType:      ErrInvalidInput,
		},
		{
			name:         "already deleted lead - failure",
			deletionType: DeletionTypeSoft,
			wantError:    true,
			errType:      ErrJobDoesNotExist,
			setup: func(t *testing.T) uint64 {
				// Create and delete a lead
				lead := testutils.GenerateRandomLead()
				created, err := conn.CreateLead(context.Background(), createdJob.Id, lead)
				require.NoError(t, err)
				require.NotNil(t, created)

				err = conn.DeleteLead(context.Background(), created.Id, DeletionTypeSoft)
				require.NoError(t, err)

				return created.Id
			},
		},
		{
			name:         "context timeout - failure",
			id:           createdLead.Id,
			deletionType: DeletionTypeSoft,
			wantError:    true,
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
			if tt.name == "context timeout - failure" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()
				time.Sleep(2 * time.Millisecond)
			}

			err := conn.DeleteLead(ctx, id, tt.deletionType)

			if tt.wantError {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, id)
			}
		})
	}
}

func TestDeleteLead_ConcurrentDeletions(t *testing.T) {
	createdJob, _ := setupTestLeadAndJob(t)

	// Create test leads
	numLeads := 5
	createdLeads := make([]uint64, numLeads)
	for i := 0; i < numLeads; i++ {
		lead := testutils.GenerateRandomLead()
		created, err := conn.CreateLead(context.Background(), createdJob.Id, lead)
		require.NoError(t, err)
		require.NotNil(t, created)
		createdLeads[i] = created.Id
	}

	// Clean up job after test
	defer func() {
		err := conn.DeleteScrapingJob(context.Background(), createdJob.Id)
		require.NoError(t, err)
	}()

	var wg sync.WaitGroup
	errors := make(chan error, numLeads)

	// Delete leads concurrently
	for i := 0; i < numLeads; i++ {
		wg.Add(1)
		go func(id uint64) {
			defer wg.Done()
			if err := conn.DeleteLead(context.Background(), id, DeletionTypeSoft); err != nil {
				errors <- err
			}
		}(createdLeads[i])
	}

	wg.Wait()
	close(errors)

	// Check for errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}
	require.Empty(t, errs, "Expected no errors during concurrent deletions, got: %v", errs)

	// Verify all leads were deleted
	for _, id := range createdLeads {
		_, err := conn.GetLead(context.Background(), id)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrJobDoesNotExist)
	}
}
