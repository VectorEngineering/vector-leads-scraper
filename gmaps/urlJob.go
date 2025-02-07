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
type URLExtractJobOptions func(*URLExtractJob)

// URLExtractJob represents a job that extracts URLs from a webpage.
type URLExtractJob struct {
	scrapemate.Job

	// Entry should have a field (e.g., Urls []string) to hold the extracted URLs.
	Entry       *Entry
	ExitMonitor exiter.Exiter
}

// NewURLJob creates a new job for URL extraction.
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
func WithURLJobExitMonitor(exitMonitor exiter.Exiter) URLExtractJobOptions {
	return func(j *URLExtractJob) {
		j.ExitMonitor = exitMonitor
	}
}

// Process performs the URL extraction by operating on the goquery.Document.
// It follows a similar pattern to your email job.
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
