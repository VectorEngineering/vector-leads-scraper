// Package gmaps provides functionality for scraping and processing Google Maps data.
package gmaps

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/Vector/vector-leads-scraper/deduper"
	"github.com/Vector/vector-leads-scraper/exiter"
	"github.com/google/uuid"
	"github.com/gosom/scrapemate"
	"github.com/playwright-community/playwright-go"
)

// GenericCrawlJobOptions defines option setters for a GenericCrawlJob.
// It follows the functional options pattern for flexible job configuration.
type GenericCrawlJobOptions func(*GenericCrawlJob)

// WithGenericDeduper sets the deduper for the job.
// The deduper helps prevent processing the same URL multiple times during crawling.
//
// Example:
//
//	deduper := &MyDeduper{}
//	job := NewGenericCrawlJob(
//	    "job-123",
//	    "https://example.com",
//	    3, // maxDepth
//	    0, // currentDepth
//	    WithGenericDeduper(deduper),
//	)
func WithGenericDeduper(d deduper.Deduper) GenericCrawlJobOptions {
	return func(j *GenericCrawlJob) {
		j.Deduper = d
	}
}

// WithGenericExitMonitor sets the exit monitor for the job.
// The exit monitor is used to track job completion and manage crawler lifecycle.
//
// Example:
//
//	monitor := &MyExitMonitor{}
//	job := NewGenericCrawlJob(
//	    "job-123",
//	    "https://example.com",
//	    3, // maxDepth
//	    0, // currentDepth
//	    WithGenericExitMonitor(monitor),
//	)
func WithGenericExitMonitor(e exiter.Exiter) GenericCrawlJobOptions {
	return func(j *GenericCrawlJob) {
		j.ExitMonitor = e
	}
}

func WithGenericWorkspaceID(workspaceID uint64) GenericCrawlJobOptions {
	return func(j *GenericCrawlJob) {
		j.WorkspaceID = workspaceID
	}
}

// GenericCrawlJob represents a generic crawl job that accepts any URL and extracts links
// recursively up to MaxDepth. It implements the scrapemate.IJob interface and is designed
// to work with the scrapemate crawler.
//
// The job crawls web pages in a breadth-first manner, collecting links and creating new jobs
// for each discovered URL until the maximum depth is reached.
//
// Example usage:
//
//	// Create a new job with default settings
//	job := NewGenericCrawlJob(
//	    "job-123",
//	    "https://example.com",
//	    3, // maxDepth
//	    0, // currentDepth
//	)
//	
//	// Or with deduplication and monitoring
//	deduper := &MyDeduper{}
//	monitor := &MyExitMonitor{}
//	job := NewGenericCrawlJob(
//	    "job-123",
//	    "https://example.com",
//	    3, // maxDepth
//	    0, // currentDepth
//	    WithGenericDeduper(deduper),
//	    WithGenericExitMonitor(monitor),
//	)
type GenericCrawlJob struct {
	scrapemate.Job

	// MaxDepth is the maximum recursion depth for link extraction
	MaxDepth int
	// Depth is the current recursion level
	Depth int

	// Deduper helps prevent processing the same URL multiple times
	Deduper deduper.Deduper
	// ExitMonitor allows external tracking of progress
	ExitMonitor exiter.Exiter
	WorkspaceID uint64
}

// NewGenericCrawlJob creates a new GenericCrawlJob for crawling web pages.
// The provided urlStr is used directly (it can be any URL), and depth should start at 0.
//
// Parameters:
// - id: Unique identifier for the job (if empty, a UUID will be generated)
// - urlStr: The starting URL to crawl
// - maxDepth: Maximum recursion depth for link extraction
// - depth: Current recursion level (usually 0 for new jobs)
// - opts: Optional functional options for customizing the job
//
// Example:
//
//	// Basic usage
//	job := NewGenericCrawlJob(
//	    "job-123",
//	    "https://example.com",
//	    3, // maxDepth
//	    0, // currentDepth
//	)
//	
//	// With all options
//	job := NewGenericCrawlJob(
//	    "job-123",
//	    "https://example.com",
//	    3, // maxDepth
//	    0, // currentDepth
//	    WithGenericDeduper(deduper),
//	    WithGenericExitMonitor(monitor),
//	)
func NewGenericCrawlJob(id, urlStr string, maxDepth, depth int, opts ...GenericCrawlJobOptions) *GenericCrawlJob {
	if id == "" {
		id = uuid.New().String()
	}

	job := GenericCrawlJob{
		Job: scrapemate.Job{
			ID:         id,
			Method:     http.MethodGet,
			URL:        urlStr,
			MaxRetries: 3,
			Priority:   scrapemate.PriorityLow,
		},
		MaxDepth: maxDepth,
		Depth:    depth,
	}

	for _, opt := range opts {
		opt(&job)
	}

	return &job
}

// Process is called after the page is fetched and parsed.
// It implements the scrapemate.IJob interface's Process method.
//
// The method will:
// 1. Clean up resources after processing
// 2. Extract all links (<a href="...">) from the page
// 3. Create new jobs for discovered URLs if not at max depth
// 4. Update the exit monitor if configured
//
// Example usage through the scrapemate crawler:
//
//	crawler := scrapemate.NewCrawler(scrapemate.CrawlerConfig{
//	    Fetcher: &myFetcher{},
//	})
//	
//	job := NewGenericCrawlJob("job-123", "https://example.com", 3, 0)
//	result, err := crawler.Process(context.Background(), job)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (j *GenericCrawlJob) Process(ctx context.Context, resp *scrapemate.Response) (any, []scrapemate.IJob, error) {
	// Clean up resources.
	defer func() {
		resp.Document = nil
		resp.Body = nil
	}()

	// Ensure we have a goquery.Document.
	doc, ok := resp.Document.(*goquery.Document)
	if !ok {
		return nil, nil, fmt.Errorf("could not convert to goquery document")
	}

	var nextJobs []scrapemate.IJob

	// Use a generic selector: extract all <a href="..."> links.
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" {
			return
		}

		// Resolve relative URLs against the current page URL.
		resolvedURL := resolveURL(resp.URL, href)

		// Deduplication: skip if the URL was already seen.
		if j.Deduper != nil && !j.Deduper.AddIfNotExists(ctx, resolvedURL) {
			return
		}

		// If we haven't reached the maximum depth, schedule a new crawl job.
		if j.Depth < j.MaxDepth {
			nextJob := NewGenericCrawlJob(j.ID, resolvedURL, j.MaxDepth, j.Depth+1, WithGenericDeduper(j.Deduper), WithGenericExitMonitor(j.ExitMonitor))
			nextJobs = append(nextJobs, nextJob)
		}
	})

	// Optionally update the exit monitor.
	if j.ExitMonitor != nil {
		j.ExitMonitor.IncrPlacesFound(len(nextJobs))
		j.ExitMonitor.IncrSeedCompleted(1)
	}

	scrapemate.GetLoggerFromContext(ctx).Info(fmt.Sprintf("%d jobs found at depth %d", len(nextJobs), j.Depth))
	return nil, nextJobs, nil
}

// BrowserActions defines how the job interacts with a browser-based fetcher.
// It implements the scrapemate.IJob interface's BrowserActions method.
//
// The method will:
// 1. Navigate to the URL
// 2. Wait for the page to load
// 3. Collect response data
//
// This method is called when using a browser-based fetcher like Playwright.
func (j *GenericCrawlJob) BrowserActions(ctx context.Context, page playwright.Page) scrapemate.Response {
	var resp scrapemate.Response

	pageResponse, err := page.Goto(j.GetFullURL(), playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		resp.Error = err
		return resp
	}

	// Wait for the page to load.
	const defaultTimeout = 5000
	err = page.WaitForURL(page.URL(), playwright.PageWaitForURLOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(defaultTimeout),
	})
	if err != nil {
		resp.Error = err
		return resp
	}

	// Set basic response data.
	resp.URL = pageResponse.URL()
	resp.StatusCode = pageResponse.Status()
	resp.Headers = make(http.Header, len(pageResponse.Headers()))
	for k, v := range pageResponse.Headers() {
		resp.Headers.Add(k, v)
	}

	// Get the full page content.
	body, err := page.Content()
	if err != nil {
		resp.Error = err
		return resp
	}
	resp.Body = []byte(body)
	return resp
}

// resolveURL takes a base URL and a reference (which may be relative)
// and returns the absolute URL as a string. It handles both absolute and
// relative URLs correctly.
//
// Example:
//
//	base := "https://example.com/path/"
//	ref := "../other"
//	absolute := resolveURL(base, ref)
//	fmt.Println(absolute) // prints: https://example.com/other
func resolveURL(baseStr, ref string) string {
	base, err := url.Parse(baseStr)
	if err != nil {
		return ref
	}
	refURL, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return base.ResolveReference(refURL).String()
}
