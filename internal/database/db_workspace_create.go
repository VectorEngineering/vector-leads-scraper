package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/go-playground/validator/v10"
	"go.uber.org/multierr"
)

// CreateWorkspaceInput holds the input parameters for the CreateWorkspace function
type CreateWorkspaceInput struct {
	Workspace      *lead_scraper_servicev1.Workspace `validate:"required"`
	AccountID      uint64                            `validate:"required,gt=0"`
	TenantID       uint64                            `validate:"required,gt=0"`
	OrganizationID uint64                            `validate:"required,gt=0"`
}

func (d *CreateWorkspaceInput) validate() error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(d); err != nil {
		return multierr.Append(ErrInvalidInput, err)
	}

	if d.Workspace == nil {
		return ErrInvalidInput
	}

	if err := d.Workspace.ValidateAll(); err != nil {
		return multierr.Append(ErrInvalidInput, err)
	}

	return nil
}

// CreateWorkspace creates a new workspace in the database
func (db *Db) CreateWorkspace(ctx context.Context, input *CreateWorkspaceInput) (*lead_scraper_servicev1.Workspace, error) {
	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if input == nil {
		return nil, ErrInvalidInput
	}

	if err := input.validate(); err != nil {
		return nil, err
	}

	// Get the query operators
	orgQop := db.QueryOperator.OrganizationORM
	tenantQop := db.QueryOperator.TenantORM
	accountQop := db.QueryOperator.AccountORM
	workspaceQop := db.QueryOperator.WorkspaceORM

	// Check if organization exists
	_, err := orgQop.WithContext(ctx).Where(orgQop.Id.Eq(input.OrganizationID)).First()
	if err != nil {
		if err.Error() == "record not found" {
			return nil, ErrOrganizationDoesNotExist
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Check if tenant exists and belongs to the organization
	tenant, err := tenantQop.WithContext(ctx).Where(tenantQop.Id.Eq(input.TenantID)).First()
	if err != nil {
		if err.Error() == "record not found" {
			return nil, ErrTenantDoesNotExist
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Verify tenant belongs to the organization
	if tenant.OrganizationId == nil || *tenant.OrganizationId != input.OrganizationID {
		return nil, fmt.Errorf("tenant does not belong to the specified organization")
	}

	// Check if account exists and belongs to the tenant
	account, err := accountQop.WithContext(ctx).Where(accountQop.Id.Eq(input.AccountID)).First()
	if err != nil {
		if err.Error() == "record not found" {
			return nil, ErrAccountDoesNotExist
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Verify account belongs to the tenant
	if account.TenantId == nil || *account.TenantId != input.TenantID {
		return nil, fmt.Errorf("account does not belong to the specified tenant")
	}

	// Convert workspace to ORM
	workspaceORM, err := input.Workspace.ToORM(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert workspace to ORM: %w", err)
	}

	// Create the workspace
	if err := workspaceQop.WithContext(ctx).Create(&workspaceORM); err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// Append workspace to account's workspaces
	if err := accountQop.Workspaces.WithContext(ctx).Model(account).Append(&workspaceORM); err != nil {
		return nil, fmt.Errorf("failed to append workspace to account: %w", err)
	}

	// Update the account
	if _, err := accountQop.WithContext(ctx).Where(accountQop.Id.Eq(input.AccountID)).Updates(account); err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	// Get the created workspace with all relationships preloaded
	createdWorkspace, err := workspaceQop.WithContext(ctx).
		Preload(workspaceQop.ApiKeys).
		Preload(workspaceQop.Webhooks).
		Where(workspaceQop.Id.Eq(workspaceORM.Id)).
		First()
	if err != nil {
		return nil, fmt.Errorf("failed to get created workspace: %w", err)
	}
	
	// Convert to protobuf
	workspacePb, err := createdWorkspace.ToPB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert workspace to protobuf: %w", err)
	}

	return &workspacePb, nil
}
