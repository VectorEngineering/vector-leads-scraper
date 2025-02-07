package gmaps

import (
	"context"

	html2md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/Vector/vector-leads-scraper/exiter"
	"github.com/google/uuid"
	"github.com/gosom/scrapemate"
)

// MarkdownExtractJobOptions defines options for configuring a MarkdownExtractJob.
type MarkdownExtractJobOptions func(*MarkdownExtractJob)

// MarkdownExtractJob represents a job that extracts a page’s content as Markdown.
type MarkdownExtractJob struct {
	scrapemate.Job

	// Entry should have a field (e.g., Markdown string) to hold the converted Markdown content.
	Entry       *Entry
	ExitMonitor exiter.Exiter
}

// NewMarkdownJob creates a new job for extracting Markdown from a webpage.
func NewMarkdownJob(parentID string, entry *Entry, opts ...MarkdownExtractJobOptions) *MarkdownExtractJob {
	const (
		defaultPrio       = scrapemate.PriorityHigh
		defaultMaxRetries = 0
	)

	job := MarkdownExtractJob{
		Job: scrapemate.Job{
			ID:         uuid.New().String(),
			ParentID:   parentID,
			Method:     "GET",
			URL:        entry.WebSite,
			MaxRetries: defaultMaxRetries,
			Priority:   defaultPrio,
		},
		Entry: entry,
	}

	for _, opt := range opts {
		opt(&job)
	}

	return &job
}

// WithMarkdownJobExitMonitor sets an exit monitor on the job.
func WithMarkdownJobExitMonitor(exitMonitor exiter.Exiter) MarkdownExtractJobOptions {
	return func(j *MarkdownExtractJob) {
		j.ExitMonitor = exitMonitor
	}
}

// Process retrieves the HTML content and converts it to Markdown.
func (j *MarkdownExtractJob) Process(ctx context.Context, resp *scrapemate.Response) (any, []scrapemate.IJob, error) {
	// Clean up the document and body after processing.
	defer func() {
		resp.Document = nil
		resp.Body = nil
	}()

	// Update the exit monitor, if provided.
	defer func() {
		if j.ExitMonitor != nil {
			j.ExitMonitor.IncrPlacesCompleted(1)
		}
	}()

	log := scrapemate.GetLoggerFromContext(ctx)
	log.Info("Processing markdown extraction job", "url", j.URL)

	// If there was an error during the fetch, return immediately.
	if resp.Error != nil {
		return j.Entry, nil, nil
	}

	var htmlContent string

	// Prefer using the parsed document if available.
	if doc, ok := resp.Document.(*goquery.Document); ok {
		content, err := doc.Html()
		if err != nil {
			return j.Entry, nil, err
		}
		htmlContent = content
	} else {
		// Fallback to converting the raw response body.
		htmlContent = string(resp.Body)
	}

	// Convert the HTML content to Markdown.
	converter := html2md.NewConverter("", true, nil)
	markdownContent, err := converter.ConvertString(htmlContent)
	if err != nil {
		return j.Entry, nil, err
	}

	j.Entry.Markdown = markdownContent

	return j.Entry, nil, nil
}

// ProcessOnFetchError indicates that this job should be retried if a fetch error occurs.
func (j *MarkdownExtractJob) ProcessOnFetchError() bool {
	return true
}
