package resultwriter

import (
	"context"
	"errors"
	"sync"
	"time"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/gosom/scrapemate"
	"go.uber.org/zap"

	"github.com/Vector/vector-leads-scraper/gmaps"
	"github.com/Vector/vector-leads-scraper/internal/database"
	"github.com/Vector/vector-leads-scraper/pkg/webhook"
)

var (
	ErrInvalidDB = errors.New("invalid database connection")
)

// ResultWriter implements scrapemate.ResultWriter interface using GORM
type ResultWriter struct {
	db     *database.Db
	logger *zap.Logger
	cfg    *Config

	webhookClient webhook.Client
	closeOnce     sync.Once
}

// New creates a new ResultWriter instance
func New(db *database.Db, logger *zap.Logger, cfg *Config) (*ResultWriter, error) {
	if db == nil {
		return nil, ErrInvalidDB
	}

	if cfg == nil {
		cfg = DefaultConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	w := &ResultWriter{
		db:     db,
		logger: logger,
		cfg:    cfg,
	}

	// Initialize webhook client if enabled
	if cfg.WebhookEnabled && len(cfg.WebhookEndpoints) > 0 {
		webhookCfg := webhook.Config{
			Endpoints:     cfg.WebhookEndpoints,
			MaxBatchSize: cfg.WebhookBatchSize,
			BatchTimeout: cfg.WebhookFlushInterval,
			RetryConfig: webhook.RetryConfig{
				MaxRetries:        cfg.WebhookRetryMax,
				InitialBackoff:    cfg.WebhookRetryInterval,
				MaxBackoff:        cfg.WebhookRetryInterval * 10,
				BackoffMultiplier: 2,
			},
		}

		webhookClient, err := webhook.NewClient(webhookCfg)
		if err != nil {
			return nil, err
		}
		w.webhookClient = webhookClient
	}

	return w, nil
}

func (r *ResultWriter) Run(ctx context.Context, in <-chan scrapemate.Result) error {
	const maxBatchSize = 50

	// Map of job ID to leads buffer
	buffers := make(map[uint64][]*lead_scraper_servicev1.Lead)
	lastSave := time.Now().UTC()

	for result := range in {
		entry, ok := result.Data.(*gmaps.Entry)
		if !ok {
			return errors.New("invalid data type")
		}
		
		runnableJob, ok := result.Job.(*gmaps.GmapJob)
		if !ok {
			return errors.New("invalid job type")
		}

		jobID := runnableJob.ScrapingJobID
		buffers[jobID] = append(buffers[jobID], convertEntryToLead(entry))

		// Check if any buffer needs to be flushed
		if len(buffers[jobID]) >= maxBatchSize || time.Now().UTC().Sub(lastSave) >= time.Minute {
			// Flush all buffers
			for id, buff := range buffers {
				if len(buff) > 0 {
					if _, err := r.db.BatchCreateLeads(ctx, id, buff); err != nil {
						r.logger.Error("failed to write leads", 
							zap.Error(err),
							zap.Uint64("job_id", id),
							zap.Int("count", len(buff)))
						return err
					}
					buffers[id] = buffers[id][:0]
				}
			}
			lastSave = time.Now().UTC()
		}
	}

	// Flush any remaining buffers
	for id, buff := range buffers {
		if len(buff) > 0 {
			if _, err := r.db.BatchCreateLeads(ctx, id, buff); err != nil {
				r.logger.Error("failed to write leads", 
					zap.Error(err),
					zap.Uint64("job_id", id),
					zap.Int("count", len(buff)))
				return err
			}
		}
	}

	// Send to webhook if enabled
	if r.cfg.WebhookEnabled && r.webhookClient != nil {
		// Combine all leads for webhook notification
		var allLeads []*lead_scraper_servicev1.Lead
		for _, buff := range buffers {
			allLeads = append(allLeads, buff...)
		}
		if err := r.webhookClient.Send(ctx, webhook.Record{Data: allLeads}); err != nil {
			r.logger.Error("failed to send webhook notification",
				zap.Error(err))
		}
	}

	return nil
}

// Close closes the result writer
func (w *ResultWriter) Close() error {
	var err error
	w.closeOnce.Do(func() {
		if w.webhookClient != nil {
			err = w.webhookClient.Shutdown(context.Background())
		}
	})
	return err
}

// convertEntryToLead converts a gmaps.Entry to a lead_scraper_servicev1.Lead
func convertEntryToLead(entry *gmaps.Entry) *lead_scraper_servicev1.Lead {
	lead := &lead_scraper_servicev1.Lead{
		// Basic business details
		Name:          entry.Title,
		Website:       entry.WebSite,
		Phone:         entry.Phone,
		Address:       entry.Address,
		City:         entry.CompleteAddress.City,
		State:        entry.CompleteAddress.State,
		Country:      entry.CompleteAddress.Country,
		
		// Location data
		Latitude:     entry.Latitude,
		Longitude:    entry.Longtitude,
		
		// Google-specific data
		GoogleRating:  float32(entry.ReviewRating),
		ReviewCount:   int32(entry.ReviewCount),
		PlaceId:      entry.Cid,
		GoogleMapsUrl: entry.Link,
		
		// Business categorization
		Industry:                entry.Category,
		GoogleMyBusinessCategory: entry.Category,
		Types:                   entry.Categories,
		
		// Business status and hours
		BusinessStatus: entry.Status,
		
		// Reviews and ratings
		Rating:         float32(entry.ReviewRating),
		RatingCategory: entry.Category,
		Count:          int32(entry.ReviewCount),
		
		// Media
		MainPhotoUrl: entry.Thumbnail,
		
		// Additional metadata
		Timezone:     entry.Timezone,
		BusinessType: "UNSPECIFIED", // Default value
		
		// Contact info
		AlternatePhones: []string{entry.Phone}, // Add main phone as alternate
	}

	// Add business hours if available
	if len(entry.OpenHours) > 0 {
		regularHours := make([]*lead_scraper_servicev1.BusinessHours, 0, len(entry.OpenHours))
		for day, schedules := range entry.OpenHours {
			if len(schedules) > 0 {
				regularHours = append(regularHours, &lead_scraper_servicev1.BusinessHours{
					Day:       convertDayOfWeek(day),
					OpenTime:  schedules[0],
					CloseTime: schedules[len(schedules)-1],
					Closed:    len(schedules) == 0,
				})
			}
		}
		lead.RegularHours = regularHours
	}

	return lead
}

// convertDayOfWeek converts a day string to BusinessHours_DayOfWeek enum
func convertDayOfWeek(day string) lead_scraper_servicev1.BusinessHours_DayOfWeek {
	switch day {
	case "Monday":
		return lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_MONDAY
	case "Tuesday":
		return lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_TUESDAY
	case "Wednesday":
		return lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_WEDNESDAY
	case "Thursday":
		return lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_THURSDAY
	case "Friday":
		return lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_FRIDAY
	case "Saturday":
		return lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_SATURDAY
	case "Sunday":
		return lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_SUNDAY
	default:
		return lead_scraper_servicev1.BusinessHours_DAY_OF_WEEK_UNSPECIFIED
	}
}

// getMainPhotoURL returns the first photo URL as the main photo if available
func getMainPhotoURL(photos []string) string {
	if len(photos) > 0 {
		return photos[0]
	}
	return ""
} 