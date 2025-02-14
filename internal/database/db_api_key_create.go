package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

// CreateAPIKey creates a new API key in the database
func (db *Db) CreateAPIKey(ctx context.Context, workspaceId uint64, apiKey *lead_scraper_servicev1.APIKey) (*lead_scraper_servicev1.APIKey, error) {
	if apiKey == nil {
		return nil, ErrInvalidInput
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// validate the API key
	if err := apiKey.ValidateAll(); err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	// convert to ORM model
	apiKeyORM, err := apiKey.ToORM(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to ORM model: %w", err)
	}

	// get the workspace by id 
	wQop := db.QueryOperator.WorkspaceORM
	workspace, err := wQop.WithContext(ctx).Where(wQop.Id.Eq(workspaceId)).First()
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	if workspace == nil {
		return nil, ErrNotFound
	}

	// append the api key to the workspace
	if err := wQop.ApiKeys.Model(workspace).Append(&apiKeyORM); err != nil {
		return nil, fmt.Errorf("failed to append API key to workspace: %w", err)
	}

	// update the workspace 
	res, err := wQop.WithContext(ctx).Updates(workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}

	// if the number of rows affected is 0, then the workspace was not found
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}

	// convert back to protobuf
	pbResult, err := apiKeyORM.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to protobuf: %w", err)
	}

	return &pbResult, nil
}
