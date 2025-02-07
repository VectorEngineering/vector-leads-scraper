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
	ctx := context.Background()

	// Clean up any existing data
	cleanup := func() {
		// Check counts before cleanup
		var beforeCount int64
		result := conn.Client.Engine.Table("leads").Count(&beforeCount)
		require.NoError(t, result.Error)
		t.Logf("Before cleanup: %d leads", beforeCount)

		var beforeUnscopedCount int64
		result = conn.Client.Engine.Unscoped().Table("leads").Count(&beforeUnscopedCount)
		require.NoError(t, result.Error)
		t.Logf("Before cleanup (unscoped): %d leads", beforeUnscopedCount)

		result = conn.Client.Engine.Unscoped().Exec("DELETE FROM leads")
		require.NoError(t, result.Error)
		result = conn.Client.Engine.Unscoped().Exec("DELETE FROM gmaps_jobs")
		require.NoError(t, result.Error)

		// Verify cleanup worked
		var afterCount int64
		result = conn.Client.Engine.Table("leads").Count(&afterCount)
		require.NoError(t, result.Error)
		require.Equal(t, int64(0), afterCount, "Should have no leads after cleanup")

		var afterUnscopedCount int64
		result = conn.Client.Engine.Unscoped().Table("leads").Count(&afterUnscopedCount)
		require.NoError(t, result.Error)
		require.Equal(t, int64(0), afterUnscopedCount, "Should have no leads after cleanup (unscoped)")
	}
	cleanup()
	defer cleanup()

	// Create a test organization and tenant first
	org := testutils.GenerateRandomizedOrganization()
	createdOrg, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{Organization: org})
	require.NoError(t, err)
	require.NotNil(t, createdOrg)

	tenant := testutils.GenerateRandomizedTenant()
	createdTenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
		Tenant:         tenant,
		OrganizationID: createdOrg.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createdTenant)

	// create an account for the tenant
	account := testutils.GenerateRandomizedAccount()
	createdAccount, err := conn.CreateAccount(ctx, &CreateAccountInput{
		Account: account,
		TenantID: createdTenant.Id,
		OrgID: createdOrg.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createdAccount)

	// create a workspace for the account
	workspace := testutils.GenerateRandomWorkspace()
	createdWorkspace, err := conn.CreateWorkspace(ctx, &CreateWorkspaceInput{
		Workspace: workspace,
		AccountID: createdAccount.Id,
		OrganizationID: createdOrg.Id,
		TenantID: createdTenant.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createdWorkspace)	

	// Create a test scraping job
	testJob := testutils.GenerateRandomizedScrapingJob()
	testJob.Leads = nil  // Ensure no leads are attached to the job
	createdJob, err := conn.CreateScrapingJob(ctx, createdWorkspace.Id, 	testJob)
	require.NoError(t, err)
	require.NotNil(t, createdJob)

	// Create multiple test leads with explicit ordering
	numLeads := 5
	leadIDs := make([]uint64, 0, numLeads)

	// Create leads with a small delay to ensure ordering
	for i := 0; i < numLeads; i++ {
		lead := testutils.GenerateRandomLead()
		lead.Name = fmt.Sprintf("Test Lead %d", i)

		created, err := conn.CreateLead(ctx, createdJob.Id, lead)
		require.NoError(t, err)
		require.NotNil(t, created)
		leadIDs = append(leadIDs, created.Id)

		// Small delay to ensure consistent ordering
		time.Sleep(time.Millisecond)

		// Check count after each creation
		var currentCount int64
		result := conn.Client.Engine.Table("leads").Count(&currentCount)
		require.NoError(t, result.Error)
		t.Logf("After creating lead %d: %d leads in database", i+1, currentCount)
	}

	// Verify we have exactly the number of leads we expect
	var count int64
	result := conn.Client.Engine.Table("leads").Count(&count)
	require.NoError(t, result.Error)
	require.Equal(t, int64(numLeads), count, "Should have exactly %d leads in the database", numLeads)

	// Also check unscoped count
	var unscopedCount int64
	result = conn.Client.Engine.Unscoped().Table("leads").Count(&unscopedCount)
	require.NoError(t, result.Error)
	require.Equal(t, int64(numLeads), unscopedCount, "Should have exactly %d leads in the database (unscoped)", numLeads)

	// Clean up after all tests
	defer func() {
		// Clean up leads
		for _, id := range leadIDs {
			err := conn.DeleteLead(ctx, id, DeletionTypeHard)
			require.NoError(t, err)
		}
		// Clean up job
		if createdJob != nil {
			err := conn.DeleteScrapingJob(ctx, createdJob.Id)
			require.NoError(t, err)
		}
		// Clean up tenant
		if createdTenant != nil {
			err := conn.DeleteTenant(ctx, &DeleteTenantInput{ID: createdTenant.Id})
			require.NoError(t, err)
		}
		// Clean up organization
		if createdOrg != nil {
			err := conn.DeleteOrganization(ctx, &DeleteOrganizationInput{ID: createdOrg.Id})
			require.NoError(t, err)
		}
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
				// Verify ordering (descending by ID)
				for i := 1; i < len(leads); i++ {
					assert.True(t, leads[i-1].Id > leads[i].Id, "Leads should be ordered by ID in descending order")
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testCtx context.Context
			if tt.name == "context timeout - failure" {
				var cancel context.CancelFunc
				testCtx, cancel = context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()
				time.Sleep(2 * time.Millisecond)
			} else {
				testCtx = ctx
			}

			results, err := conn.ListLeads(testCtx, tt.limit, tt.offset)

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
	// Clean up any existing data
	cleanup := func() {
		result := conn.Client.Engine.Unscoped().Exec("DELETE FROM leads")
		require.NoError(t, result.Error)
		result = conn.Client.Engine.Unscoped().Exec("DELETE FROM gmaps_jobs")
		require.NoError(t, result.Error)
	}
	cleanup()
	defer cleanup()

	results, err := conn.ListLeads(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.NotNil(t, results)
	assert.Empty(t, results)
}
