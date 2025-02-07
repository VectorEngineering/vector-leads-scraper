package gmaps

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gosom/scrapemate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewURLJob(t *testing.T) {
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
			job := NewURLJob(tt.parentID, tt.entry)
			assert.Equal(t, tt.wantURL, job.URL)
			assert.Equal(t, tt.parentID, job.ParentID)
			assert.Equal(t, tt.entry, job.Entry)
			assert.Equal(t, scrapemate.PriorityHigh, job.Priority)
		})
	}
}

func TestURLExtractJob_Process(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		expectedURLs []string
		withError   bool
	}{
		{
			name: "extracts href links",
			html: `
				<html>
					<body>
						<a href="https://example.com">Link 1</a>
						<a href="https://test.com">Link 2</a>
						<a href="">Empty link</a>
					</body>
				</html>
			`,
			expectedURLs: []string{"https://example.com", "https://test.com"},
			withError:    false,
		},
		{
			name: "extracts src attributes",
			html: `
				<html>
					<body>
						<img src="https://example.com/image.jpg">
						<script src="https://test.com/script.js"></script>
					</body>
				</html>
			`,
			expectedURLs: []string{"https://example.com/image.jpg", "https://test.com/script.js"},
			withError:    false,
		},
		{
			name: "handles mixed href and src",
			html: `
				<html>
					<body>
						<a href="https://example.com">Link</a>
						<img src="https://example.com/image.jpg">
						<a href="https://example.com">Duplicate Link</a>
					</body>
				</html>
			`,
			expectedURLs: []string{"https://example.com", "https://example.com/image.jpg"},
			withError:    false,
		},
		{
			name:        "handles empty document",
			html:        "",
			expectedURLs: nil,
			withError:    false,
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
			job := &URLExtractJob{
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
			if tt.expectedURLs != nil {
				assert.ElementsMatch(t, tt.expectedURLs, entry.Urls)
			} else {
				assert.Empty(t, entry.Urls)
			}
		})
	}
}

func Test_docURLExtractor(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected []string
	}{
		{
			name: "extracts href links",
			html: `
				<a href="https://example.com">Link 1</a>
				<a href="https://test.com">Link 2</a>
			`,
			expected: []string{"https://example.com", "https://test.com"},
		},
		{
			name: "extracts src attributes",
			html: `
				<img src="https://example.com/image.jpg">
				<script src="https://test.com/script.js"></script>
			`,
			expected: []string{"https://example.com/image.jpg", "https://test.com/script.js"},
		},
		{
			name: "handles duplicate URLs",
			html: `
				<a href="https://example.com">Link 1</a>
				<a href="https://example.com">Link 1 Again</a>
			`,
			expected: []string{"https://example.com"},
		},
		{
			name: "skips empty URLs",
			html: `
				<a href="">Empty</a>
				<img src="">
				<a href="https://example.com">Valid</a>
			`,
			expected: []string{"https://example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			urls := docURLExtractor(doc)
			assert.ElementsMatch(t, tt.expected, urls)
		})
	}
}

func TestURLExtractJob_ProcessOnFetchError(t *testing.T) {
	job := &URLExtractJob{}
	assert.True(t, job.ProcessOnFetchError())
}

func TestWithURLJobExitMonitor(t *testing.T) {
	mockExitMonitor := &mockExiter{}
	job := NewURLJob("parent123", &Entry{}, WithURLJobExitMonitor(mockExitMonitor))
	assert.Equal(t, mockExitMonitor, job.ExitMonitor)
}

func TestURLExtractJob_ProcessWithExitMonitor(t *testing.T) {
	mockExitMonitor := &mockExiter{}
	
	job := &URLExtractJob{
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