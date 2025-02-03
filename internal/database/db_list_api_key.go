package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

// ListAPIKeys retrieves a list of API keys with pagination
func (db *Db) ListAPIKeys(ctx context.Context, limit, offset int) ([]*lead_scraper_servicev1.APIKey, error) {
	// validate the input
	if limit <= 0 {
		return nil, ErrInvalidInput
	}
	if offset < 0 {
		return nil, ErrInvalidInput
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	var apiKeysORM []lead_scraper_servicev1.APIKeyORM
	result := db.Client.Engine.WithContext(ctx).
		Order("id asc").
		Limit(limit).
		Offset(offset).
		Find(&apiKeysORM)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", result.Error)
	}

	apiKeys := make([]*lead_scraper_servicev1.APIKey, 0, len(apiKeysORM))
	for _, apiKeyORM := range apiKeysORM {
		apiKey, err := apiKeyORM.ToPB(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
		}
		apiKeys = append(apiKeys, &apiKey)
	}

	return apiKeys, nil
}
