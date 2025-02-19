package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

// GetAPIKey retrieves an API key by ID
func (db *Db) GetAPIKey(ctx context.Context, id uint64) (*lead_scraper_servicev1.APIKey, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// Get the query operator
	apiKeyQop := db.QueryOperator.APIKeyORM

	// Get the API key
	apiKeyORM, err := apiKeyQop.WithContext(ctx).Where(apiKeyQop.Id.Eq(id)).First()
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	pbResult, err := apiKeyORM.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
	}

	return &pbResult, nil
}
