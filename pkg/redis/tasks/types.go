package tasks

import (
	"time"
)

// Package tasks provides types and structures for handling various Redis-based tasks.
// This file defines the payload structures for different types of tasks supported
// by the system, including Google Maps scraping and email extraction.
//
// Example usage of task payloads:
//
//	// Create and configure a Google Maps scraping task
//	scrapePayload := &tasks.ScrapePayload{
//		JobID: "restaurants-nyc-2024",
//		Keywords: []string{
//			"best restaurants in manhattan",
//			"top rated restaurants nyc",
//			"michelin star restaurants new york",
//		},
//		FastMode: true,        // Use fast scraping mode
//		Lang:     "en",        // Search in English
//		Depth:    3,           // Crawl up to 3 levels deep
//		Email:    true,        // Extract email addresses
//		Lat:      "40.7128",   // New York latitude
//		Lon:      "-74.0060",  // New York longitude
//		Zoom:     13,          // Map zoom level
//		Radius:   5000,        // 5km radius
//		MaxTime:  time.Hour,   // Maximum runtime
//		Proxies: []string{     // Use rotating proxies
//			"http://proxy1.example.com",
//			"http://proxy2.example.com",
//		},
//	}
//
//	// Create and configure an email extraction task
//	emailPayload := &tasks.EmailPayload{
//		URL:       "https://example.com",
//		MaxDepth:  2,                     // Crawl up to 2 levels deep
//		UserAgent: "CustomBot/1.0",       // Custom user agent
//	}

// ScrapePayload represents the payload for a scrape task.
// It contains all necessary configuration for scraping Google Maps,
// including search parameters, location data, and execution constraints.
//
// Example usage:
//
//	// Basic local business search
//	payload := &ScrapePayload{
//		JobID:    "local-cafes",
//		Keywords: []string{"cafes", "coffee shops"},
//		Lang:     "en",
//		Depth:    2,
//	}
//
//	// Advanced location-based search with timing
//	payload := &ScrapePayload{
//		JobID:    "sf-restaurants",
//		Keywords: []string{"restaurants in san francisco"},
//		FastMode: true,
//		Lang:     "en",
//		Depth:    3,
//		Email:    true,
//		Lat:      "37.7749",
//		Lon:      "-122.4194",
//		Zoom:     14,
//		Radius:   3000,
//		MaxTime:  30 * time.Minute,
//	}
type ScrapePayload struct {
	JobID    string        `json:"job_id"`             // Unique identifier for the scraping job
	Keywords []string      `json:"keywords"`           // List of search terms to scrape
	FastMode bool          `json:"fast_mode"`          // Enable fast scraping mode
	Lang     string        `json:"lang"`               // Language for search results
	Depth    int           `json:"depth"`              // Maximum crawl depth
	Email    bool          `json:"email"`              // Whether to extract email addresses
	Lat      string        `json:"lat,omitempty"`      // Latitude for location-based search
	Lon      string        `json:"lon,omitempty"`      // Longitude for location-based search
	Zoom     int           `json:"zoom,omitempty"`     // Map zoom level
	Radius   int           `json:"radius,omitempty"`   // Search radius in meters
	MaxTime  time.Duration `json:"max_time,omitempty"` // Maximum execution time
	Proxies  []string      `json:"proxies,omitempty"`  // List of proxy servers to use
}

// EmailPayload represents the payload for an email extraction task.
// It contains configuration for crawling websites and extracting email addresses.
//
// Example usage:
//
//	// Basic email extraction from a single page
//	payload := &EmailPayload{
//		URL: "https://example.com/contact",
//	}
//
//	// Advanced email extraction with crawling
//	payload := &EmailPayload{
//		URL:       "https://example.com",
//		MaxDepth:  3,                    // Crawl up to 3 levels deep
//		UserAgent: "EmailBot/1.0",       // Custom user agent
//	}
type EmailPayload struct {
	URL       string `json:"url"`                  // Target URL to extract emails from
	MaxDepth  int    `json:"max_depth,omitempty"`  // Maximum crawl depth
	UserAgent string `json:"user_agent,omitempty"` // Custom user agent string
}
