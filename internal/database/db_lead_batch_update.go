package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
)

// BatchUpdateLeads updates multiple leads in a single batch operation.
// It processes leads in batches of 100 to avoid overwhelming the database.
// If an error occurs during batch processing, it will be logged and the operation
// will continue with the next batch.
func (db *Db) BatchUpdateLeads(ctx context.Context, leads []*lead_scraper_servicev1.Lead) (bool, error) {
	var (
		qOp = db.QueryOperator.LeadORM
	)

	if len(leads) == 0 {
		return false, ErrInvalidInput
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// Start a transaction
	tx := db.Client.Engine.WithContext(ctx).Begin()
	if tx.Error != nil {
		return false, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	batches := BreakIntoBatches[*lead_scraper_servicev1.Lead](leads, batchSize)

	// Process each batch
	for _, batch := range batches {
			// convert batch to ORM objects
		ormLeads, err := db.convertLeadsToORM(ctx, batch)
		if err != nil {
			return false, fmt.Errorf("failed to convert leads to ORM: %w", err)
		}

		// get all the ids from the batch
		ids := make([]uint64, 0, len(batch))
		for _, lead := range batch {
			ids = append(ids, lead.Id)
		}

		result, err := qOp.WithContext(ctx).Where(qOp.Id.In(ids...)).Updates(ormLeads)
		if err != nil {
			db.Logger.Error("failed to update batch",
				zap.Error(err),
				zap.Int("batchSize", len(batch)),
			)
			continue
		}
		
		if result.Error != nil {
			db.Logger.Error("failed to update batch",
				zap.Error(result.Error),
				zap.Int("batchSize", len(batch)),
			)
			continue
		}
	}

	return true, nil
} 

func (db *Db) convertLeadsToORM(ctx context.Context, leads []*lead_scraper_servicev1.Lead) ([]*lead_scraper_servicev1.LeadORM, error) {
	ormLeads := make([]*lead_scraper_servicev1.LeadORM, 0, len(leads))
	for _, lead := range leads {
		ormLead, err := lead.ToORM(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to convert lead to ORM: %w", err)
		}
		ormLeads = append(ormLeads, &ormLead)
	}
	return ormLeads, nil
}			