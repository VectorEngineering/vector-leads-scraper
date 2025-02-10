// Package database provides access and utility functions to interact with the database.
// This includes methods to create, read, update, and delete records in various tables.
package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/Vector/vector-leads-scraper/internal/constants"
	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/go-playground/validator/v10"
	"go.uber.org/multierr"
)

var ErrAccountAlreadyExists = errors.New("account already exists")

// CreateAccountInput holds the input parameters for the CreateAccount function.
type CreateAccountInput struct {
	OrgID    uint64                          `validate:"required"`
	TenantID uint64                          `validate:"required"`
	Account  *lead_scraper_servicev1.Account `validate:"required"`
}

func (d *CreateAccountInput) validate() error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(d); err != nil {
		return multierr.Append(ErrInvalidInput, err)
	}

	if d.Account == nil {
		return multierr.Append(ErrInvalidInput, errors.New("account cannot be nil"))
	}

	// Add any additional account-specific validation here
	if d.Account.Email == "" {
		return multierr.Append(ErrInvalidInput, errors.New("email is required"))
	}

	return nil
}

// CreateAccount creates a new account in the database with the specified organization and tenant IDs.
// It validates the input parameters, checks for duplicate accounts, and creates the account record
// with proper associations.
//
// The function performs the following steps:
// 1. Validates input parameters and account data
// 2. Checks for existing accounts with the same email
// 3. Creates the account record with organization and tenant associations
// 4. Sets up default account settings and status
//
// Parameters:
//   - ctx: A context.Context for timeout and tracing control
//   - input: A CreateAccountInput struct containing:
//   - OrgID: Organization ID that owns the account
//   - TenantID: Tenant ID to associate the account with
//   - Account: Account details following the lead scraper service schema
//
// Returns:
//   - *lead_scraper_servicev1.Account: A pointer to the created Account object
//   - error: An error if the operation fails, or nil if successful
//
// Errors:
//   - Returns ErrInvalidInput if input validation fails
//   - Returns ErrAccountAlreadyExists if an account with the same email exists
//   - Returns error if database operations fail
//
// Example usage:
//
//	db := database.NewDb()
//	input := &database.CreateAccountInput{
//	    OrgID: "org123",
//	    TenantID: "tenant456",
//	    Account: &lead_scraper_servicev1.Account{
//	        Email: "user@example.com",
//	        // ... other account fields
//	    },
//	}
//
//	account, err := db.CreateAccount(ctx, input)
//	if err != nil {
//	    log.Printf("Failed to create account: %v", err)
//	    return err
//	}
func (db *Db) CreateAccount(ctx context.Context, input *CreateAccountInput) (*lead_scraper_servicev1.Account, error) {
	// ensure the db operation executes within the specified timeout
	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if input == nil {
		return nil, ErrInvalidInput
	}

	// validate the input parameters
	if err := input.validate(); err != nil {
		return nil, err
	}

	// make sure the org and tenant exist
	tenantQop := db.QueryOperator.TenantORM
	tenantOrm, err := tenantQop.WithContext(ctx).Where(
		tenantQop.Id.Eq(input.TenantID),
		tenantQop.OrganizationId.Eq(input.OrgID),
	).First()
	if err != nil {
		return nil, err
	}

	if tenantOrm == nil {
		return nil, ErrTenantDoesNotExist
	}

	// Check if account with same email already exists
	existing, err := db.GetAccountByEmail(ctx, input.Account.Email)
	if err == nil && existing != nil {
		return nil, ErrAccountAlreadyExists
	}

	// Set default account status if not specified
	if input.Account.AccountStatus == lead_scraper_servicev1.Account_ACCOUNT_STATUS_UNSPECIFIED {
		input.Account.AccountStatus = lead_scraper_servicev1.Account_ACCOUNT_STATUS_ACTIVE
	}

	// convert account to orm
	accountORM, err := input.Account.ToORM(ctx)
	if err != nil {
		return nil, err
	}

	if err := tenantQop.Accounts.WithContext(ctx).Model(tenantOrm).Append(&accountORM); err != nil {
		return nil, err
	}

	// update the tenant with the new account
	res, err := tenantQop.WithContext(ctx).Updates(tenantOrm)
	if err != nil {
		return nil, err
	}

	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("failed to update tenant")
	}

	// convert orm to pb
	acct, err := accountORM.ToPB(ctx)
	if err != nil {
		return nil, err
	}

	return &acct, nil
}

// GetAccountByEmail retrieves an account by email address
func (db *Db) GetAccountByEmail(ctx context.Context, accountEmail string) (*lead_scraper_servicev1.AccountORM, error) {
	var account *lead_scraper_servicev1.AccountORM

	ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
	defer cancel()

	if accountEmail == constants.EMPTY {
		return nil, errors.New("invalid input parameters. account email cannot be empty")
	}

	u := db.QueryOperator.AccountORM
	queryRef := u.WithContext(ctx)

	queryRef = queryRef.Where(u.Email.Eq(accountEmail))

	account, err := db.PreloadAccount(queryRef)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToGetAccountByEmail, err)
	}

	return account, nil
}
