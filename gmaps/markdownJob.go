// Package gmaps provides functionality for scraping and processing Google Maps data.
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
// It follows the functional options pattern for flexible job configuration.
type MarkdownExtractJobOptions func(*MarkdownExtractJob)

// MarkdownExtractJob represents a job that extracts a page's content and converts it to Markdown format.
// It implements the scrapemate.IJob interface and is designed to work with the scrapemate crawler.
//
// The job fetches the HTML content from the specified URL, converts it to Markdown,
// and stores the result in the Entry.Markdown field.
//
// Example usage:
//
//	entry := &Entry{
//	    WebSite: "https://example.com",
//	}
//	
//	// Create a new job with default settings
//	job := NewMarkdownJob("parent-123", entry)
//	
//	// Or with an exit monitor
//	monitor := &MyExitMonitor{}
//	job := NewMarkdownJob("parent-123", entry, WithMarkdownJobExitMonitor(monitor))
type MarkdownExtractJob struct {
	scrapemate.Job

	// Entry holds the data structure where the extracted Markdown will be stored
	Entry       *Entry
	ExitMonitor exiter.Exiter
	WorkspaceID uint64
}

func WithMarkdownJobWorkspaceID(workspaceID uint64) MarkdownExtractJobOptions {
	return func(j *MarkdownExtractJob) {
		j.WorkspaceID = workspaceID
	}
}

// NewMarkdownJob creates a new job for extracting Markdown from a webpage.
// It takes a parent ID for tracking job relationships, an Entry containing the target URL,
// and optional functional options for customizing the job's behavior.
//
// Example:
//
//	entry := &Entry{
//	    WebSite: "https://example.com",
//	}
//	
//	// Basic usage
//	job := NewMarkdownJob("parent-123", entry)
//	
//	// With exit monitoring
//	monitor := &MyExitMonitor{}
//	job := NewMarkdownJob(
//	    "parent-123",
//	    entry,
//	    WithMarkdownJobExitMonitor(monitor),
//	)
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
// The exit monitor is used to track job completion and manage crawler lifecycle.
//
// Example:
//
//	monitor := &MyExitMonitor{}
//	job := NewMarkdownJob(
//	    "parent-123",
//	    entry,
//	    WithMarkdownJobExitMonitor(monitor),
//	)
func WithMarkdownJobExitMonitor(exitMonitor exiter.Exiter) MarkdownExtractJobOptions {
	return func(j *MarkdownExtractJob) {
		j.ExitMonitor = exitMonitor
	}
}

// Process retrieves the HTML content and converts it to Markdown.
// It implements the scrapemate.IJob interface's Process method.
//
// The method will:
// 1. Clean up resources after processing
// 2. Update the exit monitor if configured
// 3. Convert the HTML content to Markdown using html-to-markdown
// 4. Store the result in Entry.Markdown
//
// Example usage through the scrapemate crawler:
//
//	crawler := scrapemate.NewCrawler(scrapemate.CrawlerConfig{
//	    Fetcher: &myFetcher{},
//	})
//	
//	job := NewMarkdownJob("parent-123", entry)
//	result, err := crawler.Process(context.Background(), job)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	
//	// Access the markdown
//	entry := result.(*Entry)
//	fmt.Println(entry.Markdown)
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
