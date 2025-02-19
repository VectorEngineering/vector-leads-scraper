package workers

import (
	"time"

	"github.com/Vector/vector-leads-scraper/internal/database"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// ProcessorWorker handles lead data processing and enrichment
type ProcessorWorker struct {
	db      *database.Db
	client  *asynq.Client
	logger  *zap.Logger
	Cadence struct {
		LeadProcessing  time.Duration // How often to process individual leads
		BatchProcessing time.Duration // How often to run batch processing
		MetricsUpdate   time.Duration // How often to update processing metrics
	}
}

// NewProcessorWorker creates a new processor worker
func NewProcessorWorker(db *database.Db, client *asynq.Client, logger *zap.Logger) *ProcessorWorker {
	worker := &ProcessorWorker{
		db:     db,
		client: client,
		logger: logger.Named("processor-worker"),
	}

	// Set default cadences
	worker.Cadence.LeadProcessing = time.Minute       // Every minute
	worker.Cadence.BatchProcessing = 10 * time.Minute // Every 10 minutes
	worker.Cadence.MetricsUpdate = 5 * time.Minute    // Every 5 minutes

	return worker
}

// // ProcessLead processes and enriches a single lead
// func (w *ProcessorWorker) ProcessLead(ctx context.Context, leadID string) error {
// 	w.logger.Info("Starting lead processing", zap.String("lead_id", leadID))

// 	// Get lead data from database
// 	lead, err := w.db.GetLead(ctx, leadID)
// 	if err != nil {
// 		return fmt.Errorf("failed to get lead data: %w", err)
// 	}

// 	// Enrich lead data
// 	enrichedLead, err := w.processor.EnrichLeadData(ctx, lead)
// 	if err != nil {
// 		w.logger.Error("Failed to enrich lead data",
// 			zap.Error(err),
// 			zap.String("lead_id", leadID),
// 		)
// 		return fmt.Errorf("failed to enrich lead data: %w", err)
// 	}

// 	// Validate enriched data
// 	if err := w.processor.ValidateLeadData(enrichedLead); err != nil {
// 		w.logger.Error("Lead data validation failed",
// 			zap.Error(err),
// 			zap.String("lead_id", leadID),
// 		)
// 		return fmt.Errorf("lead data validation failed: %w", err)
// 	}

// 	// Store enriched lead data
// 	if err := w.db.UpdateLead(ctx, enrichedLead); err != nil {
// 		return fmt.Errorf("failed to update lead data: %w", err)
// 	}

// 	// Queue for export if needed
// 	if enrichedLead.ReadyForExport {
// 		task := asynq.NewTask("export_lead", []byte(leadID))
// 		_, err = w.client.Enqueue(task, asynq.MaxRetry(3))
// 		if err != nil {
// 			w.logger.Error("Failed to queue lead for export",
// 				zap.Error(err),
// 				zap.String("lead_id", leadID),
// 			)
// 		}
// 	}

// 	w.logger.Info("Completed lead processing", zap.String("lead_id", leadID))
// 	return nil
// }

// // BatchProcessLeads processes multiple leads in batch
// func (w *ProcessorWorker) BatchProcessLeads(ctx context.Context, batchSize int) error {
// 	w.logger.Info("Starting batch lead processing", zap.Int("batch_size", batchSize))

// 	// Get batch of unprocessed leads
// 	leads, err := w.db.GetUnprocessedLeads(ctx, batchSize)
// 	if err != nil {
// 		return fmt.Errorf("failed to get unprocessed leads: %w", err)
// 	}

// 	successCount := 0
// 	failureCount := 0
// 	startTime := time.Now()

// 	for _, lead := range leads {
// 		if err := w.ProcessLead(ctx, lead.ID); err != nil {
// 			failureCount++
// 			w.logger.Error("Failed to process lead",
// 				zap.Error(err),
// 				zap.String("lead_id", lead.ID),
// 			)
// 			continue
// 		}
// 		successCount++
// 	}

// 	// Record batch processing metrics
// 	metrics := &processor.BatchMetrics{
// 		BatchSize:     len(leads),
// 		SuccessCount:  successCount,
// 		FailureCount:  failureCount,
// 		ProcessingTime: time.Since(startTime),
// 		Timestamp:     time.Now(),
// 	}

// 	if err := w.db.StoreBatchMetrics(ctx, metrics); err != nil {
// 		w.logger.Error("Failed to store batch metrics", zap.Error(err))
// 	}

// 	w.logger.Info("Completed batch lead processing",
// 		zap.Int("total_leads", len(leads)),
// 		zap.Int("success_count", successCount),
// 		zap.Int("failure_count", failureCount),
// 		zap.Duration("processing_time", time.Since(startTime)),
// 	)

// 	return nil
// }
