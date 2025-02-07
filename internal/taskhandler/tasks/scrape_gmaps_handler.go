package tasks

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Vector/vector-leads-scraper/deduper"
	"github.com/Vector/vector-leads-scraper/exiter"
	"github.com/Vector/vector-leads-scraper/runner"
	"github.com/gosom/scrapemate"
	"github.com/gosom/scrapemate/adapters/writers/csvwriter"
	"github.com/gosom/scrapemate/scrapemateapp"
	"github.com/hibiken/asynq"
)

// Package tasks provides functionality for handling Redis-based tasks,
// including Google Maps scraping capabilities. This file implements the Google Maps
// scraping functionality with support for various search parameters and configurations.
//
// Example usage of Google Maps scraping:
//
//	// Create a handler with scraping-specific configuration
//	handler := tasks.NewHandler(
//		tasks.WithConcurrency(10),
//		tasks.WithTaskTimeout(30 * time.Minute),
//		tasks.WithDataFolder("/path/to/results"),
//		tasks.WithProxies([]string{
//			"http://proxy1.example.com",
//			"http://proxy2.example.com",
//		}),
//	)
//
//	// Create a scraping task for restaurants in New York
//	payload := &tasks.ScrapePayload{
//		JobID:    "nyc-restaurants-2024",
//		Keywords: []string{"restaurants in new york", "cafes in manhattan"},
//		FastMode: true,
//		Lang:     "en",
//		Depth:    3,
//		Email:    true,
//		Lat:      "40.7128",
//		Lon:      "-74.0060",
//		Zoom:     13,
//		Radius:   5000,  // 5km radius
//		MaxTime:  2 * time.Hour,
//	}
//
//	task, err := tasks.CreateScrapeTask(payload)
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
// Example of background processing with asynq:
//
//	// Create an asynq client
//	client := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
//	defer client.Close()
//
//	// Create and enqueue multiple scraping tasks
//	cities := []struct {
//		name string
//		lat  string
//		lon  string
//	}{
//		{"new-york", "40.7128", "-74.0060"},
//		{"london", "51.5074", "-0.1278"},
//		{"tokyo", "35.6762", "139.6503"},
//	}
//
//	for _, city := range cities {
//		payload := &tasks.ScrapePayload{
//			JobID:    fmt.Sprintf("restaurants-%s-2024", city.name),
//			Keywords: []string{"best restaurants", "popular cafes"},
//			Lang:     "en",
//			Depth:    2,
//			Email:    true,
//			Lat:      city.lat,
//			Lon:      city.lon,
//			Zoom:     12,
//			Radius:   10000,
//		}
//
//		task, _ := tasks.CreateScrapeTask(payload)
//		info, err := client.Enqueue(task,
//			asynq.Queue("scraping"),
//			asynq.MaxRetry(3),
//			asynq.Timeout(hour),
//		)
//		if err != nil {
//			log.Printf("Failed to enqueue task for %s: %v", city.name, err)
//			continue
//		}
//		log.Printf("Enqueued task for %s: id=%s queue=%s",
//			city.name, info.ID, info.Queue)
//	}

// CreateScrapeTask creates a new scrape task with the given payload.
// It prepares a task for scraping Google Maps based on the provided configuration.
//
// Example usage:
//
//	// Basic scraping task
//	payload := &tasks.ScrapePayload{
//		JobID:    "cafes-search",
//		Keywords: []string{"cafes near central park"},
//		Lang:     "en",
//		Depth:    2,
//	}
//	task, err := tasks.CreateScrapeTask(payload)
//
//	// Advanced scraping task with location and timing
//	payload = &tasks.ScrapePayload{
//		JobID:    "restaurants-manhattan",
//		Keywords: []string{"fine dining manhattan"},
//		FastMode: true,
//		Lang:     "en",
//		Depth:    3,
//		Email:    true,
//		Lat:      "40.7831",
//		Lon:      "-73.9712",
//		Zoom:     14,
//		Radius:   2000,
//		MaxTime:  45 * time.Minute,
//		Proxies:  []string{"http://proxy.example.com"},
//	}
//	task, err = tasks.CreateScrapeTask(payload)
func CreateScrapeTask(payload *ScrapePayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scrape payload: %w", err)
	}
	return asynq.NewTask(TypeScrapeGMaps.String(), data), nil
}

// processScrapeTask processes a scrape task
func (h *Handler) processScrapeTask(ctx context.Context, task *asynq.Task) error {
	var payload ScrapePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal scrape payload: %w", err)
	}

	if len(payload.Keywords) == 0 {
		return fmt.Errorf("no keywords provided")
	}

	outpath := filepath.Join(h.dataFolder, payload.JobID+".csv")
	outfile, err := os.Create(outpath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outfile.Close()

	mate, err := h.setupMate(ctx, outfile, &payload)
	if err != nil {
		return fmt.Errorf("failed to setup scrapemate: %w", err)
	}
	defer mate.Close()

	var coords string
	if payload.Lat != "" && payload.Lon != "" {
		coords = payload.Lat + "," + payload.Lon
	}

	dedup := deduper.New()
	exitMonitor := exiter.New()

	seedJobs, err := runner.CreateSeedJobs(
		payload.FastMode,
		payload.Lang,
		strings.NewReader(strings.Join(payload.Keywords, "\n")),
		payload.Depth,
		payload.Email,
		coords,
		payload.Zoom,
		func() float64 {
			if payload.Radius <= 0 {
				return 10000 // 10 km
			}
			return float64(payload.Radius)
		}(),
		dedup,
		exitMonitor,
	)
	if err != nil {
		return fmt.Errorf("failed to create seed jobs: %w", err)
	}

	if len(seedJobs) > 0 {
		exitMonitor.SetSeedCount(len(seedJobs))

		allowedSeconds := max(60, len(seedJobs)*10*payload.Depth/50+120)
		if payload.MaxTime > 0 {
			if payload.MaxTime.Seconds() < 180 {
				allowedSeconds = 180
			} else {
				allowedSeconds = int(payload.MaxTime.Seconds())
			}
		}

		log.Printf("running job %s with %d seed jobs and %d allowed seconds", payload.JobID, len(seedJobs), allowedSeconds)

		jobCtx, jobCancel := context.WithTimeout(ctx, time.Duration(allowedSeconds)*time.Second)
		defer jobCancel()

		exitMonitor.SetCancelFunc(jobCancel)
		go exitMonitor.Run(jobCtx)

		if err := mate.Start(jobCtx, seedJobs...); err != nil {
			if err != context.DeadlineExceeded && err != context.Canceled {
				return fmt.Errorf("failed to run scraping: %w", err)
			}
		}
	}

	return nil
}

// setupMate sets up the scrapemate app with the given writer and payload
func (h *Handler) setupMate(_ context.Context, writer io.Writer, payload *ScrapePayload) (*scrapemateapp.ScrapemateApp, error) {
	opts := []func(*scrapemateapp.Config) error{
		scrapemateapp.WithConcurrency(h.concurrency),
		scrapemateapp.WithExitOnInactivity(time.Minute * 3),
	}

	if !payload.FastMode {
		opts = append(opts,
			scrapemateapp.WithJS(scrapemateapp.DisableImages()),
		)
	} else {
		opts = append(opts,
			scrapemateapp.WithStealth("firefox"),
		)
	}

	hasProxy := false

	if len(h.proxies) > 0 {
		opts = append(opts, scrapemateapp.WithProxies(h.proxies))
		hasProxy = true
	} else if len(payload.Proxies) > 0 {
		opts = append(opts,
			scrapemateapp.WithProxies(payload.Proxies),
		)
		hasProxy = true
	}

	if !h.disableReuse {
		opts = append(opts,
			scrapemateapp.WithPageReuseLimit(2),
			scrapemateapp.WithPageReuseLimit(200),
		)
	}

	log.Printf("job %s has proxy: %v", payload.JobID, hasProxy)

	csvWriter := csvwriter.NewCsvWriter(csv.NewWriter(writer))
	writers := []scrapemate.ResultWriter{csvWriter}

	matecfg, err := scrapemateapp.NewConfig(
		writers,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return scrapemateapp.NewScrapeMateApp(matecfg)
}
