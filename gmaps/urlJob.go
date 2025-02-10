// Package gmaps provides functionality for scraping and processing Google Maps data.
package gmaps

import (
	"context"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Vector/vector-leads-scraper/exiter"
	"github.com/google/uuid"
	"github.com/gosom/scrapemate"
)

// URLExtractJobOptions defines options for configuring a URLExtractJob.
// It follows the functional options pattern for flexible job configuration.
type URLExtractJobOptions func(*URLExtractJob)

// URLExtractJob represents a job that extracts URLs from a webpage.
// It implements the scrapemate.IJob interface and is designed to work with the scrapemate crawler.
//
// The job scans the HTML content for both href and src attributes, collecting unique URLs
// and storing them in the Entry.Urls field.
//
// Example usage:
//
//	entry := &Entry{
//	    WebSite: "https://example.com",
//	}
//
//	// Create a new job with default settings
//	job := NewURLJob("parent-123", entry)
//
//	// Or with an exit monitor
//	monitor := &MyExitMonitor{}
//	job := NewURLJob("parent-123", entry, WithURLJobExitMonitor(monitor))
type URLExtractJob struct {
	scrapemate.Job

	// Entry holds the data structure where the extracted URLs will be stored
	Entry       *Entry
	ExitMonitor exiter.Exiter
	WorkspaceID uint64
}

func WithURLExtractJobWorkspaceID(workspaceID uint64) URLExtractJobOptions {
	return func(j *URLExtractJob) {
		j.WorkspaceID = workspaceID
	}
}

// NewURLJob creates a new job for URL extraction from a webpage.
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
//	job := NewURLJob("parent-123", entry)
//
//	// With exit monitoring
//	monitor := &MyExitMonitor{}
//	job := NewURLJob(
//	    "parent-123",
//	    entry,
//	    WithURLJobExitMonitor(monitor),
//	)
func NewURLJob(parentID string, entry *Entry, opts ...URLExtractJobOptions) *URLExtractJob {
	const (
		defaultPrio       = scrapemate.PriorityHigh
		defaultMaxRetries = 0
	)

	job := URLExtractJob{
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

// WithURLJobExitMonitor sets an exit monitor on the job.
// The exit monitor is used to track job completion and manage crawler lifecycle.
//
// Example:
//
//	monitor := &MyExitMonitor{}
//	job := NewURLJob(
//	    "parent-123",
//	    entry,
//	    WithURLJobExitMonitor(monitor),
//	)
func WithURLJobExitMonitor(exitMonitor exiter.Exiter) URLExtractJobOptions {
	return func(j *URLExtractJob) {
		j.ExitMonitor = exitMonitor
	}
}

// Process performs the URL extraction by operating on the goquery.Document.
// It implements the scrapemate.IJob interface's Process method.
//
// The method will:
// 1. Clean up resources after processing
// 2. Update the exit monitor if configured
// 3. Extract URLs from both href and src attributes
// 4. Store unique URLs in Entry.Urls
//
// Example usage through the scrapemate crawler:
//
//	crawler := scrapemate.NewCrawler(scrapemate.CrawlerConfig{
//	    Fetcher: &myFetcher{},
//	})
//
//	job := NewURLJob("parent-123", entry)
//	result, err := crawler.Process(context.Background(), job)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Access the extracted URLs
//	entry := result.(*Entry)
//	for _, url := range entry.Urls {
//	    fmt.Println(url)
//	}
func (j *URLExtractJob) Process(ctx context.Context, resp *scrapemate.Response) (any, []scrapemate.IJob, error) {
	// Clear the document and body after processing.
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
	log.Info("Processing URL extraction job", "url", j.URL)

	// If there was an error during the fetch, return immediately.
	if resp.Error != nil {
		return j.Entry, nil, nil
	}

	doc, ok := resp.Document.(*goquery.Document)
	if !ok {
		return j.Entry, nil, nil
	}

	// Extract URLs from the document.
	urls := docURLExtractor(doc)
	j.Entry.Urls = urls

	return j.Entry, nil, nil
}

// ProcessOnFetchError defines whether the job should be retried when a fetch error occurs.
func (j *URLExtractJob) ProcessOnFetchError() bool {
	return true
}

// docURLExtractor scans the document for URLs in both href and src attributes.
// It returns a slice of unique URLs found in the document.
//
// The function looks for:
// - href attributes in <a> tags and other elements
// - src attributes in <img>, <script>, and other elements
//
// Example:
//
//	doc, err := goquery.NewDocument("https://example.com")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	urls := docURLExtractor(doc)
//	for _, url := range urls {
//	    fmt.Println(url)
//	}
func docURLExtractor(doc *goquery.Document) []string {
	seen := make(map[string]bool)
	var urls []string

	// Find and process all elements with an "href" attribute.
	doc.Find("[href]").Each(func(_ int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			href = strings.TrimSpace(href)
			if href != "" && !seen[href] {
				urls = append(urls, href)
				seen[href] = true
			}
		}
	})

	// Find and process all elements with a "src" attribute.
	doc.Find("[src]").Each(func(_ int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			src = strings.TrimSpace(src)
			if src != "" && !seen[src] {
				urls = append(urls, src)
				seen[src] = true
			}
		}
	})

	return urls
}
