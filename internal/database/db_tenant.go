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
	Name           string `validate:"required"`
	OrganizationID uint64 `validate:"required,gt=0"`
	Description    string
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

	tenant := &lead_scraper_servicev1.TenantORM{
		Name:           input.Name,
		OrganizationId: &input.OrganizationID,
		Description:    input.Description,
	}

	if err := db.Client.Engine.WithContext(ctx).Create(tenant).Error; err != nil {
		db.Logger.Error("failed to create tenant",
			zap.Error(err),
			zap.String("name", input.Name))
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Convert to protobuf
	tenantPb, err := tenant.ToPB(ctx)
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

	var tenant *lead_scraper_servicev1.TenantORM
	if err := db.Client.Engine.WithContext(ctx).First(&tenant, input.ID).Error; err != nil {
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
	Name           string `validate:"required"`
	OrganizationID uint64 `validate:"required,gt=0"`
	Description    string
}

func (d *UpdateTenantInput) validate() error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(d); err != nil {
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

	tenant := &lead_scraper_servicev1.TenantORM{
		Id:             input.ID,
		Name:           input.Name,
		OrganizationId: &input.OrganizationID,
		Description:    input.Description,
	}

	if err := db.Client.Engine.WithContext(ctx).Save(tenant).Error; err != nil {
		db.Logger.Error("failed to update tenant",
			zap.Error(err),
			zap.Uint64("tenant_id", input.ID))
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	// Convert to protobuf
	tenantPb, err := tenant.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tenant to protobuf: %w", err)
	}

	return &tenantPb, nil
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

	if err := db.Client.Engine.WithContext(ctx).Delete(&lead_scraper_servicev1.TenantORM{}, input.ID).Error; err != nil {
		db.Logger.Error("failed to delete tenant",
			zap.Error(err),
			zap.Uint64("tenant_id", input.ID))
		return fmt.Errorf("failed to delete tenant: %w", err)
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

	query := db.Client.Engine.WithContext(ctx)

	if input.OrganizationID > 0 {
		query = query.Where("organization_id = ?", input.OrganizationID)
	}

	var tenants []*lead_scraper_servicev1.TenantORM
	if err := query.
		Order("id desc").
		Limit(input.Limit).
		Offset(input.Offset).
		Find(&tenants).Error; err != nil {
		db.Logger.Error("failed to list tenants", zap.Error(err))
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	// Convert to protobuf
	result := make([]*lead_scraper_servicev1.Tenant, len(tenants))
	for i, tenant := range tenants {
		pb, err := tenant.ToPB(ctx)
		if err != nil {
			db.Logger.Error("failed to convert tenant to protobuf",
				zap.Error(err))
			return nil, fmt.Errorf("failed to convert tenant to protobuf: %w", err)
		}
		result[i] = &pb
	}

	return result, nil
} 