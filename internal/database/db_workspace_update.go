package database

import (
	"context"
	"errors"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UpdateWorkspace updates an existing workspace in the database
func (db *Db) UpdateWorkspace(ctx context.Context, workspace *lead_scraper_servicev1.Workspace) (*lead_scraper_servicev1.Workspace, error) {
	if workspace == nil {
		return nil, ErrInvalidInput
	}

	if workspace.Id == 0 {
		return nil, ErrInvalidInput
	}

	// query the workspace
	existing, err := db.GetWorkspace(ctx, workspace.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkspaceDoesNotExist
		}
		db.Logger.Error("failed to get workspace", zap.Error(err))
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// check if the workspace exists
	if existing == nil {
		return nil, ErrWorkspaceDoesNotExist
	}

	// convert the workspace to a gorm model
	gormWorkspace, err := workspace.ToORM(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert workspace to gorm model: %w", err)
	}

	wQop := db.QueryOperator.WorkspaceORM
	res, err := wQop.Where(wQop.Id.Eq(workspace.Id)).Updates(gormWorkspace)
	if err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}

	if res.Error != nil || res.RowsAffected == 0 {
		if res.Error != nil {
			return nil, fmt.Errorf("failed to update workspace: %w", res.Error)
		}

		return nil, ErrWorkspaceDoesNotExist
	}

	// convert the gorm workspace to a protobuf workspace
	updatedWorkspace, err := gormWorkspace.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert gorm workspace to protobuf workspace: %w", err)
	}

	return &updatedWorkspace, nil
}
