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
type GenericCrawlJobOptions func(*GenericCrawlJob)

// WithGenericDeduper sets the deduper for the job.
func WithGenericDeduper(d deduper.Deduper) GenericCrawlJobOptions {
	return func(j *GenericCrawlJob) {
		j.Deduper = d
	}
}

// WithGenericExitMonitor sets the exit monitor for the job.
func WithGenericExitMonitor(e exiter.Exiter) GenericCrawlJobOptions {
	return func(j *GenericCrawlJob) {
		j.ExitMonitor = e
	}
}

// GenericCrawlJob represents a generic crawl job that accepts any URL and extracts links
// recursively up to MaxDepth. It embeds scrapemate.Job.
type GenericCrawlJob struct {
	scrapemate.Job

	// MaxDepth is the maximum recursion depth.
	MaxDepth int
	// Depth is the current recursion level.
	Depth int

	// Deduper helps prevent processing the same URL multiple times.
	Deduper deduper.Deduper
	// ExitMonitor allows external tracking of progress.
	ExitMonitor exiter.Exiter
}

// NewGenericCrawlJob creates a new GenericCrawlJob.
// The provided urlStr is used directly (it can be any URL), and depth should start at 0.
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

// Process is called after the page is fetched and parsed (into a goquery.Document).
// It extracts all links (<a href="...">) and schedules new GenericCrawlJobs if the max depth has not been reached.
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
// and returns the absolute URL as a string.
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
