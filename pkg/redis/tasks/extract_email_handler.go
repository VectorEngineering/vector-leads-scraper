// Package tasks provides functionality for handling various Redis-based tasks.
// It implements email extraction functionality that can crawl websites and extract
// email addresses from both mailto links and raw text content.
//
// Example usage of email extraction:
//
//	// Create a new handler with custom configuration
//	handler := tasks.NewHandler(
//		tasks.WithConcurrency(5),
//		tasks.WithTaskTimeout(2 * time.Minute),
//		tasks.WithProxies([]string{"http://proxy1.example.com", "http://proxy2.example.com"}),
//	)
//
//	// Create an email extraction task
//	task, err := tasks.CreateEmailTask(
//		"https://example.com",
//		3, // crawl up to 3 levels deep
//		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Process the task
//	ctx := context.Background()
//	if err := handler.ProcessTask(ctx, task); err != nil {
//		log.Fatal(err)
//	}
//
// The package also supports integration with asynq for background processing:
//
//	// Create an asynq server
//	srv := asynq.NewServer(
//		asynq.RedisClientOpt{Addr: "localhost:6379"},
//		asynq.Config{Concurrency: 10},
//	)
//
//	// Register the email task handler
//	mux := asynq.NewServeMux()
//	mux.Handle(tasks.TypeEmailExtract.String(), handler)
//
//	// Start the server
//	if err := srv.Run(mux); err != nil {
//		log.Fatal(err)
//	}
package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gosom/scrapemate"
	"github.com/gosom/scrapemate/scrapemateapp"
	"github.com/hibiken/asynq"
	"github.com/mcnijman/go-emailaddress"
)

// EmailJob represents a scrapemate job for email extraction from web pages.
// It implements the scrapemate.IJob interface and includes configuration
// for crawl depth and user agent settings.
//
// Example usage:
//
//	// Create a new email extraction job
//	job := NewEmailJob(
//		"https://example.com",
//		2,                    // crawl 2 levels deep
//		"Mozilla/5.0",        // user agent
//	)
//
//	// Create a scrapemate app for processing
//	cfg, _ := scrapemateapp.NewConfig(nil,
//		scrapemateapp.WithConcurrency(5),
//		scrapemateapp.WithExitOnInactivity(30 * time.Second),
//	)
//	mate, _ := scrapemateapp.NewScrapeMateApp(cfg)
//	defer mate.Close()
//
//	// Process the job
//	ctx := context.Background()
//	if err := mate.Start(ctx, job); err != nil {
//		log.Fatal(err)
//	}
type EmailJob struct {
	scrapemate.Job
	maxDepth  int      // Maximum depth for recursive crawling
	userAgent string   // User agent string for HTTP requests
}

// NewEmailJob creates a new email extraction job with the specified parameters.
// It initializes a job that will crawl the given URL up to maxDepth levels deep,
// using the specified userAgent for HTTP requests.
//
// Parameters:
//   - url: The starting URL to crawl for email addresses
//   - maxDepth: Maximum number of links to follow from the initial URL
//   - userAgent: The user agent string to use in HTTP requests
//
// Returns:
//   - scrapemate.IJob: A new email extraction job instance
func NewEmailJob(url string, maxDepth int, userAgent string) scrapemate.IJob {
	return &EmailJob{
		Job: scrapemate.Job{
			Method: "GET",
			URL:    url,
		},
		maxDepth:  maxDepth,
		userAgent: userAgent,
	}
}

// Process implements the scrapemate.IJob interface for email extraction.
// It processes a web page to extract email addresses using two methods:
// 1. Extracting from mailto links in the HTML
// 2. Using regex pattern matching on the page content
//
// The function will also generate new jobs for linked pages if maxDepth > 0.
//
// Parameters:
//   - ctx: Context for the operation
//   - resp: The scrapemate response containing the page content
//
// Returns:
//   - any: Map containing the extracted emails and the source URL
//   - []scrapemate.IJob: New jobs for crawling linked pages (if maxDepth > 0)
//   - error: Any error encountered during processing
func (j *EmailJob) Process(ctx context.Context, resp *scrapemate.Response) (any, []scrapemate.IJob, error) {
	// Get logger from context
	log := scrapemate.GetLoggerFromContext(ctx)
	log.Info("Processing email job", "url", j.URL)

	// if html fetch failed just return
	if resp.Error != nil {
		return nil, nil, nil
	}

	// Extract emails using both methods
	var emails []string

	// Try extracting from mailto links first
	if doc, ok := resp.Document.(*goquery.Document); ok {
		emails = docEmailExtractor(doc)
	}

	// If no emails found, try regex extraction
	if len(emails) == 0 {
		emails = regexEmailExtractor(resp.Body)
	}

	// Prepare result
	result := map[string]interface{}{
		"url":    j.URL,
		"emails": emails,
	}

	// If we haven't reached max depth, extract links and create new jobs
	if j.maxDepth > 0 && resp.Document != nil {
		var newJobs []scrapemate.IJob
		if doc, ok := resp.Document.(*goquery.Document); ok {
			baseURL, err := url.Parse(j.URL)
			if err != nil {
				return result, nil, fmt.Errorf("failed to parse base URL: %w", err)
			}

			doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
				if href, exists := s.Attr("href"); exists {
					parsedURL, err := url.Parse(href)
					if err != nil {
						return
					}
					absURL := baseURL.ResolveReference(parsedURL).String()
					newJobs = append(newJobs, NewEmailJob(absURL, j.maxDepth-1, j.userAgent))
				}
			})
		}
		return result, newJobs, nil
	}

	return result, nil, nil
}

// docEmailExtractor extracts email addresses from mailto links in an HTML document.
// It uses goquery to find all mailto links and validates each email address.
//
// Parameters:
//   - doc: The parsed HTML document
//
// Returns:
//   - []string: A deduplicated list of valid email addresses
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

// regexEmailExtractor finds email addresses in raw text using regex patterns.
// It uses the emailaddress package to find and validate email addresses.
//
// Parameters:
//   - body: Raw text content to search for email addresses
//
// Returns:
//   - []string: A deduplicated list of valid email addresses
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
//
// Parameters:
//   - s: The email address string to validate
//
// Returns:
//   - string: The normalized email address
//   - error: Any validation error encountered
func getValidEmail(s string) (string, error) {
	email, err := emailaddress.Parse(strings.TrimSpace(s))
	if err != nil {
		return "", err
	}
	return email.String(), nil
}

// CreateEmailTask creates a new asynq task for email extraction.
// It prepares the task payload with the specified parameters.
//
// Example usage:
//
//	// Create a task with default settings
//	task, err := CreateEmailTask(
//		"example.com",  // URL will be normalized to https://example.com
//		2,              // crawl 2 levels deep
//		"",             // empty user agent will use default
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create a task with custom settings
//	task, err = CreateEmailTask(
//		"https://example.com/contact",
//		5,              // deeper crawl
//		"CustomBot/1.0" // custom user agent
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Enqueue the task with asynq
//	client := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
//	info, err := client.Enqueue(task)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Enqueued task: id=%s queue=%s\n", info.ID, info.Queue)
func CreateEmailTask(url string, maxDepth int, userAgent string) (*asynq.Task, error) {
	payload := map[string]interface{}{
		"url":        url,
		"max_depth":  maxDepth,
		"user_agent": userAgent,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal email payload: %w", err)
	}
	return asynq.NewTask(TypeEmailExtract.String(), data), nil
}

// processEmailTask handles the execution of an email extraction task.
// It configures and runs a scrapemate job to crawl the specified URL
// and extract email addresses.
//
// Parameters:
//   - ctx: Context for the operation
//   - task: The asynq task containing the job parameters
//
// Returns:
//   - error: Any error encountered during task processing
func (h *Handler) processEmailTask(ctx context.Context, task *asynq.Task) error {
	var payload struct {
		URL       string `json:"url"`
		MaxDepth  int    `json:"max_depth"`
		UserAgent string `json:"user_agent"`
	}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal email payload: %w", err)
	}

	// Set default values if not provided
	if payload.MaxDepth == 0 {
		payload.MaxDepth = 2 // Default to crawling 2 levels deep
	}
	if payload.UserAgent == "" {
		payload.UserAgent = "Mozilla/5.0"
	}

	// Normalize URL
	if !strings.HasPrefix(payload.URL, "http") {
		payload.URL = "https://" + payload.URL
	}

	// Setup scrapemate
	opts := []func(*scrapemateapp.Config) error{
		scrapemateapp.WithConcurrency(h.concurrency),
		scrapemateapp.WithExitOnInactivity(h.taskTimeout),
	}

	if len(h.proxies) > 0 {
		opts = append(opts, scrapemateapp.WithProxies(h.proxies))
	}

	matecfg, err := scrapemateapp.NewConfig(nil, opts...)
	if err != nil {
		return fmt.Errorf("failed to create scrapemate config: %w", err)
	}

	mate, err := scrapemateapp.NewScrapeMateApp(matecfg)
	if err != nil {
		return fmt.Errorf("failed to create scrapemate app: %w", err)
	}
	defer mate.Close()

	// Create and run the email extraction job
	job := NewEmailJob(payload.URL, payload.MaxDepth, payload.UserAgent)
	if err := mate.Start(ctx, job); err != nil {
		if err != context.DeadlineExceeded && err != context.Canceled {
			return fmt.Errorf("failed to run email extraction: %w", err)
		}
	}

	return nil
}
