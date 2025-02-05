package database

import (
	"context"
	"fmt"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

// UpdateAccount updates an existing account in the database
func (db *Db) UpdateAccount(ctx context.Context, orgId, tenantId uint64, account *lead_scraper_servicev1.Account) (*lead_scraper_servicev1.Account, error) {
	// validate the org and tenant id
	if orgId == 0 || tenantId == 0 {
		return nil, fmt.Errorf("%w: invalid org or tenant id", ErrInvalidInput)
	}

	if account == nil {
		return nil, fmt.Errorf("%w: account is nil", ErrInvalidInput)
	}

	// Validate email
	if account.Email == "" {
		return nil, fmt.Errorf("%w: account email is empty", ErrInvalidInput)
	}

	// check that the account exists
	existing, err := db.GetAccount(ctx, &GetAccountInput{
		ID: account.Id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("account not found")
	}

	acOrm := db.QueryOperator.AccountORM
	tenantOrm := db.QueryOperator.TenantORM


	// ensure the account exist based on the id in the database
	existingAcct, err := acOrm.WithContext(ctx).Where(
		acOrm.Id.Eq(account.Id),
	).First()
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}	
	if existingAcct == nil {
		return nil, fmt.Errorf("account not found")
	}

	// ensure the tenant exists
	tenant, err := tenantOrm.WithContext(ctx).Where(
		tenantOrm.OrganizationId.Eq(orgId),
		tenantOrm.Id.Eq(tenantId),
	).First()
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	if tenant == nil {
		return nil, fmt.Errorf("tenant not found")
	}

	// convert the account to an account orm
	updatedAcct, err := account.ToORM(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert account to orm: %w", err)
	}

	// update the account tied to the tenant
	if err := tenantOrm.Accounts.WithContext(ctx).Model(tenant).Replace(&updatedAcct); err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}


	// update the tenant
	res, err := tenantOrm.WithContext(ctx).Updates(tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	if res.RowsAffected == 0 {	
		return nil, fmt.Errorf("tenant not found")
	}

	return account, nil
}
