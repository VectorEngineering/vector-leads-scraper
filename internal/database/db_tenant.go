package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/go-playground/validator/v10"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

// CreateTenantInput holds the input parameters for the CreateTenant function
type CreateTenantInput struct {
	Tenant         *lead_scraper_servicev1.Tenant `validate:"required"`
	OrganizationID uint64                          `validate:"required,gt=0"`
}

func (d *CreateTenantInput) validate() error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(d); err != nil {
		return multierr.Append(ErrInvalidInput, err)
	}
	return nil
}

// CreateTenant creates a new tenant in the database
func (db *Db) CreateTenant(ctx context.Context, input *CreateTenantInput) (*lead_scraper_servicev1.Tenant, error) {
	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if input == nil {
		return nil, ErrInvalidInput
	}

	if err := input.validate(); err != nil {
		return nil, err
	}

	tenant := input.Tenant
	// convert to orm
	tenantORM, err := tenant.ToORM(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tenant to orm: %w", err)
	}

	// Get the query operators
	orgQop := db.QueryOperator.OrganizationORM
	tenantQop := db.QueryOperator.TenantORM

	// Begin a transaction
	tx := db.Client.Engine.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// query for the organization and check if it exists
	org, err := orgQop.WithContext(ctx).Where(orgQop.Id.Eq(input.OrganizationID)).First()
	if err != nil {
		tx.Rollback()
		if err.Error() == "record not found" {
			return nil, ErrOrganizationDoesNotExist
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Set the organization ID on the tenant
	tenantORM.OrganizationId = &input.OrganizationID

	// Create the tenant
	if err := tenantQop.WithContext(ctx).Create(&tenantORM); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Append tenant to organization's tenants
	if err := orgQop.Tenants.WithContext(ctx).Model(org).Append(&tenantORM); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to append tenant to organization: %w", err)
	}

	// Update the organization
	if _, err := orgQop.WithContext(ctx).Where(orgQop.Id.Eq(input.OrganizationID)).Updates(org); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	// Get the created tenant with organization preloaded
	createdTenant, err := tenantQop.WithContext(ctx).
		Preload(tenantQop.Organization).
		Where(tenantQop.Id.Eq(tenantORM.Id)).
		First()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get created tenant: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Convert to protobuf
	tenantPb, err := createdTenant.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tenant to protobuf: %w", err)
	}

	return &tenantPb, nil
}

// GetTenantInput holds the input parameters for the GetTenant function
type GetTenantInput struct {
	ID uint64 `validate:"required,gt=0"`
}

func (d *GetTenantInput) validate() error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(d); err != nil {
		return multierr.Append(ErrInvalidInput, err)
	}
	return nil
}

// GetTenant retrieves a tenant from the database using the provided ID
func (db *Db) GetTenant(ctx context.Context, input *GetTenantInput) (*lead_scraper_servicev1.Tenant, error) {
	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if input == nil {
		return nil, ErrInvalidInput
	}

	if err := input.validate(); err != nil {
		return nil, err
	}

	// Get the query operator
	tenantQop := db.QueryOperator.TenantORM

	// Get the tenant with organization preloaded
	tenant, err := tenantQop.WithContext(ctx).
		Preload(tenantQop.Organization).
		Where(tenantQop.Id.Eq(input.ID)).
		First()
	if err != nil {
		if err.Error() == "record not found" {
			return nil, ErrTenantDoesNotExist
		}
		db.Logger.Error("failed to get tenant",
			zap.Error(err),
			zap.Uint64("tenant_id", input.ID))
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Convert to protobuf
	tenantPb, err := tenant.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tenant to protobuf: %w", err)
	}

	return &tenantPb, nil
}

// UpdateTenantInput holds the input parameters for the UpdateTenant function
type UpdateTenantInput struct {
	ID             uint64 `validate:"required,gt=0"`
	Tenant         *lead_scraper_servicev1.Tenant
}

func (d *UpdateTenantInput) validate() error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(d); err != nil {
		return multierr.Append(ErrInvalidInput, err)
	}

	if d.Tenant == nil {
		return ErrInvalidInput
	}

	if err := d.Tenant.ValidateAll(); err != nil {
		return multierr.Append(ErrInvalidInput, err)
	}

	return nil
}

// UpdateTenant updates an existing tenant in the database
func (db *Db) UpdateTenant(ctx context.Context, input *UpdateTenantInput) (*lead_scraper_servicev1.Tenant, error) {
	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if input == nil {
		return nil, ErrInvalidInput
	}

	if err := input.validate(); err != nil {
		return nil, err
	}

	// Get the query operator
	tenantQop := db.QueryOperator.TenantORM

	// Begin a transaction
	tx := db.Client.Engine.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the existing tenant
	existingTenant, err := tenantQop.WithContext(ctx).Where(tenantQop.Id.Eq(input.ID)).First()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	if existingTenant == nil {
		tx.Rollback()
		return nil, ErrTenantDoesNotExist
	}

	// Update tenant fields from input
	updatedTenant, err := input.Tenant.ToORM(ctx)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to convert tenant to ORM: %w", err)
	}

	// Preserve the organization ID
	updatedTenant.OrganizationId = existingTenant.OrganizationId

	// Update the tenant
	if _, err := tenantQop.WithContext(ctx).Where(tenantQop.Id.Eq(input.ID)).Updates(&updatedTenant); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	// Get the updated tenant
	updatedTenantRecord, err := tenantQop.WithContext(ctx).Where(tenantQop.Id.Eq(input.ID)).First()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get updated tenant: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Convert updated tenant to protobuf
	result, err := updatedTenantRecord.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tenant to protobuf: %w", err)
	}

	return &result, nil
}

// DeleteTenantInput holds the input parameters for the DeleteTenant function
type DeleteTenantInput struct {
	ID uint64 `validate:"required,gt=0"`
}

func (d *DeleteTenantInput) validate() error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(d); err != nil {
		return multierr.Append(ErrInvalidInput, err)
	}
	return nil
}

// DeleteTenant deletes a tenant from the database
func (db *Db) DeleteTenant(ctx context.Context, input *DeleteTenantInput) error {
	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if input == nil {
		return ErrInvalidInput
	}

	if err := input.validate(); err != nil {
		return err
	}

	// Get the query operator
	tenantQop := db.QueryOperator.TenantORM

	// First check if the tenant exists
	_, err := tenantQop.WithContext(ctx).Where(tenantQop.Id.Eq(input.ID)).First()
	if err != nil {
		if err.Error() == "record not found" {
			return ErrTenantDoesNotExist
		}
		db.Logger.Error("failed to get tenant",
			zap.Error(err),
			zap.Uint64("tenant_id", input.ID))
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	// Begin a transaction to ensure consistency
	tx := db.Client.Engine.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete the tenant
	if _, err := tenantQop.WithContext(ctx).Where(tenantQop.Id.Eq(input.ID)).Delete(); err != nil {
		tx.Rollback()
		db.Logger.Error("failed to delete tenant",
			zap.Error(err),
			zap.Uint64("tenant_id", input.ID))
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ListTenantsInput holds the input parameters for the ListTenants function
type ListTenantsInput struct {
	Limit          int    `validate:"required,gt=0"`
	Offset         int    `validate:"gte=0"`
	OrganizationID uint64 `validate:"omitempty,gt=0"`
}

func (d *ListTenantsInput) validate() error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(d); err != nil {
		return multierr.Append(ErrInvalidInput, err)
	}
	return nil
}

// ListTenants retrieves a paginated list of tenants
func (db *Db) ListTenants(ctx context.Context, input *ListTenantsInput) ([]*lead_scraper_servicev1.Tenant, error) {
	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if input == nil {
		return nil, ErrInvalidInput
	}

	if err := input.validate(); err != nil {
		return nil, err
	}

	// Get the query operator
	tenantQop := db.QueryOperator.TenantORM

	// Build the query with preloading
	query := tenantQop.WithContext(ctx).
		Select(tenantQop.ALL).
		Preload(tenantQop.Organization)

	// Add organization filter if specified
	if input.OrganizationID > 0 {
		query = query.Where(tenantQop.OrganizationId.Eq(input.OrganizationID))
	}

	// Execute the query with ordering, limit and offset
	tenants, err := query.
		Order(tenantQop.Id.Desc()).
		Limit(input.Limit).
		Offset(input.Offset).
		Find()
	if err != nil {
		db.Logger.Error("failed to list tenants", zap.Error(err))
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	// Convert to protobuf
	result := make([]*lead_scraper_servicev1.Tenant, len(tenants))
	for i, tenant := range tenants {
		pb, err := tenant.ToPB(ctx)
		if err != nil {
			db.Logger.Error("failed to convert tenant to protobuf",
				zap.Error(err),
				zap.Any("tenant", tenant))
			return nil, fmt.Errorf("failed to convert tenant to protobuf: %w", err)
		}
		result[i] = &pb
	}

	return result, nil
}
