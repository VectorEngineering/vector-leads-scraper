package resultwriter

import (
	"context"
	"sort"
	"testing"

	postgresdb "github.com/SolomonAIEngineering/backend-core-library/database/postgres"
	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/gosom/scrapemate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/Vector/vector-leads-scraper/gmaps"
	"github.com/Vector/vector-leads-scraper/internal/database"
	"github.com/Vector/vector-leads-scraper/internal/testutils"
	"github.com/Vector/vector-leads-scraper/testcontainers"
)

// testSetup encapsulates common test setup logic and returns cleanup function
type testSetup struct {
	db        *database.Db
	logger    *zap.Logger
	workspace *lead_scraper_servicev1.Workspace
	job       *lead_scraper_servicev1.ScrapingJob
	cleanup   func()
}

func setupTest(t *testing.T) *testSetup {
	t.Helper()
	logger := zap.NewNop()

	// Setup test containers
	tc := testcontainers.NewTestContext(t)

	// Create test database
	client, err := postgresdb.NewInMemoryTestDbClient(lead_scraper_servicev1.GetDatabaseSchemas()...)
	require.NoError(t, err)
	db, err := database.New(client, logger)
	require.NoError(t, err)

	// Create test organization
	org := testutils.GenerateRandomizedOrganization()
	createdOrg, err := db.CreateOrganization(context.Background(), &database.CreateOrganizationInput{
		Organization: org,
	})
	require.NoError(t, err)
	require.NotNil(t, createdOrg)

	// Create test tenant
	tenant := testutils.GenerateRandomizedTenant()
	createdTenant, err := db.CreateTenant(context.Background(), &database.CreateTenantInput{
		Tenant:         tenant,
		OrganizationID: createdOrg.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createdTenant)

	// Create test account
	account := testutils.GenerateRandomizedAccount()
	createdAccount, err := db.CreateAccount(context.Background(), &database.CreateAccountInput{
		Account:  account,
		TenantID: createdTenant.Id,
		OrgID:    createdOrg.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createdAccount)

	// Create test workspace
	workspace := testutils.GenerateRandomWorkspace()
	createdWorkspace, err := db.CreateWorkspace(context.Background(), &database.CreateWorkspaceInput{
		Workspace:      workspace,
		AccountID:      createdAccount.Id,
		OrganizationID: createdOrg.Id,
		TenantID:      createdTenant.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createdWorkspace)

	// Create test scraping job
	job := testutils.GenerateRandomizedScrapingJob()
	job.Status = lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_IN_PROGRESS
	job.PayloadType = "scraping_job"
	job.Url = "https://maps.google.com/test-search"
	createdJob, err := db.CreateScrapingJob(context.Background(), createdWorkspace.Id, job)
	require.NoError(t, err)
	require.NotNil(t, createdJob)

	return &testSetup{
		db:        db,
		logger:    logger,
		workspace: createdWorkspace,
		job:       createdJob,
		cleanup:   tc.Cleanup,
	}
}

func TestNew(t *testing.T) {
	setup := setupTest(t)
	defer setup.cleanup()

	tests := []struct {
		name      string
		db        *database.Db
		cfg       *Config
		wantError bool
	}{
		{
			name:      "success with default config",
			db:        setup.db,
			cfg:       nil,
			wantError: false,
		},
		{
			name:      "success with custom config",
			db:        setup.db,
			cfg:       DefaultConfig(),
			wantError: false,
		},
		{
			name:      "failure - nil database",
			db:        nil,
			cfg:       DefaultConfig(),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, err := New(tt.db, setup.logger, tt.cfg)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, writer)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, writer)
			assert.NotNil(t, writer.cfg)
			assert.Equal(t, tt.db, writer.db)
		})
	}
}

func TestConvertEntryToLead(t *testing.T) {
	entry := &gmaps.Entry{
		ID:           "test-id",
		Link:         "https://maps.google.com/test",
		Title:        "Test Business",
		Category:     "Test Category",
		Address:      "123 Test St",
		WebSite:      "https://test.com",
		Phone:        "123-456-7890",
		ReviewCount:  100,
		ReviewRating: 4.5,
		Latitude:     40.7128,
		Longtitude:   -74.0060,
		Cid:          "test-cid",
		CompleteAddress: gmaps.Address{
			City:    "Test City",
			State:   "Test State",
			Country: "Test Country",
		},
	}

	lead := convertEntryToLead(entry)
	require.NotNil(t, lead)

	// Test all field mappings
	assert.Equal(t, entry.Title, lead.Name)
	assert.Equal(t, entry.WebSite, lead.Website)
	assert.Equal(t, entry.Phone, lead.Phone)
	assert.Equal(t, entry.Address, lead.Address)
	assert.Equal(t, entry.CompleteAddress.City, lead.City)
	assert.Equal(t, entry.CompleteAddress.State, lead.State)
	assert.Equal(t, entry.CompleteAddress.Country, lead.Country)
	assert.Equal(t, entry.Category, lead.Industry)
	assert.Equal(t, entry.Cid, lead.PlaceId)
	assert.Equal(t, entry.Link, lead.GoogleMapsUrl)
	assert.Equal(t, entry.Latitude, lead.Latitude)
	assert.Equal(t, entry.Longtitude, lead.Longitude)
	assert.Equal(t, float32(entry.ReviewRating), lead.GoogleRating)
	assert.Equal(t, int32(entry.ReviewCount), lead.ReviewCount)
}

// generateTestEntries creates a slice of test gmaps.Entry for testing
func generateTestEntries() []*gmaps.Entry {
	return []*gmaps.Entry{
		{
			ID:           "test-id-1",
			Title:        "Test Business 1",
			Category:     "Category 1",
			Address:      "123 Test St",
			WebSite:      "https://test1.com",
			Phone:        "123-456-7890",
			ReviewCount:  100,
			ReviewRating: 4.5,
			Latitude:     40.7128,
			Longtitude:   -74.0060,
			CompleteAddress: gmaps.Address{
				City:    "Test City",
				State:   "Test State",
				Country: "Test Country",
			},
		},
		{
			ID:           "test-id-2",
			Title:        "Test Business 2",
			Category:     "Category 2",
			Address:      "456 Test St",
			WebSite:      "https://test2.com",
			Phone:        "098-765-4321",
			ReviewCount:  200,
			ReviewRating: 4.8,
			Latitude:     40.7129,
			Longtitude:   -74.0061,
			CompleteAddress: gmaps.Address{
				City:    "Test City",
				State:   "Test State",
				Country: "Test Country",
			},
		},
	}
}

func TestResultWriter_Run(t *testing.T) {
	setup := setupTest(t)
	defer setup.cleanup()

	// Create result writer with default configuration
	writer, err := New(setup.db, setup.logger, DefaultConfig())
	require.NoError(t, err)

	// Get test entries
	entries := generateTestEntries()

	// Create test job
	testJob := &gmaps.GmapJob{
		ScrapingJobID: setup.job.Id,
		Job: scrapemate.Job{
			ID:       setup.job.Name,
			Priority: int(setup.job.Priority),
		},
	}

	// Create input channel and context
	in := make(chan scrapemate.Result)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run writer in goroutine
	errCh := make(chan error)
	go func() {
		errCh <- writer.Run(ctx, in)
	}()

	// Send test entries with job information
	for _, entry := range entries {
		in <- scrapemate.Result{
			Data: entry,
			Job:  testJob,
		}
	}

	// Close input channel and wait for writer to finish
	close(in)
	err = <-errCh
	require.NoError(t, err)

	// Verify leads were written to database
	leads, err := setup.db.ListLeads(context.Background(), len(entries), 0)
	require.NoError(t, err)
	require.Len(t, leads, len(entries))

	// Sort both entries and leads by name for consistent comparison
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Title < entries[j].Title
	})
	sort.Slice(leads, func(i, j int) bool {
		return leads[i].Name < leads[j].Name
	})

	// Verify each lead's data matches the corresponding entry
	for i, lead := range leads {
		assert.Equal(t, entries[i].Title, lead.Name)
		assert.Equal(t, entries[i].WebSite, lead.Website)
		assert.Equal(t, entries[i].Phone, lead.Phone)
		assert.Equal(t, entries[i].Address, lead.Address)
		assert.Equal(t, entries[i].Category, lead.Industry)
		assert.Equal(t, float32(entries[i].ReviewRating), lead.GoogleRating)
		assert.Equal(t, int32(entries[i].ReviewCount), lead.ReviewCount)
	}
} 