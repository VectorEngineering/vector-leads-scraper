package gmaps

import (
	"testing"

	"github.com/Vector/vector-leads-scraper/exiter"
	"github.com/gosom/scrapemate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for playwright.Response
type MockResponse struct {
	mock.Mock
}

func (m *MockResponse) URL() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockResponse) Status() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockResponse) Headers() map[string]string {
	args := m.Called()
	return args.Get(0).(map[string]string)
}

// Mock for playwright.ElementHandle
type MockElementHandle struct {
	mock.Mock
}

func TestNewPlaceJob(t *testing.T) {
	tests := []struct {
		name         string
		parentID     string
		langCode     string
		url          string
		extractEmail bool
		opts         []PlaceJobOptions
		want         *PlaceJob
	}{
		{
			name:         "creates job with default options",
			parentID:     "test-parent",
			langCode:     "en",
			url:          "https://maps.google.com/place/123",
			extractEmail: true,
			opts:         nil,
			want: &PlaceJob{
				Job: scrapemate.Job{
					ParentID:   "test-parent",
					Method:     "GET",
					URL:        "https://maps.google.com/place/123",
					URLParams:  map[string]string{"hl": "en"},
					MaxRetries: 3,
					Priority:   scrapemate.PriorityMedium,
				},
				UsageInResultststs: true,
				ExtractEmail:       true,
			},
		},
		{
			name:         "creates job with exit monitor",
			parentID:     "test-parent",
			langCode:     "en",
			url:          "https://maps.google.com/place/123",
			extractEmail: false,
			opts: []PlaceJobOptions{
				WithPlaceJobExitMonitor(exiter.New()),
			},
			want: &PlaceJob{
				Job: scrapemate.Job{
					ParentID:   "test-parent",
					Method:     "GET",
					URL:        "https://maps.google.com/place/123",
					URLParams:  map[string]string{"hl": "en"},
					MaxRetries: 3,
					Priority:   scrapemate.PriorityMedium,
				},
				UsageInResultststs: true,
				ExtractEmail:       false,
				ExitMonitor:        exiter.New(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPlaceJob(tt.parentID, tt.langCode, tt.url, tt.extractEmail, tt.opts...)
			
			// Check all fields except ID which is randomly generated
			assert.Equal(t, tt.want.ParentID, got.ParentID)
			assert.Equal(t, tt.want.Method, got.Method)
			assert.Equal(t, tt.want.URL, got.URL)
			assert.Equal(t, tt.want.URLParams, got.URLParams)
			assert.Equal(t, tt.want.MaxRetries, got.MaxRetries)
			assert.Equal(t, tt.want.Priority, got.Priority)
			assert.Equal(t, tt.want.UsageInResultststs, got.UsageInResultststs)
			assert.Equal(t, tt.want.ExtractEmail, got.ExtractEmail)
		})
	}
}

func TestPlaceJob_UseInResults(t *testing.T) {
	tests := []struct {
		name string
		job  *PlaceJob
		want bool
	}{
		{
			name: "returns true when UsageInResultststs is true",
			job: &PlaceJob{
				UsageInResultststs: true,
			},
			want: true,
		},
		{
			name: "returns false when UsageInResultststs is false",
			job: &PlaceJob{
				UsageInResultststs: false,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.job.UseInResults()
			assert.Equal(t, tt.want, got)
		})
	}
}

// Skip browser-dependent tests in unit tests
func TestPlaceJob_BrowserActions(t *testing.T) {
	t.Skip("Skipping browser-dependent tests")
} 