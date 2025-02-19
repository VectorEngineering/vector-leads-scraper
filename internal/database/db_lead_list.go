package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

// ListLeads retrieves a list of leads with pagination
func (db *Db) ListLeads(ctx context.Context, limit, offset int) ([]*lead_scraper_servicev1.Lead, error) {
	// validate the input
	if limit <= 0 {
		return nil, ErrInvalidInput
	}
	if offset < 0 {
		return nil, ErrInvalidInput
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// Get the query operator
	leadQop := db.QueryOperator.LeadORM

	// Get the leads
	leadsORM, err := leadQop.WithContext(ctx).
		Order(leadQop.Id.Desc()).
		Limit(limit).
		Offset(offset).
		Find()
	if err != nil {
		return nil, fmt.Errorf("failed to list leads: %w", err)
	}

	leads := make([]*lead_scraper_servicev1.Lead, 0, len(leadsORM))
	for _, leadORM := range leadsORM {
		lead, err := leadORM.ToPB(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
		}
		leads = append(leads, &lead)
	}

	return leads, nil
}
