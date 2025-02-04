package database

import (
	"context"
	"fmt"

	"github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/dal"
	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	user "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

const (
	batchSize = 100
)

// PreloadUserAccount preloads related data for a user account query based on the provided query reference.
// It preloads associated entities such as Address, Tags, and various settings for a complete data load.
//
// Parameters:
//   - queryRef: Reference to the user account ORM query interface.
//
// Returns:
//   - *schema.UserAccountORM: Preloaded user account ORM object if found.
//   - error: Non-nil error if query fails.
func (db *Db) PreloadAccount(queryRef dal.IAccountORMDo) (*user.AccountORM, error) {
	u := db.QueryOperator.AccountORM
	return queryRef.
		Preload(u.Workspaces.ScrapingJobs).
		Preload(u.Workspaces.Workflows).
		Preload(u.Workspaces.ScrapingJobs.Leads).
		Preload(u.Settings).
		First()
}

// BreakIntoBatches splits a slice of any type into smaller batches of the specified size.
// Type parameter T can be any type (uint64, string, custom structs, etc.)
func BreakIntoBatches[T any](items []T, batchSize int) [][]T {
	if batchSize <= 0 {
		batchSize = 1 // Ensure minimum batch size of 1
	}

	numBatches := (len(items) + batchSize - 1) / batchSize
	batches := make([][]T, 0, numBatches)

	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batches = append(batches, items[i:end])
	}

	return batches
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
