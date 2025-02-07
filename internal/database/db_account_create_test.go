// Package database provides access and utility functions to interact with the database.
// This includes methods to create, read, update, and delete records in various tables.
package database

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type accountTestContext struct {
	Organization *lead_scraper_servicev1.Organization
	Tenant       *lead_scraper_servicev1.Tenant
	Account      *lead_scraper_servicev1.Account
	Workspace    *lead_scraper_servicev1.Workspace
	Cleanup      func()
}

func setupAccountTestContext(t *testing.T) (*accountTestContext) {
	ctx := context.Background()

	// Create test organization
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Organization: testutils.GenerateRandomizedOrganization(),
	})

	tenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
		Tenant: testutils.GenerateRandomizedTenant(),
		OrganizationID: org.Id,
	})
	require.NoError(t, err)
	
	cleanup := func() {	
		conn.DeleteOrganization(ctx, &DeleteOrganizationInput{ID: org.Id})
		conn.DeleteTenant(ctx, &DeleteTenantInput{ID: tenant.Id})
	}

	account, err := conn.CreateAccount(ctx, &CreateAccountInput{
		Account: testutils.GenerateRandomizedAccount(),
		OrgID:   org.Id,
		TenantID: tenant.Id,
	})
	require.NoError(t, err)

	// create a workspace
	workspace, err := conn.CreateWorkspace(ctx, &CreateWorkspaceInput{
		Workspace: testutils.GenerateRandomWorkspace(),
		AccountID: account.Id,
		OrganizationID: org.Id,
		TenantID: tenant.Id,
	})
	require.NoError(t, err)	

	return &accountTestContext{
		Organization: org,
		Tenant:       tenant,
		Account:      account,
		Workspace:    workspace,
		Cleanup:      cleanup,
	}
}

func TestCreateAccountInput_validate(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	tests := []struct {
		name    string
		d       *CreateAccountInput
		wantErr bool
	}{
		{
			name: "success - valid input",
			d: &CreateAccountInput{
				Account:  testutils.GenerateRandomizedAccount(),
				OrgID:    tc.Organization.Id,
				TenantID: tc.Tenant.Id,
			},
			wantErr: false,
		},
		{
			name:    "failure - nil input",
			d:       nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.validate(); (err != nil) != tt.wantErr {
				t.Errorf("CreateAccountInput.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDb_CreateAccount(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	type args struct {
		ctx   context.Context
		input *CreateAccountInput
	}
	tests := []struct {
		name    string
		args    args
		want    *lead_scraper_servicev1.AccountORM
		wantErr bool
		errType error
	}{
		{
			name: "success - create valid account",
			args: args{
				ctx: context.Background(),
				input: &CreateAccountInput{
					Account:  testutils.GenerateRandomizedAccount(),
					OrgID:    tc.Organization.Id,
					TenantID: tc.Tenant.Id,
				},
			},
			wantErr: false,
		},
		{
			name: "error - nil input",
			args: args{
				ctx:   context.Background(),
				input: nil,
			},
			wantErr: true,
			errType: ErrInvalidInput,
		},
		{
			name: "error - nil account",
			args: args{
				ctx: context.Background(),
				input: &CreateAccountInput{
					Account:  nil,
					OrgID:    tc.Organization.Id,
					TenantID: tc.Tenant.Id,
				},
			},
			wantErr: true,
			errType: ErrInvalidInput,
		},
		{
			name: "error - empty org ID",
			args: args{
				ctx: context.Background(),
				input: &CreateAccountInput{
					Account:  testutils.GenerateRandomizedAccount(),
					OrgID:    0,
					TenantID: tc.Tenant.Id,
				},
			},
			wantErr: true,
			errType: ErrInvalidInput,
		},
		{
			name: "error - empty tenant ID",
			args: args{
				ctx: context.Background(),
				input: &CreateAccountInput{
					Account:  testutils.GenerateRandomizedAccount(),
					OrgID:    tc.Organization.Id,
					TenantID: 0,
				},
			},
			wantErr: true,
			errType: ErrInvalidInput,
		},
		{
			name: "error - context timeout",
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
					defer cancel()
					time.Sleep(2 * time.Millisecond)
					return ctx
				}(),
				input: &CreateAccountInput{
					Account:  testutils.GenerateRandomizedAccount(),
					OrgID:    tc.Organization.Id,
					TenantID: tc.Tenant.Id,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conn.CreateAccount(tt.args.ctx, tt.args.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.NotEmpty(t, got.Id)
			assert.Equal(t, tt.args.input.Account.Email, got.Email)

			// Clean up created account
			err = conn.DeleteAccount(context.Background(), &DeleteAccountParams{
				ID:           got.Id,
				DeletionType: DeletionTypeSoft,
			})
			require.NoError(t, err)
		})
	}
}

func TestDb_GetAccountByEmail(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create a test account
	validAccount := testutils.GenerateRandomizedAccount()
	createdAccount, err := conn.CreateAccount(context.Background(), &CreateAccountInput{
		Account:  validAccount,
		OrgID:    tc.Organization.Id,
		TenantID: tc.Tenant.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createdAccount)

	// Clean up after test
	defer func() {
		if createdAccount != nil {
			err := conn.DeleteAccount(context.Background(), &DeleteAccountParams{
				ID:           createdAccount.Id,
				DeletionType: DeletionTypeSoft,
			})
			require.NoError(t, err)
		}
	}()

	type args struct {
		ctx          context.Context
		accountEmail string
	}
	tests := []struct {
		name    string
		args    args
		want    *lead_scraper_servicev1.AccountORM
		wantErr bool
		errType error
	}{
		{
			name: "success - get existing account",
			args: args{
				ctx:          context.Background(),
				accountEmail: validAccount.Email,
			},
			wantErr: false,
		},
		{
			name: "error - empty email",
			args: args{
				ctx:          context.Background(),
				accountEmail: "",
			},
			wantErr: true,
			errType: ErrInvalidInput,
		},
		{
			name: "error - non-existent email",
			args: args{
				ctx:          context.Background(),
				accountEmail: "nonexistent@example.com",
			},
			wantErr: true,
			errType: ErrFailedToGetAccountByEmail,
		},
		{
			name: "error - context timeout",
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
					defer cancel()
					time.Sleep(2 * time.Millisecond)
					return ctx
				}(),
				accountEmail: validAccount.Email,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conn.GetAccountByEmail(tt.args.ctx, tt.args.accountEmail)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorContains(t, err, tt.errType.Error())
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tt.args.accountEmail, got.Email)
		})
	}
}

func TestDb_CreateAccount_DuplicateEmail(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create initial account
	validAccount := testutils.GenerateRandomizedAccount()
	validAccount.AccountStatus = lead_scraper_servicev1.Account_ACCOUNT_STATUS_ACTIVE

	createdAccount, err := conn.CreateAccount(context.Background(), &CreateAccountInput{
		Account:  validAccount,
		OrgID:    tc.Organization.Id,
		TenantID: tc.Tenant.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createdAccount)

	// Clean up after test
	defer func() {
		if createdAccount != nil {
			err := conn.DeleteAccount(context.Background(), &DeleteAccountParams{
				ID:           createdAccount.Id,
				DeletionType: DeletionTypeSoft,
			})
			if err != nil {
				t.Logf("Failed to cleanup test account: %v", err)
			}
		}
	}()

	// Try to create account with same email
	duplicateAccount := testutils.GenerateRandomizedAccount()
	duplicateAccount.Email = validAccount.Email
	duplicateAccount.AccountStatus = lead_scraper_servicev1.Account_ACCOUNT_STATUS_ACTIVE

	_, err = conn.CreateAccount(context.Background(), &CreateAccountInput{
		Account:  duplicateAccount,
		OrgID:    tc.Organization.Id,
		TenantID: tc.Tenant.Id,
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrAccountAlreadyExists)
}

func TestDb_CreateAccount_ConcurrentCreation(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	numAccounts := 5
	var wg sync.WaitGroup
	errors := make(chan error, numAccounts)
	accounts := make(chan *lead_scraper_servicev1.Account, numAccounts)

	// Create accounts concurrently
	for i := 0; i < numAccounts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mockAccount := testutils.GenerateRandomizedAccount()
			mockAccount.AccountStatus = lead_scraper_servicev1.Account_ACCOUNT_STATUS_ACTIVE

			account, err := conn.CreateAccount(context.Background(), &CreateAccountInput{
				Account:  mockAccount,
				OrgID:    tc.Organization.Id,
				TenantID: tc.Tenant.Id,
			})
			if err != nil {
				errors <- err
				return
			}
			accounts <- account
		}()
	}

	wg.Wait()
	close(errors)
	close(accounts)

	// Clean up created accounts
	createdAccounts := make([]*lead_scraper_servicev1.Account, 0)
	for account := range accounts {
		createdAccounts = append(createdAccounts, account)
	}

	// Clean up in a deferred function to ensure cleanup happens even if test fails
	defer func() {
		for _, account := range createdAccounts {
			if account != nil {
				err := conn.DeleteAccount(context.Background(), &DeleteAccountParams{
					ID:           account.Id,
					DeletionType: DeletionTypeSoft,
				})
				if err != nil {
					t.Logf("Failed to cleanup test account: %v", err)
				}
			}
		}
	}()

	// Check for errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}
	require.Empty(t, errs, "Expected no errors during concurrent creation, got: %v", errs)

	// Verify all accounts were created successfully
	require.Equal(t, numAccounts, len(createdAccounts), "Expected %d accounts to be created, got %d", numAccounts, len(createdAccounts))
	for _, account := range createdAccounts {
		require.NotNil(t, account)
		require.NotZero(t, account.Id)
		require.Equal(t, lead_scraper_servicev1.Account_ACCOUNT_STATUS_ACTIVE, account.AccountStatus)
	}
}

func TestCreateAccountInput_Validate(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	validAccount := testutils.GenerateRandomizedAccount()

	tests := []struct {
		name    string
		input   *CreateAccountInput
		wantErr bool
	}{
		{
			name: "success - valid input",
			input: &CreateAccountInput{
				Account:  validAccount,
				OrgID:    tc.Organization.Id,
				TenantID: tc.Tenant.Id,
			},
			wantErr: false,
		},
		{
			name:    "failure - nil input",
			input:   nil,
			wantErr: true,
		},
		{
			name: "failure - nil account",
			input: &CreateAccountInput{
				Account:  nil,
				OrgID:    tc.Organization.Id,
				TenantID: tc.Tenant.Id,
			},
			wantErr: true,
		},
		{
			name: "failure - empty email",
			input: &CreateAccountInput{
				Account: &lead_scraper_servicev1.Account{
					Email: "",
				},
				OrgID:    tc.Organization.Id,
				TenantID: tc.Tenant.Id,
			},
			wantErr: true,
		},
		{
			name: "failure - empty org ID",
			input: &CreateAccountInput{
				Account:  validAccount,
				OrgID:    0,
				TenantID: tc.Tenant.Id,
			},
			wantErr: true,
		},
		{
			name: "failure - empty tenant ID",
			input: &CreateAccountInput{
				Account:  validAccount,
				OrgID:    tc.Organization.Id,
				TenantID: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
