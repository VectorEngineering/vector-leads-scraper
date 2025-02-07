package resultwriter

import (
	"context"
	"fmt"
	"sync"
	"time"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/gosom/scrapemate"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Vector/vector-leads-scraper/gmaps"
	"github.com/Vector/vector-leads-scraper/internal/database"
)

const (
	defaultBatchSize = 50
	defaultWorkspaceID = 1 // Default workspace ID for testing
)

// Provider implements scrapemate.JobProvider interface
type Provider struct {
	db        *database.Db
	logger    *zap.Logger
	mu        *sync.Mutex
	jobc      chan scrapemate.IJob
	errc      chan error
	started   bool
	batchSize int
}

// NewProvider creates a new job provider
func NewProvider(db *database.Db, logger *zap.Logger, opts ...ProviderOption) *Provider {
	if logger == nil {
		logger = zap.NewNop()
	}

	p := &Provider{
		db:          db,
		logger:      logger,
		mu:          &sync.Mutex{},
		errc:        make(chan error, 1),
		batchSize:   defaultBatchSize,
	}

	for _, opt := range opts {
		opt(p)
	}

	p.jobc = make(chan scrapemate.IJob, 2*p.batchSize)

	return p
}

// ProviderOption allows configuring the provider
type ProviderOption func(*Provider)

// WithBatchSize sets custom batch size
func WithBatchSize(size int) ProviderOption {
	return func(p *Provider) {
		if size > 0 {
			p.batchSize = size
		}
	}
}

// Jobs implements scrapemate.JobProvider interface
func (p *Provider) Jobs(ctx context.Context) (<-chan scrapemate.IJob, <-chan error) {
	outc := make(chan scrapemate.IJob)
	errc := make(chan error, 1)

	p.mu.Lock()
	if !p.started {
		go p.fetchJobs(ctx)
		p.started = true
	}
	p.mu.Unlock()

	go func() {
		defer close(outc)
		defer close(errc)

		for {
			select {
			case <-ctx.Done():
				return
			case err := <-p.errc:
				errc <- err
				return
			case job, ok := <-p.jobc:
				if !ok {
					return
				}
				if job == nil || job.GetID() == "" {
					continue
				}

				select {
				case outc <- job:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return outc, errc
}

// Push implements scrapemate.JobProvider interface
func (p *Provider) Push(ctx context.Context, job scrapemate.IJob) error {
	gmapJob, ok := job.(*gmaps.GmapJob)
	if !ok {
		return fmt.Errorf("invalid job type: expected *gmaps.GmapJob")
	}

	scrapingJob := &lead_scraper_servicev1.ScrapingJob{
		Name:        job.GetID(),
		Status:      lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_UNSPECIFIED,
		Priority:    int32(job.GetPriority()),
		PayloadType: "scraping_job",
		CreatedAt:   timestamppb.Now(),
		Url:         gmapJob.GetURL(),
	}

	_, err := p.db.CreateScrapingJob(ctx, gmapJob.WorkspaceID, scrapingJob)
	if err != nil {
		return fmt.Errorf("failed to create scraping job: %w", err)
	}

	return nil
}

func (p *Provider) fetchJobs(ctx context.Context) {
	defer close(p.jobc)
	defer close(p.errc)

	baseDelay := time.Millisecond * 50
	maxDelay := time.Millisecond * 300
	factor := 2
	currentDelay := baseDelay

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Get jobs in NEW status
		jobs, err := p.db.ListScrapingJobs(ctx, uint64(p.batchSize), 0)
		if err != nil {
			p.errc <- fmt.Errorf("failed to list scraping jobs: %w", err)
			return
		}

		if len(jobs) > 0 {
			// Update jobs to QUEUED status
			for _, job := range jobs {
				if job.Status != lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_UNSPECIFIED {
					continue
				}

				job.Status = lead_scraper_servicev1.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED
				if _, err := p.db.UpdateScrapingJob(ctx, job); err != nil {
					p.errc <- fmt.Errorf("failed to update scraping job status: %w", err)
					return
				}

				// Convert to scrapemate.IJob and send to channel
				scrapemateJob := &scrapemate.Job{
					ID:       job.Name,
					Priority: int(job.Priority),
				}

				select {
				case p.jobc <- scrapemateJob:
				case <-ctx.Done():
					return
				}
			}
		} else {
			// No jobs found, apply backoff
			select {
			case <-time.After(currentDelay):
				currentDelay = time.Duration(float64(currentDelay) * float64(factor))
				if currentDelay > maxDelay {
					currentDelay = maxDelay
				}
			case <-ctx.Done():
				return
			}
		}
	}
} 