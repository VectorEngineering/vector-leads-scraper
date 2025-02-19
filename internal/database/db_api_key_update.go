package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

// UpdateAPIKey updates an existing API key
func (db *Db) UpdateAPIKey(ctx context.Context, apiKey *lead_scraper_servicev1.APIKey) (*lead_scraper_servicev1.APIKey, error) {
	if apiKey == nil || apiKey.Id == 0 {
		return nil, ErrInvalidInput
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// convert to ORM model
	apiKeyORM, err := apiKey.ToORM(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to ORM model: %w", err)
	}

	// Get the query operator
	apiKeyQop := db.QueryOperator.APIKeyORM

	// update the API key
	res, err := apiKeyQop.WithContext(ctx).Where(apiKeyQop.Id.Eq(apiKey.Id)).Updates(&apiKeyORM)
	if err != nil {
		return nil, fmt.Errorf("failed to update API key: %w", err)
	}

	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("API key not found")
	}

	pbResult, err := apiKeyORM.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
	}

	return &pbResult, nil
}
