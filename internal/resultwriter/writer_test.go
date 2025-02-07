package resultwriter

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	postgresdb "github.com/SolomonAIEngineering/backend-core-library/database/postgres"
	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/gosom/scrapemate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/Vector/vector-leads-scraper/gmaps"
	"github.com/Vector/vector-leads-scraper/internal/database"
	"github.com/Vector/vector-leads-scraper/internal/testutils"
	"github.com/Vector/vector-leads-scraper/pkg/webhook"
	"github.com/Vector/vector-leads-scraper/testcontainers"
)

// mockWebhookClient implements webhook.Client interface for testing
type mockWebhookClient struct {
	records []webhook.Record
}

func (m *mockWebhookClient) Send(ctx context.Context, record webhook.Record) error {
	m.records = append(m.records, record)
	return nil
}

func (m *mockWebhookClient) Shutdown(ctx context.Context) error {
	return nil
}

func (m *mockWebhookClient) Flush(ctx context.Context) error {
	return nil
}

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

func TestConvertDayOfWeek(t *testing.T) {
	tests := []struct {
		name string
		day  string
		want lead_scraper_servicev1.BusinessHours_DayOfWeek
	}{
		{
			name: "Monday",
			day:  "Monday",
			want: lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_MONDAY,
		},
		{
			name: "Tuesday",
			day:  "Tuesday",
			want: lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_TUESDAY,
		},
		{
			name: "Wednesday",
			day:  "Wednesday",
			want: lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_WEDNESDAY,
		},
		{
			name: "Thursday",
			day:  "Thursday",
			want: lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_THURSDAY,
		},
		{
			name: "Friday",
			day:  "Friday",
			want: lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_FRIDAY,
		},
		{
			name: "Saturday",
			day:  "Saturday",
			want: lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_SATURDAY,
		},
		{
			name: "Sunday",
			day:  "Sunday",
			want: lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_SUNDAY,
		},
		{
			name: "Invalid day",
			day:  "InvalidDay",
			want: lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_UNSPECIFIED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertDayOfWeek(tt.day)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetMainPhotoURL(t *testing.T) {
	tests := []struct {
		name   string
		photos []string
		want   string
	}{
		{
			name:   "empty photos",
			photos: []string{},
			want:   "",
		},
		{
			name:   "single photo",
			photos: []string{"https://example.com/photo1.jpg"},
			want:   "https://example.com/photo1.jpg",
		},
		{
			name:   "multiple photos",
			photos: []string{"https://example.com/photo1.jpg", "https://example.com/photo2.jpg"},
			want:   "https://example.com/photo1.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getMainPhotoURL(tt.photos)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResultWriter_Close(t *testing.T) {
	setup := setupTest(t)
	defer setup.cleanup()

	tests := []struct {
		name      string
		cfg       *Config
		wantError bool
	}{
		{
			name: "success - with webhook",
			cfg: &Config{
				WebhookEnabled:         true,
				WebhookEndpoints:       []string{"http://example.com"},
				WebhookBatchSize:       100,
				BatchSize:              50,
				WebhookFlushInterval:   time.Second,
				FlushInterval:          time.Second,
				WebhookRetryInterval:   time.Second,
			},
			wantError: false,
		},
		{
			name: "success - without webhook",
			cfg:  DefaultConfig(),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, err := New(setup.db, setup.logger, tt.cfg)
			require.NoError(t, err)

			err = writer.Close()
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConvertEntryToLead_WithBusinessHours(t *testing.T) {
	entry := &gmaps.Entry{
		ID:           "test-id",
		Title:        "Test Business",
		Category:     "Test Category",
		OpenHours: map[string][]string{
			"Monday":    {"09:00", "17:00"},
			"Tuesday":   {"09:00", "17:00"},
			"Wednesday": {"09:00", "17:00"},
			"Thursday":  {"09:00", "17:00"},
			"Friday":    {"09:00", "17:00"},
			"Saturday":  {"10:00", "15:00"},
			"Sunday":    {},  // Closed
		},
		CompleteAddress: gmaps.Address{
			City:    "Test City",
			State:   "Test State",
			Country: "Test Country",
		},
	}

	lead := convertEntryToLead(entry)
	require.NotNil(t, lead)
	require.NotNil(t, lead.RegularHours)

	// Create a map to track which days we've seen
	seenDays := make(map[string]bool)
	for _, hours := range lead.RegularHours {
		dayStr := ""
		switch hours.Day {
		case lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_MONDAY:
			dayStr = "Monday"
		case lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_TUESDAY:
			dayStr = "Tuesday"
		case lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_WEDNESDAY:
			dayStr = "Wednesday"
		case lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_THURSDAY:
			dayStr = "Thursday"
		case lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_FRIDAY:
			dayStr = "Friday"
		case lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_SATURDAY:
			dayStr = "Saturday"
		case lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_SUNDAY:
			dayStr = "Sunday"
		}
		seenDays[dayStr] = true

		switch hours.Day {
		case lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_SUNDAY:
			assert.True(t, hours.Closed)
			assert.Empty(t, hours.OpenTime)
			assert.Empty(t, hours.CloseTime)
		case lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_SATURDAY:
			assert.Equal(t, "10:00", hours.OpenTime)
			assert.Equal(t, "15:00", hours.CloseTime)
			assert.False(t, hours.Closed)
		default:
			assert.Equal(t, "09:00", hours.OpenTime)
			assert.Equal(t, "17:00", hours.CloseTime)
			assert.False(t, hours.Closed)
		}
	}

	// Verify we've seen all days
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	for _, day := range days {
		assert.True(t, seenDays[day], "Missing day: %v", day)
	}
}

func TestResultWriter_Run_WithWebhook(t *testing.T) {
	setup := setupTest(t)
	defer setup.cleanup()

	// Create mock webhook client
	mockClient := &mockWebhookClient{}

	// Create writer with webhook enabled
	cfg := &Config{
		WebhookEnabled:         true,
		WebhookEndpoints:       []string{"http://example.com"},
		WebhookBatchSize:       100,
		BatchSize:              50,
		WebhookFlushInterval:   time.Second,
		FlushInterval:          time.Second,
		WebhookRetryInterval:   time.Second,
	}

	writer, err := New(setup.db, setup.logger, cfg)
	require.NoError(t, err)
	writer.webhookClient = mockClient

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

	// Send test entries
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

	// Close writer
	err = writer.Close()
	require.NoError(t, err)

	// Verify webhook records
	require.NotEmpty(t, mockClient.records)
	for _, record := range mockClient.records {
		require.NotNil(t, record.Data)
	}
}

// cleanupJobs deletes all jobs from the database
func cleanupJobs(ctx context.Context, db *database.Db) error {
	jobs, err := db.ListScrapingJobs(ctx, 1000, 0)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	for _, job := range jobs {
		if err := db.DeleteScrapingJob(ctx, job.Id); err != nil {
			return fmt.Errorf("failed to delete job %d: %w", job.Id, err)
		}
	}
	return nil
}

func TestProvider(t *testing.T) {
	setup := setupTest(t)
	defer setup.cleanup()

	// Clean up any existing jobs before each test
	err := cleanupJobs(context.Background(), setup.db)
	require.NoError(t, err)

	t.Run("NewProvider", func(t *testing.T) {
		// Test with default options
		p := NewProvider(setup.db, setup.logger)
		require.NotNil(t, p)
		assert.Equal(t, defaultBatchSize, p.batchSize)
		assert.NotNil(t, p.jobc)
		assert.NotNil(t, p.errc)
		assert.NotNil(t, p.mu)
		assert.False(t, p.started)

		// Test with custom batch size
		customBatchSize := 100
		p = NewProvider(setup.db, setup.logger, WithBatchSize(customBatchSize))
		require.NotNil(t, p)
		assert.Equal(t, customBatchSize, p.batchSize)
	})

	t.Run("WithBatchSize", func(t *testing.T) {
		// Test with valid batch size
		p := NewProvider(setup.db, setup.logger, WithBatchSize(100))
		assert.Equal(t, 100, p.batchSize)

		// Test with invalid batch size (should use default)
		p = NewProvider(setup.db, setup.logger, WithBatchSize(-1))
		assert.Equal(t, defaultBatchSize, p.batchSize)
	})

	t.Run("Jobs", func(t *testing.T) {
		p := NewProvider(setup.db, setup.logger)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Get job channels
		jobChan, errChan := p.Jobs(ctx)
		require.NotNil(t, jobChan)
		require.NotNil(t, errChan)

		// Create a valid GmapJob
		scrapingJob := testutils.GenerateRandomizedScrapingJob()
		// Ensure zoom is within valid range (1-20)
		if scrapingJob.Zoom < 1 || scrapingJob.Zoom > 20 {
			scrapingJob.Zoom = 15
		}

		job := gmaps.NewGmapJob(
			fmt.Sprintf("%d", scrapingJob.Id),
			scrapingJob.Lang.String(),
			string(scrapingJob.Keywords[0]),
			1,
			true,
			fmt.Sprintf("%s,%s", scrapingJob.Lat, scrapingJob.Lon),
			int(scrapingJob.Zoom),
			gmaps.WithWorkspaceID(setup.workspace.Id),
		)

		// Push the job
		err := p.Push(ctx, job)
		require.NoError(t, err)

		// Wait for job or error
		select {
		case receivedJob := <-jobChan:
			require.NotNil(t, receivedJob)
			gmapJob, ok := receivedJob.(*gmaps.GmapJob)
			require.True(t, ok)
			assert.Equal(t, job.ID, gmapJob.GetID())
			assert.Equal(t, job.Priority, gmapJob.GetPriority())
		case err := <-errChan:
			t.Fatalf("Unexpected error: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for job")
		}

		// Test context cancellation
		cancel()
		_, ok := <-jobChan
		assert.False(t, ok, "jobChan should be closed after context cancellation")
	})

	t.Run("Jobs_WithNonUnspecifiedStatus", func(t *testing.T) {
		// Clean up any existing jobs before the test
		err := cleanupJobs(context.Background(), setup.db)
		require.NoError(t, err)

		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Create a job with non-UNSPECIFIED status
		scrapingJob := testutils.GenerateRandomizedScrapingJob()
		scrapingJob.Status = lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED
		scrapingJob.Url = "https://example.com"  // Set a valid URL
		job, err := setup.db.CreateScrapingJob(ctx, setup.workspace.Id, scrapingJob)
		require.NoError(t, err)
		require.NotNil(t, job)

		// Verify that there are no UNSPECIFIED jobs
		jobs, err := setup.db.ListScrapingJobsByStatus(ctx, lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_UNSPECIFIED, 50, 0)
		require.NoError(t, err)
		require.Empty(t, jobs, "There should be no UNSPECIFIED jobs")

		// Create provider and get jobs channel
		provider := NewProvider(setup.db, setup.logger)
		jobsChan, errChan := provider.Jobs(ctx)

		// Try to receive a job with timeout
		select {
		case job, ok := <-jobsChan:
			if ok {
				t.Errorf("Received unexpected job: %v", job)
			}
		case err := <-errChan:
			t.Errorf("Unexpected error: %v", err)
		case <-time.After(1 * time.Second):
			// Success - no job received
		}
	})

	t.Run("Jobs_WithBackoff", func(t *testing.T) {
		// Clean up any existing jobs before the test
		err := cleanupJobs(context.Background(), setup.db)
		require.NoError(t, err)

		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Create provider and get jobs channel
		provider := NewProvider(setup.db, setup.logger)
		jobsChan, errChan := provider.Jobs(ctx)

		// Try to receive a job with timeout
		select {
		case job, ok := <-jobsChan:
			if ok {
				t.Errorf("Received unexpected job: %v", job)
			}
		case err := <-errChan:
			if err != nil && !strings.Contains(err.Error(), "context canceled") {
				t.Errorf("Unexpected error: %v", err)
			}
		case <-time.After(1 * time.Second):
			// Success - no job received
		}
	})

	t.Run("Push", func(t *testing.T) {
		p := NewProvider(setup.db, setup.logger)
		ctx := context.Background()

		// Test pushing valid job
		scrapingJob := testutils.GenerateRandomizedScrapingJob()
		// Ensure zoom is within valid range (1-20)
		if scrapingJob.Zoom < 1 || scrapingJob.Zoom > 20 {
			scrapingJob.Zoom = 15
		}

		job := gmaps.NewGmapJob(
			fmt.Sprintf("%d", scrapingJob.Id),
			scrapingJob.Lang.String(),
			string(scrapingJob.Keywords[0]),
			1,
			true,
			fmt.Sprintf("%s,%s", scrapingJob.Lat, scrapingJob.Lon),
			int(scrapingJob.Zoom),
			gmaps.WithWorkspaceID(setup.workspace.Id),
		)
		err := p.Push(ctx, job)
		require.NoError(t, err)

		// Test pushing invalid job type
		invalidJob := &scrapemate.Job{
			ID:       "invalid-job",
			Priority: 1,
		}
		err = p.Push(ctx, invalidJob)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid job type")

		// Test pushing job with invalid zoom
		scrapingJob = testutils.GenerateRandomizedScrapingJob()
		invalidZoomJob := gmaps.NewGmapJob(
			fmt.Sprintf("%d", scrapingJob.Id),
			scrapingJob.Lang.String(),
			string(scrapingJob.Keywords[0]),
			1,
			true,
			fmt.Sprintf("%s,%s", scrapingJob.Lat, scrapingJob.Lon),
			0, // Invalid zoom level
			gmaps.WithWorkspaceID(setup.workspace.Id),
		)
		err = p.Push(ctx, invalidZoomJob)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ScrapingJob.Zoom")
	})
} 