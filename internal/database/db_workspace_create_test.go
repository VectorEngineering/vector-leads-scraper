package database

import (
	"context"
	"sync"
	"testing"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateWorkspace(t *testing.T) {
	ctx := context.Background()

	// Create test organization
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Organization: testutils.GenerateRandomizedOrganization(),
	})
	require.NoError(t, err)

	// Create test tenant
	tenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
		Tenant:         testutils.GenerateRandomizedTenant(),
		OrganizationID: org.Id,
	})
	require.NoError(t, err)

	// Create test account
	account, err := conn.CreateAccount(ctx, &CreateAccountInput{
		Account:        testutils.GenerateRandomizedAccount(),
		TenantID:      tenant.Id,
		OrgID: org.Id,
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		input   *CreateWorkspaceInput
		setup   func() *CreateWorkspaceInput
		wantErr bool
		errType error
	}{
		{
			name:  "success - create workspace with valid input",
			setup: func() *CreateWorkspaceInput {
				return &CreateWorkspaceInput{
					Workspace:      testutils.GenerateRandomWorkspace(),
					AccountID:      account.Id,
					TenantID:       tenant.Id,
					OrganizationID: org.Id,
				}
			},
			wantErr: false,
		},
		{
			name:    "error - nil input",
			input:   nil,
			wantErr: true,
			errType: ErrInvalidInput,
		},
		{
			name: "error - nil workspace",
			setup: func() *CreateWorkspaceInput {
				return &CreateWorkspaceInput{
					Workspace:      nil,
					AccountID:      account.Id,
					TenantID:       tenant.Id,
					OrganizationID: org.Id,
				}
			},
			wantErr: true,
			errType: ErrInvalidInput,
		},
		{
			name: "error - invalid account ID",
			setup: func() *CreateWorkspaceInput {
				return &CreateWorkspaceInput{
					Workspace:      testutils.GenerateRandomWorkspace(),
					AccountID:      0,
					TenantID:       tenant.Id,
					OrganizationID: org.Id,
				}
			},
			wantErr: true,
			errType: ErrInvalidInput,
		},
		{
			name: "error - invalid tenant ID",
			setup: func() *CreateWorkspaceInput {
				return &CreateWorkspaceInput{
					Workspace:      testutils.GenerateRandomWorkspace(),
					AccountID:      account.Id,
					TenantID:       0,
					OrganizationID: org.Id,
				}
			},
			wantErr: true,
			errType: ErrInvalidInput,
		},
		{
			name: "error - invalid organization ID",
			setup: func() *CreateWorkspaceInput {
				return &CreateWorkspaceInput{
					Workspace:      testutils.GenerateRandomWorkspace(),
					AccountID:      account.Id,
					TenantID:       tenant.Id,
					OrganizationID: 0,
				}
			},
			wantErr: true,
			errType: ErrInvalidInput,
		},
		{
			name: "error - non-existent organization",
			setup: func() *CreateWorkspaceInput {
				return &CreateWorkspaceInput{
					Workspace:      testutils.GenerateRandomWorkspace(),
					AccountID:      account.Id,
					TenantID:       tenant.Id,
					OrganizationID: 999999,
				}
			},
			wantErr: true,
			errType: ErrOrganizationDoesNotExist,
		},
		{
			name: "error - non-existent tenant",
			setup: func() *CreateWorkspaceInput {
				return &CreateWorkspaceInput{
					Workspace:      testutils.GenerateRandomWorkspace(),
					AccountID:      account.Id,
					TenantID:       999999,
					OrganizationID: org.Id,
				}
			},
			wantErr: true,
			errType: ErrTenantDoesNotExist,
		},
		{
			name: "error - non-existent account",
			setup: func() *CreateWorkspaceInput {
				return &CreateWorkspaceInput{
					Workspace:      testutils.GenerateRandomWorkspace(),
					AccountID:      999999,
					TenantID:       tenant.Id,
					OrganizationID: org.Id,
				}
			},
			wantErr: true,
			errType: ErrAccountDoesNotExist,
		},
		{
			name: "error - tenant from different organization",
			setup: func() *CreateWorkspaceInput {
				// Create another organization and tenant
				otherOrg, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
					Organization: testutils.GenerateRandomizedOrganization(),
				})
				require.NoError(t, err)

				otherTenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
					Tenant:         testutils.GenerateRandomizedTenant(),
					OrganizationID: otherOrg.Id,
				})
				require.NoError(t, err)

				return &CreateWorkspaceInput{
					Workspace:      testutils.GenerateRandomWorkspace(),
					AccountID:      account.Id,
					TenantID:       otherTenant.Id,
					OrganizationID: org.Id,
				}
			},
			wantErr: true,
		},
		{
			name: "error - account from different tenant",
			setup: func() *CreateWorkspaceInput {
				// Create another tenant and account
				otherTenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
					Tenant:         testutils.GenerateRandomizedTenant(),
					OrganizationID: org.Id,
				})
				require.NoError(t, err)

				otherAccount, err := conn.CreateAccount(ctx, &CreateAccountInput{
					Account:        testutils.GenerateRandomizedAccount(),
					TenantID:      otherTenant.Id,
					OrgID: org.Id,
				})
				require.NoError(t, err)

				return &CreateWorkspaceInput{
					Workspace:      testutils.GenerateRandomWorkspace(),
					AccountID:      otherAccount.Id,
					TenantID:       tenant.Id,
					OrganizationID: org.Id,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input *CreateWorkspaceInput
			if tt.setup != nil {
				input = tt.setup()
			} else {
				input = tt.input
			}

			workspace, err := conn.CreateWorkspace(ctx, input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, workspace)
			assert.NotZero(t, workspace.Id)

			if input != nil && input.Workspace != nil {
				assert.Equal(t, input.Workspace.Name, workspace.Name)
			}

			// Verify workspace was created in database
			fetchedWorkspace, err := conn.GetWorkspace(ctx, workspace.Id)
			require.NoError(t, err)
			assert.Equal(t, workspace.Id, fetchedWorkspace.Id)
			assert.Equal(t, workspace.Name, fetchedWorkspace.Name)

			// Cleanup
			err = conn.DeleteWorkspace(ctx, workspace.Id)
			require.NoError(t, err)
		})
	}
}

func TestDb_CreateWorkspace(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create test workspace
	validWorkspace := testutils.GenerateRandomWorkspace()

	type args struct {
		ctx       context.Context
		workspace *lead_scraper_servicev1.Workspace
		clean     func(t *testing.T, workspace *lead_scraper_servicev1.Workspace)
	}

	tests := []struct {
		name     string
		args     args
		wantErr  bool
		errType  error
		validate func(t *testing.T, workspace *lead_scraper_servicev1.Workspace)
	}{
		{
			name:    "[success scenario] - create new workspace",
			wantErr: false,
			args: args{
				ctx:       context.Background(),
				workspace: validWorkspace,
				clean: func(t *testing.T, workspace *lead_scraper_servicev1.Workspace) {
					err := conn.DeleteWorkspace(context.Background(), workspace.Id)
					require.NoError(t, err)
				},
			},
			validate: func(t *testing.T, workspace *lead_scraper_servicev1.Workspace) {
				assert.NotNil(t, workspace)
				assert.NotZero(t, workspace.Id)
				assert.Equal(t, validWorkspace.Name, workspace.Name)
			},
		},
		{
			name:    "[failure scenario] - nil workspace",
			wantErr: true,
			args: args{
				ctx:       context.Background(),
				workspace: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspace, err := conn.CreateWorkspace(tt.args.ctx, &CreateWorkspaceInput{
				Workspace: tt.args.workspace,
				AccountID:      tc.Account.Id,
				TenantID:       tc.Tenant.Id,
				OrganizationID: tc.Organization.Id,
			})
			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, workspace)

			if tt.validate != nil {
				tt.validate(t, workspace)
			}

			// Cleanup after test
			if tt.args.clean != nil {
				tt.args.clean(t, workspace)
			}
		})
	}
}

func TestDb_CreateWorkspace_ConcurrentCreation(t *testing.T) {
	numWorkspaces := 5
	var wg sync.WaitGroup
	errors := make(chan error, numWorkspaces)
	workspaces := make(chan *lead_scraper_servicev1.Workspace, numWorkspaces)

	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create workspaces concurrently
	for i := 0; i < numWorkspaces; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mockWorkspace := testutils.GenerateRandomWorkspace()
			workspace, err := conn.CreateWorkspace(context.Background(), &CreateWorkspaceInput{
				Workspace: mockWorkspace,
				AccountID:      tc.Account.Id,
				TenantID:       tc.Tenant.Id,
				OrganizationID: tc.Organization.Id,
			})
			if err != nil {
				errors <- err
				return
			}
			workspaces <- workspace
		}()
	}

	wg.Wait()
	close(errors)
	close(workspaces)

	// Check for errors
	for err := range errors {
		t.Errorf("Error during concurrent creation: %v", err)
	}

	// Clean up created workspaces
	for workspace := range workspaces {
		err := conn.DeleteWorkspace(context.Background(), workspace.Id)
		require.NoError(t, err)
	}
}
