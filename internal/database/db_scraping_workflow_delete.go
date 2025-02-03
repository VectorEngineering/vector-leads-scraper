package database

import (
	"context"
	"fmt"
)

// DeleteScrapingWorkflow deletes a scraping workflow by ID
func (db *Db) DeleteScrapingWorkflow(ctx context.Context, id uint64) error {
	var (
		swQop = db.QueryOperator.ScrapingWorkflowORM
	)

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if id == 0 {
		return ErrInvalidInput
	}

	// check and ensure the scraping workflow exists
	swORM, err := swQop.WithContext(ctx).Where(swQop.Id.Eq(id)).First()
	if err != nil {
		return fmt.Errorf("workflow does not exist: %w", err)
	}

	if swORM == nil {
		return fmt.Errorf("workflow does not exist: %w", ErrWorkflowDoesNotExist)
	}

	result, err := swQop.WithContext(ctx).Where(swQop.Id.Eq(id)).Delete()
	if err != nil {
		return fmt.Errorf("failed to delete scraping workflow: %w", err)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("scraping workflow not found: %w", ErrWorkflowDoesNotExist)
	}

	return nil
}

// BatchDeleteScrapingWorkflows deletes multiple scraping workflows by their IDs
func (db *Db) BatchDeleteScrapingWorkflows(ctx context.Context, ids []uint64) error {
	var (
		sQop = db.QueryOperator.ScrapingWorkflowORM
	)

	if len(ids) == 0 {
		return ErrInvalidInput
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// check if the workflows exist
	exists, err := sQop.WithContext(ctx).Where(sQop.Id.In(ids...)).Count()
	if err != nil {
		return fmt.Errorf("failed to check if workflows exist: %w", err)
	}

	if exists == 0 {
		return ErrWorkflowDoesNotExist
	}

	// delete the workflows
	result, err := sQop.WithContext(ctx).Where(sQop.Id.In(ids...)).Delete()
	if err != nil {
		return fmt.Errorf("failed to delete scraping workflows: %w", err)
	}

	if result.RowsAffected != int64(len(ids)) {
		return ErrWorkflowDoesNotExist
	}

	return nil
}
