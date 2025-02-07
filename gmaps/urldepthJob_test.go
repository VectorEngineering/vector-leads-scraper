package gmaps

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gosom/scrapemate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDeduper is a mock implementation of deduper.Deduper
type MockDeduper struct {
	mock.Mock
}

func (m *MockDeduper) AddIfNotExists(ctx context.Context, url string) bool {
	args := m.Called(ctx, url)
	return args.Bool(0)
}

// MockExiter is a mock implementation of exiter.Exiter
type MockExiter struct {
	mock.Mock
}

func (m *MockExiter) IncrPlacesFound(count int) {
	m.Called(count)
}

func (m *MockExiter) IncrSeedCompleted(count int) {
	m.Called(count)
}

func (m *MockExiter) IncrPlacesCompleted(count int) {
	m.Called(count)
}

func (m *MockExiter) Run(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockExiter) SetCancelFunc(cancel context.CancelFunc) {
	m.Called(cancel)
}

func (m *MockExiter) SetSeedCount(count int) {
	m.Called(count)
}

func TestNewGenericCrawlJob(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		url      string
		maxDepth int
		depth    int
	}{
		{
			name:     "with custom ID",
			id:       "test-id",
			url:      "https://example.com",
			maxDepth: 3,
			depth:    0,
		},
		{
			name:     "with empty ID",
			id:       "",
			url:      "https://example.com",
			maxDepth: 2,
			depth:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDeduper := &MockDeduper{}
			mockExiter := &MockExiter{}

			job := NewGenericCrawlJob(
				tt.id,
				tt.url,
				tt.maxDepth,
				tt.depth,
				WithGenericDeduper(mockDeduper),
				WithGenericExitMonitor(mockExiter),
			)

			assert.NotNil(t, job)
			if tt.id != "" {
				assert.Equal(t, tt.id, job.ID)
			} else {
				assert.NotEmpty(t, job.ID) // UUID should be generated
			}
			assert.Equal(t, tt.url, job.URL)
			assert.Equal(t, tt.maxDepth, job.MaxDepth)
			assert.Equal(t, tt.depth, job.Depth)
			assert.Equal(t, http.MethodGet, job.Method)
			assert.Equal(t, scrapemate.PriorityLow, job.Priority)
			assert.Equal(t, 3, job.MaxRetries)
		})
	}
}

func TestGenericCrawlJobProcess(t *testing.T) {
	html := `
		<html>
			<body>
				<a href="https://example.com/page1">Link 1</a>
				<a href="/relative/path">Link 2</a>
				<a href="">Empty Link</a>
				<a>No Href</a>
			</body>
		</html>
	`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	assert.NoError(t, err)

	tests := []struct {
		name           string
		depth          int
		maxDepth       int
		baseURL        string
		expectedJobs   int
		deduperResult  bool
		expectDeduper  bool
		expectExitCall bool
	}{
		{
			name:           "process with depth less than max",
			depth:          0,
			maxDepth:       2,
			baseURL:        "https://example.com",
			expectedJobs:   2, // Two valid links should be processed
			deduperResult:  true,
			expectDeduper:  true,
			expectExitCall: true,
		},
		{
			name:           "process at max depth",
			depth:          2,
			maxDepth:       2,
			baseURL:        "https://example.com",
			expectedJobs:   0, // No new jobs should be created at max depth
			deduperResult:  true,
			expectDeduper:  true, // Changed to true since we still check URLs even at max depth
			expectExitCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDeduper := &MockDeduper{}
			mockExiter := &MockExiter{}

			if tt.expectDeduper {
				// Set up expectations for both URLs that will be found
				mockDeduper.On("AddIfNotExists", mock.Anything, "https://example.com/page1").Return(tt.deduperResult)
				mockDeduper.On("AddIfNotExists", mock.Anything, "https://example.com/relative/path").Return(tt.deduperResult)
			}
			if tt.expectExitCall {
				mockExiter.On("IncrPlacesFound", tt.expectedJobs).Return()
				mockExiter.On("IncrSeedCompleted", 1).Return()
			}

			job := NewGenericCrawlJob(
				"test-id",
				tt.baseURL,
				tt.maxDepth,
				tt.depth,
				WithGenericDeduper(mockDeduper),
				WithGenericExitMonitor(mockExiter),
			)

			resp := &scrapemate.Response{
				URL:      tt.baseURL,
				Document: doc,
			}

			_, jobs, err := job.Process(context.Background(), resp)
			assert.NoError(t, err)
			assert.Len(t, jobs, tt.expectedJobs)

			mockDeduper.AssertExpectations(t)
			mockExiter.AssertExpectations(t)
		})
	}
}

func TestResolveURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		ref      string
		expected string
	}{
		{
			name:     "absolute URL",
			baseURL:  "https://example.com",
			ref:      "https://other.com/page",
			expected: "https://other.com/page",
		},
		{
			name:     "relative URL",
			baseURL:  "https://example.com",
			ref:      "/page",
			expected: "https://example.com/page",
		},
		{
			name:     "relative URL with query",
			baseURL:  "https://example.com/path",
			ref:      "?query=value",
			expected: "https://example.com/path?query=value",
		},
		{
			name:     "invalid base URL",
			baseURL:  "://invalid",
			ref:      "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "invalid reference URL",
			baseURL:  "https://example.com",
			ref:      "://invalid",
			expected: "://invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveURL(tt.baseURL, tt.ref)
			assert.Equal(t, tt.expected, result)
		})
	}
} 