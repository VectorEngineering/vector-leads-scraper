package grpc

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type workspaceTestContext struct {
	Organization *proto.Organization
	TenantId     uint64
	Account      *proto.Account
	Cleanup      func()
}

func initializeWorkspaceTestContext(t *testing.T) *workspaceTestContext {
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

	// Create a test account
	createAcctResp, err := MockServer.CreateAccount(context.Background(), &proto.CreateAccountRequest{
		Account:              testutils.GenerateRandomizedAccount(),
		OrganizationId:       createOrgResp.Organization.Id,
		TenantId:             createTenantResp.TenantId,
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

		// Delete tenant
		_, err = MockServer.DeleteTenant(ctx, &proto.DeleteTenantRequest{
			TenantId:       createTenantResp.TenantId,
			OrganizationId: createOrgResp.Organization.Id,
		})
		if err != nil {
			t.Logf("Failed to delete tenant %d: %v", createTenantResp.TenantId, err)
		}

		// Delete organization
		_, err = MockServer.DeleteOrganization(ctx, &proto.DeleteOrganizationRequest{
			Id: createOrgResp.Organization.Id,
		})
		if err != nil {
			t.Logf("Failed to delete organization %d: %v", createOrgResp.Organization.Id, err)
		}
	}

	return &workspaceTestContext{
		Organization: createOrgResp.Organization,
		TenantId:     createTenantResp.TenantId,
		Account:      createAcctResp.Account,
		Cleanup:      cleanup,
	}
}

func TestServer_CreateWorkspace(t *testing.T) {
	testCtx := initializeWorkspaceTestContext(t)
	defer testCtx.Cleanup()

	tests := []struct {
		name    string
		req     *proto.CreateWorkspaceRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.CreateWorkspaceRequest{
				Workspace:      testutils.GenerateRandomWorkspace(),
				AccountId:      testCtx.Account.Id,
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
			name: "invalid account id",
			req: &proto.CreateWorkspaceRequest{
				Workspace:      testutils.GenerateRandomWorkspace(),
				AccountId:      0,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid organization id",
			req: &proto.CreateWorkspaceRequest{
				Workspace:      testutils.GenerateRandomWorkspace(),
				AccountId:      testCtx.Account.Id,
				OrganizationId: 0,
				TenantId:       testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid tenant id",
			req: &proto.CreateWorkspaceRequest{
				Workspace:      testutils.GenerateRandomWorkspace(),
				AccountId:      testCtx.Account.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.CreateWorkspace(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Workspace)
			assert.NotEmpty(t, resp.Workspace.Id)
			assert.Equal(t, tt.req.Workspace.Name, resp.Workspace.Name)
		})
	}
}

func TestServer_GetWorkspace(t *testing.T) {
	testCtx := initializeWorkspaceTestContext(t)
	defer testCtx.Cleanup()

	// Create a test workspace first
	createResp, err := MockServer.CreateWorkspace(context.Background(), &proto.CreateWorkspaceRequest{
		Workspace:      testutils.GenerateRandomWorkspace(),
		AccountId:      testCtx.Account.Id,
		OrganizationId: testCtx.Organization.Id,
		TenantId:       testCtx.TenantId,
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Workspace)

	tests := []struct {
		name    string
		req     *proto.GetWorkspaceRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.GetWorkspaceRequest{
				Id:             createResp.Workspace.Id,
				AccountId:      testCtx.Account.Id,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.GetWorkspace(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Workspace)
			assert.Equal(t, createResp.Workspace.Id, resp.Workspace.Id)
			assert.Equal(t, createResp.Workspace.Name, resp.Workspace.Name)
		})
	}
}

func TestServer_UpdateWorkspace(t *testing.T) {
	testCtx := initializeWorkspaceTestContext(t)
	defer testCtx.Cleanup()

	req := &proto.CreateWorkspaceRequest{
		Workspace:      testutils.GenerateRandomWorkspace(),
		AccountId:      testCtx.Account.Id,
		OrganizationId: testCtx.Organization.Id,
		TenantId:       testCtx.TenantId,
	}

	// Create a test workspace first
	createResp, err := MockServer.CreateWorkspace(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Workspace)

	workspace := createResp.Workspace
	workspace.Name = "Updated Workspace Name"
	workspace.Industry = "Technology"
	workspace.Domain = "updated-domain.com"
	workspace.GdprCompliant = true
	workspace.HipaaCompliant = true
	workspace.Soc2Compliant = true
	workspace.StorageQuota = 1000000
	workspace.UsedStorage = 500000
	workspace.WorkspaceJobLimit = 100

	tests := []struct {
		name    string
		req     *proto.UpdateWorkspaceRequest
		wantErr bool
		errCode codes.Code
		validate func(t *testing.T, resp *proto.UpdateWorkspaceResponse)
	}{
		{
			name: "success - update all fields",
			req: &proto.UpdateWorkspaceRequest{
				Workspace: workspace,
			},
			wantErr: false,
			validate: func(t *testing.T, resp *proto.UpdateWorkspaceResponse) {
				assert.Equal(t, workspace.Id, resp.Workspace.Id)
				assert.Equal(t, "Updated Workspace Name", resp.Workspace.Name)
				assert.Equal(t, "Technology", resp.Workspace.Industry)
				assert.Equal(t, "updated-domain.com", resp.Workspace.Domain)
				assert.True(t, resp.Workspace.GdprCompliant)
				assert.True(t, resp.Workspace.HipaaCompliant)
				assert.True(t, resp.Workspace.Soc2Compliant)
				assert.Equal(t, int64(1000000), resp.Workspace.StorageQuota)
				assert.Equal(t, int64(500000), resp.Workspace.UsedStorage)
				assert.Equal(t, int32(100), resp.Workspace.WorkspaceJobLimit)
			},
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "nil workspace",
			req: &proto.UpdateWorkspaceRequest{
				Workspace: nil,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid workspace ID",
			req: &proto.UpdateWorkspaceRequest{
				Workspace: &proto.Workspace{
					Id:   0,
					Name: "Invalid ID Workspace",
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "workspace not found",
			req: &proto.UpdateWorkspaceRequest{
				Workspace: &proto.Workspace{
					Id:                999999,
					Name:              "Non-existent Workspace",
					WorkspaceJobLimit: 100,
					StorageQuota:      1000000,
					UsedStorage:       0,
				},
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "invalid storage quota",
			req: &proto.UpdateWorkspaceRequest{
				Workspace: &proto.Workspace{
					Id:           workspace.Id,
					Name:         "Invalid Storage Quota",
					StorageQuota: -1,
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid used storage",
			req: &proto.UpdateWorkspaceRequest{
				Workspace: &proto.Workspace{
					Id:          workspace.Id,
					Name:        "Invalid Used Storage",
					UsedStorage: -1,
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid job limit",
			req: &proto.UpdateWorkspaceRequest{
				Workspace: &proto.Workspace{
					Id:                workspace.Id,
					Name:              "Invalid Job Limit",
					WorkspaceJobLimit: -1,
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.UpdateWorkspace(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Workspace)
			if tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}

	// Test concurrent updates
	t.Run("concurrent updates", func(t *testing.T) {
		numUpdates := 5
		var wg sync.WaitGroup
		errors := make(chan error, numUpdates)
		results := make(chan *proto.UpdateWorkspaceResponse, numUpdates)

		for i := 0; i < numUpdates; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				updateWorkspace := &proto.Workspace{
					Id:                workspace.Id,
					Name:              fmt.Sprintf("Concurrent Update %d", index),
					Industry:          fmt.Sprintf("Industry %d", index),
					Domain:            fmt.Sprintf("domain-%d.com", index),
					GdprCompliant:     true,
					HipaaCompliant:    true,
					Soc2Compliant:     true,
					StorageQuota:      int64(1000000 + index),
					UsedStorage:       int64(500000 + index),
					WorkspaceJobLimit: int32(100 + index),
				}
				resp, err := MockServer.UpdateWorkspace(context.Background(), &proto.UpdateWorkspaceRequest{
					Workspace: updateWorkspace,
				})
				if err != nil {
					errors <- err
					return
				}
				results <- resp
			}(i)
		}

		wg.Wait()
		close(errors)
		close(results)

		// Check for errors
		var errs []error
		for err := range errors {
			errs = append(errs, err)
		}
		require.Empty(t, errs, "Expected no errors during concurrent updates")

		// Verify all updates were successful
		var updates []*proto.UpdateWorkspaceResponse
		for result := range results {
			updates = append(updates, result)
		}
		require.Len(t, updates, numUpdates, "Expected %d successful updates", numUpdates)

		// Verify final state
		getResp, err := MockServer.GetWorkspace(context.Background(), &proto.GetWorkspaceRequest{
			Id:             workspace.Id,
			AccountId:      testCtx.Account.Id,
			OrganizationId: testCtx.Organization.Id,
			TenantId:       testCtx.TenantId,
		})
		require.NoError(t, err)
		require.NotNil(t, getResp)
		require.NotNil(t, getResp.Workspace)
		assert.Contains(t, getResp.Workspace.Name, "Concurrent Update")
		assert.Contains(t, getResp.Workspace.Industry, "Industry")
		assert.Contains(t, getResp.Workspace.Domain, "domain-")
		assert.True(t, getResp.Workspace.GdprCompliant)
		assert.True(t, getResp.Workspace.HipaaCompliant)
		assert.True(t, getResp.Workspace.Soc2Compliant)
	})
}

func TestServer_DeleteWorkspace(t *testing.T) {
	testCtx := initializeWorkspaceTestContext(t)
	defer testCtx.Cleanup()

	// Create a test workspace first
	createResp, err := MockServer.CreateWorkspace(context.Background(), &proto.CreateWorkspaceRequest{
		Workspace:      testutils.GenerateRandomWorkspace(),
		AccountId:      testCtx.Account.Id,
		OrganizationId: testCtx.Organization.Id,
		TenantId:       testCtx.TenantId,
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Workspace)

	tests := []struct {
		name    string
		req     *proto.DeleteWorkspaceRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.DeleteWorkspaceRequest{
				Id: createResp.Workspace.Id,
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
			name: "workspace not found",
			req: &proto.DeleteWorkspaceRequest{
				Id: 999999,
			},
			wantErr: true,
			errCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.DeleteWorkspace(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)

			// Verify workspace is actually deleted
			getResp, err := MockServer.GetWorkspace(context.Background(), &proto.GetWorkspaceRequest{
				Id:             tt.req.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
			})
			require.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.Internal, st.Code())
			assert.Nil(t, getResp)
		})
	}
}
