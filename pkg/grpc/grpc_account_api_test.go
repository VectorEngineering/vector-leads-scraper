package grpc

import (
	"context"
	"testing"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type accountTestContext struct {
	Organization *proto.Organization
	TenantId     uint64
	Cleanup      func()
}

func initializeTestContext(t *testing.T) (*accountTestContext) {
	// create an organization and tenant first
	org := testutils.GenerateRandomizedOrganization()
	tenant := testutils.GenerateRandomizedTenant()
	
	// Ensure required fields are set
	if org.Name == "" {
		org.Name = "Test Organization"
	}
	if org.BillingEmail == "" {
		org.BillingEmail = "billing@example.com"
	}
	if org.TechnicalEmail == "" {
		org.TechnicalEmail = "tech@example.com"
	}

	createOrgResp, err := MockServer.CreateOrganization(context.Background(), &proto.CreateOrganizationRequest{
		Organization: org,
	})
	require.NoError(t, err)
	require.NotNil(t, createOrgResp)
	require.NotNil(t, createOrgResp.Organization)

	// Ensure required fields are set for tenant
	if tenant.Name == "" {
		tenant.Name = "Test Tenant"
	}
	if tenant.ApiBaseUrl == "" {
		tenant.ApiBaseUrl = "https://api.example.com"
	}

	createTenantResp, err := MockServer.CreateTenant(context.Background(), &proto.CreateTenantRequest{
		Tenant:         tenant,
		OrganizationId: createOrgResp.Organization.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createTenantResp)
	require.NotNil(t, createTenantResp.TenantId)

	createAcctResp, err := MockServer.CreateAccount(context.Background(), &proto.CreateAccountRequest{
		Account:              testutils.GenerateRandomizedAccount(),
		OrganizationId:      createOrgResp.Organization.Id,
		TenantId:            createTenantResp.TenantId,
		InitialWorkspaceName: "Default Workspace",
	})
	require.NoError(t, err)
	require.NotNil(t, createAcctResp)
	require.NotNil(t, createAcctResp.Account)

	cleanup := func() {
		ctx := context.Background()
		
		// First get all accounts for this tenant
		listAcctsResp, err := MockServer.ListAccounts(ctx, &proto.ListAccountsRequest{
			OrganizationId: createOrgResp.Organization.Id,
			TenantId:       createTenantResp.TenantId,
			PageSize:       100,
			PageNumber:     1,
		})
		if err == nil && listAcctsResp != nil {
			// For each account, delete its workspaces first
			for _, acct := range listAcctsResp.Accounts {
				// Get account details to access workspaces
				getAcctResp, err := MockServer.GetAccount(ctx, &proto.GetAccountRequest{
					Id:             acct.Id,
					OrganizationId: createOrgResp.Organization.Id,
					TenantId:       createTenantResp.TenantId,
				})
				if err == nil && getAcctResp != nil && getAcctResp.Account != nil {
					// Delete all workspaces for this account
					for _, workspace := range getAcctResp.Account.Workspaces {
						_, err := MockServer.DeleteWorkspace(ctx, &proto.DeleteWorkspaceRequest{
							Id: workspace.Id,
						})
						if err != nil {
							t.Logf("Failed to delete workspace %d: %v", workspace.Id, err)
						}
					}
				}
				
				// Now delete the account
				_, err = MockServer.DeleteAccount(ctx, &proto.DeleteAccountRequest{
					Id:             acct.Id,
					OrganizationId: createOrgResp.Organization.Id,
					TenantId:       createTenantResp.TenantId,
				})
				if err != nil {
					t.Logf("Failed to delete account %d: %v", acct.Id, err)
				}
			}
		}

		// Now that all accounts and workspaces are deleted, delete the tenant
		_, err = MockServer.DeleteTenant(ctx, &proto.DeleteTenantRequest{
			TenantId:       createTenantResp.TenantId,
			OrganizationId: createOrgResp.Organization.Id,
		})
		if err != nil {
			t.Logf("Failed to delete tenant %d: %v", createTenantResp.TenantId, err)
		}

		// Finally delete the organization
		_, err = MockServer.DeleteOrganization(ctx, &proto.DeleteOrganizationRequest{
			Id: createOrgResp.Organization.Id,
		})
		if err != nil {
			t.Logf("Failed to delete organization %d: %v", createOrgResp.Organization.Id, err)
		}
	}

	return &accountTestContext{
		Organization: createOrgResp.Organization,
		TenantId:     createTenantResp.TenantId,
		Cleanup:     cleanup,
	}
}

func TestServer_CreateAccount(t *testing.T) {
	// create an organization and tenant first
	testCtx := initializeTestContext(t)
	defer testCtx.Cleanup()

	tests := []struct {
		name    string
		req     *proto.CreateAccountRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.CreateAccountRequest{
				Account:              testutils.GenerateRandomizedAccount(),
				OrganizationId:      testCtx.Organization.Id,
				TenantId:            testCtx.TenantId,
				InitialWorkspaceName: "Test Workspace",
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid organization id",
			req: &proto.CreateAccountRequest{
				Account:              testutils.GenerateRandomizedAccount(),
				OrganizationId:      0,
				TenantId:            testCtx.TenantId,
				InitialWorkspaceName: "Test Workspace",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid tenant id",
			req: &proto.CreateAccountRequest{
				Account:              testutils.GenerateRandomizedAccount(),
				OrganizationId:      testCtx.Organization.Id,
				TenantId:            0,
				InitialWorkspaceName: "Test Workspace",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.CreateAccount(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Account)
		})
	}
}

func TestServer_GetAccount(t *testing.T) {
	// Create a test account first
	testCtx := initializeTestContext(t)
	defer testCtx.Cleanup()

	// create an account first
	createResp, err := MockServer.CreateAccount(context.Background(), &proto.CreateAccountRequest{
		Account:              testutils.GenerateRandomizedAccount(),
		OrganizationId:      testCtx.Organization.Id,
		TenantId:            testCtx.TenantId,
		InitialWorkspaceName: "Test Workspace",
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Account)

	tests := []struct {
		name    string
		req     *proto.GetAccountRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.GetAccountRequest{
				Id:             createResp.Account.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "account not found",
			req: &proto.GetAccountRequest{
				Id:             999999,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "invalid organization id",
			req: &proto.GetAccountRequest{
				Id:             createResp.Account.Id,
				OrganizationId: 0,
				TenantId:       testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid tenant id",
			req: &proto.GetAccountRequest{
				Id:             createResp.Account.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.GetAccount(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Account)
			assert.Equal(t, createResp.Account.Id, resp.Account.Id)
			assert.Equal(t, createResp.Account.Email, resp.Account.Email)
		})
	}
}

func TestServer_UpdateAccount(t *testing.T) {
	tc := initializeTestContext(t)
	defer tc.Cleanup()

	// create an account first
	createResp, err := MockServer.CreateAccount(context.Background(), &proto.CreateAccountRequest{
		Account:              testutils.GenerateRandomizedAccount(),
		OrganizationId:      tc.Organization.Id,
		TenantId:            tc.TenantId,
		InitialWorkspaceName: "Test Workspace",
	})	
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Account)

	account := createResp.Account

	tests := []struct {
		name    string
		req     *proto.UpdateAccountRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.UpdateAccountRequest{
				Payload: &proto.UpdateAccountRequestPayload{
					OrganizationId: tc.Organization.Id,
					TenantId:       tc.TenantId,
					Account:        account,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.UpdateAccount(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Account)
		})
	}
}

// func TestServer_DeleteAccount(t *testing.T) {
// 	tc := initializeTestContext(t)
// 	defer tc.Cleanup()

// 	// Create a test account first
// 	account := testutils.GenerateRandomizedAccount()
// 	createResp, err := MockServer.CreateAccount(context.Background(), &proto.CreateAccountRequest{
// 		Account:              account,
// 		OrganizationId:      tc.Organization.Id,
// 		TenantId:            tc.TenantId,
// 		InitialWorkspaceName: "Test Workspace",
// 	})
// 	require.NoError(t, err)	
// 	require.NotNil(t, createResp)
// 	require.NotNil(t, createResp.Account)

// 	account = createResp.Account
	
// 	tests := []struct {
// 		name    string
// 		req     *proto.DeleteAccountRequest
// 		wantErr bool
// 		errCode codes.Code
// 	}{
// 		{
// 			name: "success",
// 			req: &proto.DeleteAccountRequest{
// 				Id:             account.Id,
// 				OrganizationId: tc.Organization.Id,
// 				TenantId:       tc.TenantId,
// 			},
// 			wantErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// First delete all workspaces associated with the account
// 			getAcctResp, err := MockServer.GetAccount(context.Background(), &proto.GetAccountRequest{
// 				Id:             tt.req.Id,
// 				OrganizationId: tt.req.OrganizationId,
// 				TenantId:       tt.req.TenantId,
// 			})
// 			require.NoError(t, err)
// 			require.NotNil(t, getAcctResp)
// 			require.NotNil(t, getAcctResp.Account)

// 			for _, workspace := range getAcctResp.Account.Workspaces {
// 				_, err := MockServer.DeleteWorkspace(context.Background(), &proto.DeleteWorkspaceRequest{
// 					Id: workspace.Id,
// 				})
// 				require.NoError(t, err)
// 			}

// 			// Now delete the account
// 			resp, err := MockServer.DeleteAccount(context.Background(), tt.req)
// 			if tt.wantErr {
// 				require.Error(t, err)
// 				st, ok := status.FromError(err)
// 				require.True(t, ok)
// 				assert.Equal(t, tt.errCode, st.Code())
// 				return
// 			}
// 			require.NoError(t, err)
// 			require.NotNil(t, resp)
// 			assert.True(t, resp.Success)

// 			// Verify account is actually deleted
// 			getResp, err := MockServer.GetAccount(context.Background(), &proto.GetAccountRequest{
// 				Id:             tt.req.Id,
// 				OrganizationId: tt.req.OrganizationId,
// 				TenantId:       tt.req.TenantId,
// 			})
// 			require.Error(t, err)
// 			st, ok := status.FromError(err)
// 			require.True(t, ok)
// 			assert.Equal(t, codes.NotFound, st.Code())
// 			assert.Nil(t, getResp)
// 		})
// 	}
// }
