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

func TestNewMarkdownJob(t *testing.T) {
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
			job := NewMarkdownJob(tt.parentID, tt.entry)
			assert.Equal(t, tt.wantURL, job.URL)
			assert.Equal(t, tt.parentID, job.ParentID)
			assert.Equal(t, tt.entry, job.Entry)
			assert.Equal(t, scrapemate.PriorityHigh, job.Priority)
		})
	}
}

func TestMarkdownExtractJob_Process(t *testing.T) {
	tests := []struct {
		name            string
		html            string
		expectedMarkdown string
		withError       bool
	}{
		{
			name: "converts simple HTML to markdown",
			html: `
				<html>
					<body>
						<h1>Title</h1>
						<p>This is a paragraph.</p>
						<ul>
							<li>Item 1</li>
							<li>Item 2</li>
						</ul>
					</body>
				</html>
			`,
			expectedMarkdown: "# Title\n\nThis is a paragraph.\n\n- Item 1\n- Item 2",
			withError:        false,
		},
		{
			name: "handles links and formatting",
			html: `
				<html>
					<body>
						<h2>Links and Formatting</h2>
						<p><strong>Bold text</strong> and <em>italic text</em></p>
						<a href="https://example.com">Example Link</a>
					</body>
				</html>
			`,
			expectedMarkdown: "## Links and Formatting\n\n**Bold text** and _italic text_\n\n [Example Link](https://example.com)",
			withError:        false,
		},
		{
			name:            "handles empty document",
			html:            "",
			expectedMarkdown: "",
			withError:        false,
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
			job := &MarkdownExtractJob{
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
			if tt.withError {
				assert.Empty(t, entry.Markdown)
			} else {
				// Normalize line endings and whitespace for comparison
				normalizedExpected := strings.TrimSpace(strings.ReplaceAll(tt.expectedMarkdown, "\r\n", "\n"))
				normalizedActual := strings.TrimSpace(strings.ReplaceAll(entry.Markdown, "\r\n", "\n"))
				assert.Equal(t, normalizedExpected, normalizedActual)
			}
		})
	}
}

func TestMarkdownExtractJob_ProcessOnFetchError(t *testing.T) {
	job := &MarkdownExtractJob{}
	assert.True(t, job.ProcessOnFetchError())
}

func TestWithMarkdownJobExitMonitor(t *testing.T) {
	mockExitMonitor := &mockExiter{}
	job := NewMarkdownJob("parent123", &Entry{}, WithMarkdownJobExitMonitor(mockExitMonitor))
	assert.Equal(t, mockExitMonitor, job.ExitMonitor)
}

func TestMarkdownExtractJob_ProcessWithExitMonitor(t *testing.T) {
	mockExitMonitor := &mockExiter{}
	
	job := &MarkdownExtractJob{
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