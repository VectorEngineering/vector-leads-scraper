// Package gmaps provides functionality for scraping and processing Google Maps data.
package gmaps

import (
	"context"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Vector/vector-leads-scraper/exiter"
	"github.com/google/uuid"
	"github.com/gosom/scrapemate"
	"github.com/mcnijman/go-emailaddress"
)

// EmailExtractJobOptions defines options for configuring an EmailExtractJob.
// It follows the functional options pattern for flexible job configuration.
type EmailExtractJobOptions func(*EmailExtractJob)

// EmailExtractJob represents a job that extracts email addresses from a webpage.
// It implements the scrapemate.IJob interface and is designed to work with the scrapemate crawler.
//
// The job uses two methods to find email addresses:
// 1. Looking for mailto: links in the HTML
// 2. Using regex pattern matching on the page content
//
// Example usage:
//
//	entry := &Entry{
//	    WebSite: "https://example.com",
//	}
//	
//	// Create a new job with default settings
//	job := NewEmailJob("parent-123", entry)
//	
//	// Or with an exit monitor
//	monitor := &MyExitMonitor{}
//	job := NewEmailJob("parent-123", entry, WithEmailJobExitMonitor(monitor))
type EmailExtractJob struct {
	scrapemate.Job

	// Entry holds the data structure where the extracted email addresses will be stored
	Entry       *Entry
	ExitMonitor exiter.Exiter
	WorkspaceID uint64
}

func WithEmailJobWorkspaceID(workspaceID uint64) EmailExtractJobOptions {
	return func(j *EmailExtractJob) {
		j.WorkspaceID = workspaceID
	}
}

// NewEmailJob creates a new job for email address extraction from a webpage.
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
//	job := NewEmailJob("parent-123", entry)
//	
//	// With exit monitoring
//	monitor := &MyExitMonitor{}
//	job := NewEmailJob(
//	    "parent-123",
//	    entry,
//	    WithEmailJobExitMonitor(monitor),
//	)
func NewEmailJob(parentID string, entry *Entry, opts ...EmailExtractJobOptions) *EmailExtractJob {
	const (
		defaultPrio       = scrapemate.PriorityHigh
		defaultMaxRetries = 0
	)

	job := EmailExtractJob{
		Job: scrapemate.Job{
			ID:         uuid.New().String(),
			ParentID:   parentID,
			Method:     "GET",
			URL:        entry.WebSite,
			MaxRetries: defaultMaxRetries,
			Priority:   defaultPrio,
		},
	}

	job.Entry = entry

	for _, opt := range opts {
		opt(&job)
	}

	return &job
}

// WithEmailJobExitMonitor sets an exit monitor on the job.
// The exit monitor is used to track job completion and manage crawler lifecycle.
//
// Example:
//
//	monitor := &MyExitMonitor{}
//	job := NewEmailJob(
//	    "parent-123",
//	    entry,
//	    WithEmailJobExitMonitor(monitor),
//	)
func WithEmailJobExitMonitor(exitMonitor exiter.Exiter) EmailExtractJobOptions {
	return func(j *EmailExtractJob) {
		j.ExitMonitor = exitMonitor
	}
}

// Process performs the email extraction by operating on the goquery.Document.
// It implements the scrapemate.IJob interface's Process method.
//
// The method will:
// 1. Clean up resources after processing
// 2. Update the exit monitor if configured
// 3. Try to extract emails from mailto: links
// 4. If no emails found, try regex pattern matching on the page content
// 5. Store unique email addresses in Entry.Emails
//
// Example usage through the scrapemate crawler:
//
//	crawler := scrapemate.NewCrawler(scrapemate.CrawlerConfig{
//	    Fetcher: &myFetcher{},
//	})
//	
//	job := NewEmailJob("parent-123", entry)
//	result, err := crawler.Process(context.Background(), job)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	
//	// Access the extracted emails
//	entry := result.(*Entry)
//	for _, email := range entry.Emails {
//	    fmt.Println(email)
//	}
func (j *EmailExtractJob) Process(ctx context.Context, resp *scrapemate.Response) (any, []scrapemate.IJob, error) {
	defer func() {
		resp.Document = nil
		resp.Body = nil
	}()

	defer func() {
		if j.ExitMonitor != nil {
			j.ExitMonitor.IncrPlacesCompleted(1)
		}
	}()

	log := scrapemate.GetLoggerFromContext(ctx)

	log.Info("Processing email job", "url", j.URL)

	// if html fetch failed just return
	if resp.Error != nil {
		return j.Entry, nil, nil
	}

	doc, ok := resp.Document.(*goquery.Document)
	if !ok {
		return j.Entry, nil, nil
	}

	emails := docEmailExtractor(doc)
	if len(emails) == 0 {
		emails = regexEmailExtractor(resp.Body)
	}

	j.Entry.Emails = emails

	return j.Entry, nil, nil
}

func (j *EmailExtractJob) ProcessOnFetchError() bool {
	return true
}

// docEmailExtractor extracts email addresses from mailto: links in the HTML document.
// It returns a slice of unique, valid email addresses.
//
// Example:
//
//	doc, err := goquery.NewDocument("https://example.com")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	
//	emails := docEmailExtractor(doc)
//	for _, email := range emails {
//	    fmt.Println(email)
//	}
func docEmailExtractor(doc *goquery.Document) []string {
	seen := map[string]bool{}

	var emails []string

	doc.Find("a[href^='mailto:']").Each(func(_ int, s *goquery.Selection) {
		mailto, exists := s.Attr("href")
		if exists {
			value := strings.TrimPrefix(mailto, "mailto:")
			if email, err := getValidEmail(value); err == nil {
				if !seen[email] {
					emails = append(emails, email)
					seen[email] = true
				}
			}
		}
	})

	return emails
}

// regexEmailExtractor finds email addresses in raw text content using regex pattern matching.
// It returns a slice of unique, valid email addresses.
//
// Example:
//
//	content := []byte("Contact us at support@example.com or sales@example.com")
//	emails := regexEmailExtractor(content)
//	for _, email := range emails {
//	    fmt.Println(email)
//	}
func regexEmailExtractor(body []byte) []string {
	seen := map[string]bool{}

	var emails []string

	addresses := emailaddress.Find(body, false)
	for i := range addresses {
		if !seen[addresses[i].String()] {
			emails = append(emails, addresses[i].String())
			seen[addresses[i].String()] = true
		}
	}

	return emails
}

// getValidEmail validates and normalizes an email address string.
// It returns the normalized email address and an error if the email is invalid.
//
// Example:
//
//	email, err := getValidEmail("  user@example.com  ")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(email) // prints: user@example.com
func getValidEmail(s string) (string, error) {
	email, err := emailaddress.Parse(strings.TrimSpace(s))
	if err != nil {
		return "", err
	}

	return email.String(), nil
}
