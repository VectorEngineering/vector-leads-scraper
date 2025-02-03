package database

import (
	"context"
	"fmt"

	"gorm.io/gen/field"
)

// BatchDeleteLeads deletes multiple leads in batches.
// It processes leads in batches of batchSize to avoid overwhelming the database.
// If an error occurs during batch processing, it will be logged and the operation
// will continue with the next batch.
// BatchDeleteLeads deletes multiple leads by their IDs
func (db *Db) BatchDeleteLeads(ctx context.Context, ids []uint64, deletionType DeletionType) error {
	var (
		lQop = db.QueryOperator.LeadORM
	)

	if len(ids) == 0 {
		return ErrInvalidInput
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// check if the leads exist
	exists, err := lQop.WithContext(ctx).Where(lQop.Id.In(ids...)).Count()
	if err != nil {
		return fmt.Errorf("failed to check if leads exist: %w", err)
	}

	if exists == 0 {
		return ErrLeadDoesNotExist
	}

	// delete the leads
	queryRef := lQop.WithContext(ctx).Where(lQop.Id.In(ids...)).Select(field.AssociationFields)
	if deletionType == DeletionTypeHard {
		queryRef = queryRef.Unscoped()
	}

	result, err := queryRef.Delete()
	if err != nil {
		return fmt.Errorf("failed to delete leads: %w", err)
	}

	if result.RowsAffected != int64(len(ids)) {
		return ErrLeadDoesNotExist
	}

	return nil
} 