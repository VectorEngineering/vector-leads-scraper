package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/go-playground/validator/v10"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"gorm.io/gen/field"
)

// CreateOrganizationInput holds the input parameters for the CreateOrganization function
type CreateOrganizationInput struct {
	Organization *lead_scraper_servicev1.Organization
}

func (d *CreateOrganizationInput) validate() error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(d); err != nil {
		return multierr.Append(ErrInvalidInput, err)
	}

	if d.Organization == nil {
		return ErrInvalidInput
	}

	// validate the organization
	if err := d.Organization.Validate(); err != nil {
		return err
	}

	return nil
}

// CreateOrganization creates a new organization in the database
func (db *Db) CreateOrganization(ctx context.Context, input *CreateOrganizationInput) (*lead_scraper_servicev1.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if input == nil {
		return nil, ErrInvalidInput
	}

	if err := input.validate(); err != nil {
		return nil, err
	}

	org := input.Organization
	// Convert to ORM
	orgOrm, err := org.ToORM(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert organization to ORM: %w", err)
	}

	orgQop := db.QueryOperator.OrganizationORM
	if err := orgQop.WithContext(ctx).Create(&orgOrm); err != nil {
		db.Logger.Error("failed to create organization",
			zap.Error(err),
			zap.String("name", org.GetName()))
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return org, nil
}

// GetOrganizationInput holds the input parameters for the GetOrganization function
type GetOrganizationInput struct {
	ID uint64 `validate:"required,gt=0"`
}

func (d *GetOrganizationInput) validate() error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(d); err != nil {
		return multierr.Append(ErrInvalidInput, err)
	}
	return nil
}

// GetOrganization retrieves an organization from the database using the provided ID
func (db *Db) GetOrganization(ctx context.Context, input *GetOrganizationInput) (*lead_scraper_servicev1.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if input == nil {
		return nil, ErrInvalidInput
	}

	if err := input.validate(); err != nil {
		return nil, err
	}

	// Get the query operator
	orgQop := db.QueryOperator.OrganizationORM

	// Get the organization
	org, err := orgQop.WithContext(ctx).Where(orgQop.Id.Eq(input.ID)).First()
	if err != nil {
		if err.Error() == "record not found" {
			return nil, ErrOrganizationDoesNotExist
		}
		db.Logger.Error("failed to get organization",
			zap.Error(err),
			zap.Uint64("organization_id", input.ID))
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Convert to protobuf
	orgPb, err := org.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert organization to protobuf: %w", err)
	}

	return &orgPb, nil
}

// UpdateOrganizationInput holds the input parameters for the UpdateOrganization function
type UpdateOrganizationInput struct {
	ID          uint64 `validate:"required,gt=0"`
	Name        string `validate:"required"`
	Description string
}

func (d *UpdateOrganizationInput) validate() error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(d); err != nil {
		return multierr.Append(ErrInvalidInput, err)
	}
	return nil
}

// UpdateOrganization updates an existing organization in the database
func (db *Db) UpdateOrganization(ctx context.Context, input *UpdateOrganizationInput) (*lead_scraper_servicev1.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if input == nil {
		return nil, ErrInvalidInput
	}

	if err := input.validate(); err != nil {
		return nil, err
	}

	// Get the query operator
	orgQop := db.QueryOperator.OrganizationORM

	// Create the update model
	org := &lead_scraper_servicev1.OrganizationORM{
		Id:          input.ID,
		Name:        input.Name,
		Description: input.Description,
	}

	// Update the organization
	if _, err := orgQop.WithContext(ctx).Where(orgQop.Id.Eq(input.ID)).Updates(org); err != nil {
		db.Logger.Error("failed to update organization",
			zap.Error(err),
			zap.Uint64("organization_id", input.ID))
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	// Get the updated organization
	updatedOrg, err := orgQop.WithContext(ctx).Where(orgQop.Id.Eq(input.ID)).First()
	if err != nil {
		return nil, fmt.Errorf("failed to get updated organization: %w", err)
	}

	// Convert to protobuf
	orgPb, err := updatedOrg.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert organization to protobuf: %w", err)
	}

	return &orgPb, nil
}

// DeleteOrganizationInput holds the input parameters for the DeleteOrganization function
type DeleteOrganizationInput struct {
	ID uint64 `validate:"required,gt=0"`
}

func (d *DeleteOrganizationInput) validate() error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(d); err != nil {
		return multierr.Append(ErrInvalidInput, err)
	}
	return nil
}

// DeleteOrganization deletes an organization from the database
func (db *Db) DeleteOrganization(ctx context.Context, input *DeleteOrganizationInput) error {
	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if input == nil {
		return ErrInvalidInput
	}

	if err := input.validate(); err != nil {
		return err
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

	// Get the query operator
	orgQop := db.QueryOperator.OrganizationORM

	// First check if the organization exists
	_, err := orgQop.WithContext(ctx).Where(orgQop.Id.Eq(input.ID)).First()
	if err != nil {
		tx.Rollback()
		if err.Error() == "record not found" {
			return ErrOrganizationDoesNotExist
		}
		db.Logger.Error("failed to get organization",
			zap.Error(err),
			zap.Uint64("organization_id", input.ID))
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Delete all associated tenants first
	tenantQop := db.QueryOperator.TenantORM
	if _, err := tenantQop.WithContext(ctx).Where(tenantQop.OrganizationId.Eq(input.ID)).Select(field.AssociationFields).Delete(); err != nil {
		tx.Rollback()
		db.Logger.Error("failed to delete organization's tenants",
			zap.Error(err),
			zap.Uint64("organization_id", input.ID))
		return fmt.Errorf("failed to delete organization's tenants: %w", err)
	}

	// Delete the organization
	if _, err := orgQop.WithContext(ctx).Where(orgQop.Id.Eq(input.ID)).Select(field.AssociationFields).Delete(); err != nil {
		tx.Rollback()
		db.Logger.Error("failed to delete organization",
			zap.Error(err),
			zap.Uint64("organization_id", input.ID))
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ListOrganizationsInput holds the input parameters for the ListOrganizations function
type ListOrganizationsInput struct {
	Limit  int `validate:"required,gt=0"`
	Offset int `validate:"gte=0"`
}

func (d *ListOrganizationsInput) validate() error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(d); err != nil {
		return multierr.Append(ErrInvalidInput, err)
	}
	return nil
}

// ListOrganizations retrieves a paginated list of organizations
func (db *Db) ListOrganizations(ctx context.Context, input *ListOrganizationsInput) ([]*lead_scraper_servicev1.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if input == nil {
		return nil, ErrInvalidInput
	}

	if err := input.validate(); err != nil {
		return nil, err
	}

	// Get the query operator
	orgQop := db.QueryOperator.OrganizationORM

	// Get the organizations
	orgs, err := orgQop.WithContext(ctx).
		Order(orgQop.Id.Desc()).
		Limit(input.Limit).
		Offset(input.Offset).
		Find()
	if err != nil {
		db.Logger.Error("failed to list organizations", zap.Error(err))
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	// Convert to protobuf
	result := make([]*lead_scraper_servicev1.Organization, len(orgs))
	for i, org := range orgs {
		pb, err := org.ToPB(ctx)
		if err != nil {
			db.Logger.Error("failed to convert organization to protobuf",
				zap.Error(err))
			return nil, fmt.Errorf("failed to convert organization to protobuf: %w", err)
		}
		result[i] = &pb
	}

	return result, nil
}
