package database

import (
	"context"
	"fmt"

	"gorm.io/gen/field"
)

// DeleteLead deletes a lead by ID
func (db *Db) DeleteLead(ctx context.Context, id uint64, deletionType DeletionType) error {
	var (
		lQop = db.QueryOperator.LeadORM
	)

	if id == 0 {
		return ErrInvalidInput
	}

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	// First check if the lead exists
	_, err := lQop.WithContext(ctx).Where(lQop.Id.Eq(id)).First()
	if err != nil {
		if err.Error() == "record not found" {
			return ErrJobDoesNotExist
		}
		return fmt.Errorf("failed to get lead: %w", err)
	}

	queryRef := lQop.WithContext(ctx)
	if deletionType == DeletionTypeSoft {
		queryRef = queryRef.Where(lQop.Id.Eq(id)).Select(field.AssociationFields)
	} else {
		queryRef = queryRef.Where(lQop.Id.Eq(id)).Unscoped().Select(field.AssociationFields)
	}

	result, err := queryRef.Delete()
	if err != nil {
		return fmt.Errorf("failed to delete lead: %w", err)
	}

	if result.RowsAffected == 0 {
		return ErrJobDoesNotExist
	}

	return nil
}
