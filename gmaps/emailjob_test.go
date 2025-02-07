package gmaps

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/Vector/vector-leads-scraper/exiter"
	"github.com/gosom/scrapemate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmailJob(t *testing.T) {
	entry := &Entry{
		WebSite: "https://example.com",
	}

	tests := []struct {
		name     string
		parentID string
		entry    *Entry
		wantURL  string
	}{
		{
			name:     "creates job with correct parameters",
			parentID: "parent123",
			entry:    entry,
			wantURL:  "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := NewEmailJob(tt.parentID, tt.entry)
			assert.Equal(t, tt.wantURL, job.URL)
			assert.Equal(t, tt.parentID, job.ParentID)
			assert.Equal(t, tt.entry, job.Entry)
			assert.Equal(t, scrapemate.PriorityHigh, job.Priority)
		})
	}
}

func TestEmailExtractJob_Process(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		expectedEmail []string
		withError     bool
	}{
		{
			name: "extracts mailto links",
			html: `
				<html>
					<body>
						<a href="mailto:test@example.com">Email us</a>
						<a href="mailto:invalid@@email">Invalid</a>
						<a href="mailto:another@example.com">Another email</a>
					</body>
				</html>
			`,
			expectedEmail: []string{"test@example.com", "another@example.com"},
			withError:     false,
		},
		{
			name: "extracts emails from text content",
			html: `
				<html>
					<body>
						Contact us at: contact@example.com
						Or reach support: support@example.com
					</body>
				</html>
			`,
			expectedEmail: []string{"contact@example.com", "support@example.com"},
			withError:     false,
		},
		{
			name:          "handles empty document",
			html:          "",
			expectedEmail: nil,
			withError:     false,
		},
		{
			name:      "handles response error",
			html:      `<html><body></body></html>`,
			withError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the job
			job := &EmailExtractJob{
				Entry: &Entry{},
			}

			// Create response
			resp := &scrapemate.Response{
				Body: []byte(tt.html),
			}

			if tt.withError {
				resp.Error = fmt.Errorf("some error")
			} else {
				doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
				require.NoError(t, err)
				resp.Document = doc
			}

			// Process the job
			result, _, err := job.Process(context.Background(), resp)
			require.NoError(t, err)

			// Check the results
			entry, ok := result.(*Entry)
			require.True(t, ok)
			if tt.expectedEmail != nil {
				assert.ElementsMatch(t, tt.expectedEmail, entry.Emails)
			} else {
				assert.Empty(t, entry.Emails)
			}
		})
	}
}

func Test_docEmailExtractor(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected []string
	}{
		{
			name: "extracts valid mailto links",
			html: `
				<a href="mailto:test@example.com">Email 1</a>
				<a href="mailto:another@example.com">Email 2</a>
			`,
			expected: []string{"test@example.com", "another@example.com"},
		},
		{
			name: "skips invalid mailto links",
			html: `
				<a href="mailto:invalid@@email">Invalid</a>
				<a href="mailto:test@example.com">Valid</a>
			`,
			expected: []string{"test@example.com"},
		},
		{
			name: "handles duplicate emails",
			html: `
				<a href="mailto:test@example.com">Email 1</a>
				<a href="mailto:test@example.com">Email 1 Again</a>
			`,
			expected: []string{"test@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			emails := docEmailExtractor(doc)
			assert.ElementsMatch(t, tt.expected, emails)
		})
	}
}

func Test_regexEmailExtractor(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "extracts valid emails from text",
			content: `
				Contact us at test@example.com
				Support available at support@example.com
			`,
			expected: []string{"test@example.com", "support@example.com"},
		},
		{
			name: "handles duplicate emails",
			content: `
				Email: test@example.com
				Same email: test@example.com
			`,
			expected: []string{"test@example.com"},
		},
		{
			name:     "handles empty content",
			content:  "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emails := regexEmailExtractor([]byte(tt.content))
			assert.ElementsMatch(t, tt.expected, emails)
		})
	}
}

func Test_getValidEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		expected    string
		expectError bool
	}{
		{
			name:        "valid email",
			email:       "test@example.com",
			expected:    "test@example.com",
			expectError: false,
		},
		{
			name:        "invalid email",
			email:       "invalid@@email",
			expected:    "",
			expectError: true,
		},
		{
			name:        "email with spaces",
			email:       " test@example.com ",
			expected:    "test@example.com",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getValidEmail(tt.email)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestEmailExtractJob_ProcessOnFetchError(t *testing.T) {
	job := &EmailExtractJob{}
	assert.True(t, job.ProcessOnFetchError())
}

func TestWithEmailJobExitMonitor(t *testing.T) {
	mockExitMonitor := &mockExiter{}
	job := NewEmailJob("parent123", &Entry{}, WithEmailJobExitMonitor(mockExitMonitor))
	assert.Equal(t, mockExitMonitor, job.ExitMonitor)
}

func TestEmailExtractJob_ProcessWithExitMonitor(t *testing.T) {
	mockExitMonitor := &mockExiter{}

	job := &EmailExtractJob{
		Entry:       &Entry{},
		ExitMonitor: mockExitMonitor,
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html></html>"))
	require.NoError(t, err)

	resp := &scrapemate.Response{
		Document: doc,
	}

	_, _, err = job.Process(context.Background(), resp)
	require.NoError(t, err)

	assert.Equal(t, 1, mockExitMonitor.placesCompletedCount)
}

type mockExiter struct {
	exiter.Exiter
	placesCompletedCount int
}

func (m *mockExiter) IncrPlacesCompleted(count int) {
	m.placesCompletedCount += count
}

// Add missing required methods
func (m *mockExiter) IncrPlacesFound(_ int)   {}
func (m *mockExiter) IncrSeedCompleted(_ int) {}
func (m *mockExiter) ShouldExit() bool        { return false }
