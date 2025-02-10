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

type apiKeyTestContext struct {
	Organization *proto.Organization
	TenantId     uint64
	Account      *proto.Account
	Workspace    *proto.Workspace
	Cleanup      func()
}

func initializeAPIKeyTestContext(t *testing.T) *apiKeyTestContext {
	// Create organization and tenant first
	org := testutils.GenerateRandomizedOrganization()
	tenant := testutils.GenerateRandomizedTenant()

	createOrgResp, err := MockServer.CreateOrganization(context.Background(), &proto.CreateOrganizationRequest{
		Organization: org,
	})
	require.NoError(t, err)
	require.NotNil(t, createOrgResp)
	require.NotNil(t, createOrgResp.Organization)

	createTenantResp, err := MockServer.CreateTenant(context.Background(), &proto.CreateTenantRequest{
		Tenant:         tenant,
		OrganizationId: createOrgResp.Organization.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createTenantResp)
	require.NotNil(t, createTenantResp.TenantId)

	// Create account
	account := testutils.GenerateRandomizedAccount()
	createAcctResp, err := MockServer.CreateAccount(context.Background(), &proto.CreateAccountRequest{
		Account:        account,
		OrganizationId: createOrgResp.Organization.Id,
		TenantId:       createTenantResp.TenantId,
	})
	require.NoError(t, err)
	require.NotNil(t, createAcctResp)
	require.NotNil(t, createAcctResp.Account)

	// Create workspace
	workspace := testutils.GenerateRandomizedWorkspace()
	createWorkspaceResp, err := MockServer.CreateWorkspace(context.Background(), &proto.CreateWorkspaceRequest{
		Workspace:      workspace,
		AccountId:      createAcctResp.Account.Id,
		TenantId:       createTenantResp.TenantId,
		OrganizationId: createOrgResp.Organization.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createWorkspaceResp)
	require.NotNil(t, createWorkspaceResp.Workspace)

	// Return test context with cleanup function
	return &apiKeyTestContext{
		Organization: createOrgResp.Organization,
		TenantId:     createTenantResp.TenantId,
		Account:      createAcctResp.Account,
		Workspace:    createWorkspaceResp.Workspace,
		Cleanup: func() {
			ctx := context.Background()

			// Delete workspace
			_, err := MockServer.DeleteWorkspace(ctx, &proto.DeleteWorkspaceRequest{
				Id: createWorkspaceResp.Workspace.Id,
			})
			if err != nil {
				t.Logf("Failed to cleanup test workspace: %v", err)
			}

			// Delete account
			_, err = MockServer.DeleteAccount(ctx, &proto.DeleteAccountRequest{
				Id:             createAcctResp.Account.Id,
				OrganizationId: createOrgResp.Organization.Id,
				TenantId:       createTenantResp.TenantId,
			})
			if err != nil {
				t.Logf("Failed to cleanup test account: %v", err)
			}

			// Delete tenant
			_, err = MockServer.DeleteTenant(ctx, &proto.DeleteTenantRequest{
				TenantId:       createTenantResp.TenantId,
				OrganizationId: createOrgResp.Organization.Id,
			})
			if err != nil {
				t.Logf("Failed to cleanup test tenant: %v", err)
			}

			// Delete organization
			_, err = MockServer.DeleteOrganization(ctx, &proto.DeleteOrganizationRequest{
				Id: createOrgResp.Organization.Id,
			})
			if err != nil {
				t.Logf("Failed to cleanup test organization: %v", err)
			}
		},
	}
}

func TestServer_CreateAPIKey(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	tests := []struct {
		name    string
		req     *proto.CreateAPIKeyRequest
		wantErr bool
		errCode codes.Code
		verify  func(t *testing.T, resp *proto.CreateAPIKeyResponse)
	}{
		{
			name: "success",
			req: &proto.CreateAPIKeyRequest{
				Name:        "Test API Key",
				WorkspaceId: testCtx.Workspace.Id,
				Scopes:      []string{"read:leads", "write:leads"},
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.CreateAPIKeyResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.ApiKey)
				assert.NotEmpty(t, resp.ApiKey.Id)
				assert.Equal(t, "Test API Key", resp.ApiKey.Name)
				assert.Equal(t, []string{"read:leads", "write:leads"}, resp.ApiKey.Scopes)
			},
		},
		{
			name:    "error - nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - missing name",
			req: &proto.CreateAPIKeyRequest{
				WorkspaceId: testCtx.Workspace.Id,
				Scopes:      []string{"read:leads"},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - missing scopes",
			req: &proto.CreateAPIKeyRequest{
				Name:        "Test API Key",
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.CreateAPIKey(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}

			require.NoError(t, err)
			if tt.verify != nil {
				tt.verify(t, resp)
			}
		})
	}
}

func TestServer_GetAPIKey(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Create a test API key first
	createResp, err := MockServer.CreateAPIKey(context.Background(), &proto.CreateAPIKeyRequest{
		Name:        "Test API Key",
		WorkspaceId: testCtx.Workspace.Id,
		Scopes:      []string{"read:leads", "write:leads"},
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.ApiKey)

	tests := []struct {
		name    string
		req     *proto.GetAPIKeyRequest
		wantErr bool
		errCode codes.Code
		verify  func(t *testing.T, resp *proto.GetAPIKeyResponse)
	}{
		{
			name: "success",
			req: &proto.GetAPIKeyRequest{
				KeyId:       createResp.ApiKey.Id,
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.GetAPIKeyResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.ApiKey)
				assert.Equal(t, createResp.ApiKey.Id, resp.ApiKey.Id)
				assert.Equal(t, createResp.ApiKey.Name, resp.ApiKey.Name)
				assert.Equal(t, createResp.ApiKey.Scopes, resp.ApiKey.Scopes)
			},
		},
		{
			name:    "error - nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - API key not found",
			req: &proto.GetAPIKeyRequest{
				KeyId:       "non-existent",
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.GetAPIKey(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}

			require.NoError(t, err)
			if tt.verify != nil {
				tt.verify(t, resp)
			}
		})
	}
}

func TestServer_UpdateAPIKey(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Create a test API key first
	createResp, err := MockServer.CreateAPIKey(context.Background(), &proto.CreateAPIKeyRequest{
		Name:        "Test API Key",
		WorkspaceId: testCtx.Workspace.Id,
		Scopes:      []string{"read:leads", "write:leads"},
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.ApiKey)

	tests := []struct {
		name    string
		req     *proto.UpdateAPIKeyRequest
		wantErr bool
		errCode codes.Code
		verify  func(t *testing.T, resp *proto.UpdateAPIKeyResponse)
	}{
		{
			name: "success",
			req: &proto.UpdateAPIKeyRequest{
				KeyId:       createResp.ApiKey.Id,
				WorkspaceId: testCtx.Workspace.Id,
				Name:        "Updated API Key",
				Scopes:      []string{"read:leads"},
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.UpdateAPIKeyResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.ApiKey)
				assert.Equal(t, createResp.ApiKey.Id, resp.ApiKey.Id)
				assert.Equal(t, "Updated API Key", resp.ApiKey.Name)
				assert.Equal(t, []string{"read:leads"}, resp.ApiKey.Scopes)
			},
		},
		{
			name:    "error - nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - API key not found",
			req: &proto.UpdateAPIKeyRequest{
				KeyId:       "non-existent",
				WorkspaceId: testCtx.Workspace.Id,
				Name:        "Updated API Key",
				Scopes:      []string{"read:leads"},
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.UpdateAPIKey(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}

			require.NoError(t, err)
			if tt.verify != nil {
				tt.verify(t, resp)
			}
		})
	}
}

func TestServer_DeleteAPIKey(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Create a test API key first
	createResp, err := MockServer.CreateAPIKey(context.Background(), &proto.CreateAPIKeyRequest{
		Name:        "Test API Key",
		WorkspaceId: testCtx.Workspace.Id,
		Scopes:      []string{"read:leads", "write:leads"},
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.ApiKey)

	tests := []struct {
		name    string
		req     *proto.DeleteAPIKeyRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.DeleteAPIKeyRequest{
				KeyId:       createResp.ApiKey.Id,
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: false,
		},
		{
			name:    "error - nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - API key not found",
			req: &proto.DeleteAPIKeyRequest{
				KeyId:       "non-existent",
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.DeleteAPIKey(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

func TestServer_ListAPIKeys(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Create some test API keys first
	numKeys := 5
	apiKeys := make([]*proto.APIKey, 0, numKeys)
	for i := 0; i < numKeys; i++ {
		createResp, err := MockServer.CreateAPIKey(context.Background(), &proto.CreateAPIKeyRequest{
			Name:        "Test API Key",
			WorkspaceId: testCtx.Workspace.Id,
			Scopes:      []string{"read:leads", "write:leads"},
		})
		require.NoError(t, err)
		require.NotNil(t, createResp)
		require.NotNil(t, createResp.ApiKey)
		apiKeys = append(apiKeys, createResp.ApiKey)
	}

	tests := []struct {
		name    string
		req     *proto.ListAPIKeysRequest
		wantErr bool
		errCode codes.Code
		verify  func(t *testing.T, resp *proto.ListAPIKeysResponse)
	}{
		{
			name: "success - list all API keys",
			req: &proto.ListAPIKeysRequest{
				WorkspaceId: testCtx.Workspace.Id,
				PageSize:    10,
				PageNumber:  1,
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.ListAPIKeysResponse) {
				require.NotNil(t, resp)
				assert.Len(t, resp.ApiKeys, numKeys)
				assert.Equal(t, int32(0), resp.NextPageNumber) // No more pages
			},
		},
		{
			name: "success - pagination",
			req: &proto.ListAPIKeysRequest{
				WorkspaceId: testCtx.Workspace.Id,
				PageSize:    2,
				PageNumber:  1,
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.ListAPIKeysResponse) {
				require.NotNil(t, resp)
				assert.Len(t, resp.ApiKeys, 2)
				assert.Equal(t, int32(2), resp.NextPageNumber)
			},
		},
		{
			name:    "error - nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - invalid page size",
			req: &proto.ListAPIKeysRequest{
				WorkspaceId: testCtx.Workspace.Id,
				PageSize:    0,
				PageNumber:  1,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - invalid page number",
			req: &proto.ListAPIKeysRequest{
				WorkspaceId: testCtx.Workspace.Id,
				PageSize:    10,
				PageNumber:  0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.ListAPIKeys(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}

			require.NoError(t, err)
			if tt.verify != nil {
				tt.verify(t, resp)
			}
		})
	}
}

func TestServer_RotateAPIKey(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Create a test API key first
	createResp, err := MockServer.CreateAPIKey(context.Background(), &proto.CreateAPIKeyRequest{
		Name:        "Test API Key",
		WorkspaceId: testCtx.Workspace.Id,
		Scopes:      []string{"read:leads", "write:leads"},
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.ApiKey)

	tests := []struct {
		name    string
		req     *proto.RotateAPIKeyRequest
		wantErr bool
		errCode codes.Code
		verify  func(t *testing.T, resp *proto.RotateAPIKeyResponse)
	}{
		{
			name: "success",
			req: &proto.RotateAPIKeyRequest{
				KeyId:       createResp.ApiKey.Id,
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.RotateAPIKeyResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.ApiKey)
				assert.NotEqual(t, createResp.ApiKey.Id, resp.ApiKey.Id)
				assert.Equal(t, createResp.ApiKey.Name, resp.ApiKey.Name)
				assert.Equal(t, createResp.ApiKey.Scopes, resp.ApiKey.Scopes)
			},
		},
		{
			name:    "error - nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - API key not found",
			req: &proto.RotateAPIKeyRequest{
				KeyId:       "non-existent",
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.RotateAPIKey(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}

			require.NoError(t, err)
			if tt.verify != nil {
				tt.verify(t, resp)
			}
		})
	}
} 