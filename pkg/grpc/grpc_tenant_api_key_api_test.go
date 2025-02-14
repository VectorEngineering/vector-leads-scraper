package grpc

// import (
// 	"context"
// 	"testing"

// 	"github.com/Vector/vector-leads-scraper/internal/testutils"
// 	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )

// type tenantAPIKeyTestContext struct {
// 	Organization *proto.Organization
// 	TenantId     uint64
// 	Cleanup      func()
// }

// func initializeTenantAPIKeyTestContext(t *testing.T) *tenantAPIKeyTestContext {
// 	// Create organization and tenant first
// 	org := testutils.GenerateRandomizedOrganization()
// 	tenant := testutils.GenerateRandomizedTenant()

// 	createOrgResp, err := MockServer.CreateOrganization(context.Background(), &proto.CreateOrganizationRequest{
// 		Organization: org,
// 	})
// 	require.NoError(t, err)
// 	require.NotNil(t, createOrgResp)
// 	require.NotNil(t, createOrgResp.Organization)

// 	createTenantResp, err := MockServer.CreateTenant(context.Background(), &proto.CreateTenantRequest{
// 		Tenant:         tenant,
// 		OrganizationId: createOrgResp.Organization.Id,
// 	})
// 	require.NoError(t, err)
// 	require.NotNil(t, createTenantResp)
// 	require.NotNil(t, createTenantResp.TenantId)

// 	// Return test context with cleanup function
// 	return &tenantAPIKeyTestContext{
// 		Organization: createOrgResp.Organization,
// 		TenantId:     createTenantResp.TenantId,
// 		Cleanup: func() {
// 			ctx := context.Background()

// 			// Delete tenant
// 			_, err = MockServer.DeleteTenant(ctx, &proto.DeleteTenantRequest{
// 				TenantId:       createTenantResp.TenantId,
// 				OrganizationId: createOrgResp.Organization.Id,
// 			})
// 			if err != nil {
// 				t.Logf("Failed to cleanup test tenant: %v", err)
// 			}

// 			// Delete organization
// 			_, err = MockServer.DeleteOrganization(ctx, &proto.DeleteOrganizationRequest{
// 				Id: createOrgResp.Organization.Id,
// 			})
// 			if err != nil {
// 				t.Logf("Failed to cleanup test organization: %v", err)
// 			}
// 		},
// 	}
// }

// func TestServer_CreateTenantAPIKey(t *testing.T) {
// 	testCtx := initializeTenantAPIKeyTestContext(t)
// 	defer testCtx.Cleanup()

// 	tests := []struct {
// 		name    string
// 		req     *proto.CreateTenantAPIKeyRequest
// 		wantErr bool
// 		errCode codes.Code
// 		verify  func(t *testing.T, resp *proto.CreateTenantAPIKeyResponse)
// 	}{
// 		{
// 			name: "success",
// 			req: &proto.CreateTenantAPIKeyRequest{
// 				Name:     "Test Tenant API Key",
// 				TenantId: testCtx.TenantId,
// 				Scopes:   []string{"read:leads", "write:leads"},
// 			},
// 			wantErr: false,
// 			verify: func(t *testing.T, resp *proto.CreateTenantAPIKeyResponse) {
// 				require.NotNil(t, resp)
// 				require.NotNil(t, resp.ApiKey)
// 				assert.NotEmpty(t, resp.ApiKey.Id)
// 				assert.Equal(t, "Test Tenant API Key", resp.ApiKey.Name)
// 				assert.Equal(t, []string{"read:leads", "write:leads"}, resp.ApiKey.Scopes)
// 			},
// 		},
// 		{
// 			name:    "error - nil request",
// 			req:     nil,
// 			wantErr: true,
// 			errCode: codes.InvalidArgument,
// 		},
// 		{
// 			name: "error - missing name",
// 			req: &proto.CreateTenantAPIKeyRequest{
// 				TenantId: testCtx.TenantId,
// 				Scopes:   []string{"read:leads"},
// 			},
// 			wantErr: true,
// 			errCode: codes.InvalidArgument,
// 		},
// 		{
// 			name: "error - missing scopes",
// 			req: &proto.CreateTenantAPIKeyRequest{
// 				Name:     "Test Tenant API Key",
// 				TenantId: testCtx.TenantId,
// 			},
// 			wantErr: true,
// 			errCode: codes.InvalidArgument,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			resp, err := MockServer.CreateTenantAPIKey(context.Background(), tt.req)
// 			if tt.wantErr {
// 				require.Error(t, err)
// 				st, ok := status.FromError(err)
// 				require.True(t, ok)
// 				assert.Equal(t, tt.errCode, st.Code())
// 				return
// 			}

// 			require.NoError(t, err)
// 			if tt.verify != nil {
// 				tt.verify(t, resp)
// 			}
// 		})
// 	}
// }

// func TestServer_GetTenantAPIKey(t *testing.T) {
// 	testCtx := initializeTenantAPIKeyTestContext(t)
// 	defer testCtx.Cleanup()

// 	// Create a test tenant API key first
// 	createResp, err := MockServer.CreateTenantAPIKey(context.Background(), &proto.CreateTenantAPIKeyRequest{
// 		Name:     "Test Tenant API Key",
// 		TenantId: testCtx.TenantId,
// 		Scopes:   []string{"read:leads", "write:leads"},
// 	})
// 	require.NoError(t, err)
// 	require.NotNil(t, createResp)
// 	require.NotNil(t, createResp.ApiKey)

// 	tests := []struct {
// 		name    string
// 		req     *proto.GetTenantAPIKeyRequest
// 		wantErr bool
// 		errCode codes.Code
// 		verify  func(t *testing.T, resp *proto.GetTenantAPIKeyResponse)
// 	}{
// 		{
// 			name: "success",
// 			req: &proto.GetTenantAPIKeyRequest{
// 				KeyId:    createResp.ApiKey.Id,
// 				TenantId: testCtx.TenantId,
// 			},
// 			wantErr: false,
// 			verify: func(t *testing.T, resp *proto.GetTenantAPIKeyResponse) {
// 				require.NotNil(t, resp)
// 				require.NotNil(t, resp.ApiKey)
// 				assert.Equal(t, createResp.ApiKey.Id, resp.ApiKey.Id)
// 				assert.Equal(t, createResp.ApiKey.Name, resp.ApiKey.Name)
// 				assert.Equal(t, createResp.ApiKey.Scopes, resp.ApiKey.Scopes)
// 			},
// 		},
// 		{
// 			name:    "error - nil request",
// 			req:     nil,
// 			wantErr: true,
// 			errCode: codes.InvalidArgument,
// 		},
// 		{
// 			name: "error - API key not found",
// 			req: &proto.GetTenantAPIKeyRequest{
// 				KeyId:    "non-existent",
// 				TenantId: testCtx.TenantId,
// 			},
// 			wantErr: true,
// 			errCode: codes.NotFound,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			resp, err := MockServer.GetTenantAPIKey(context.Background(), tt.req)
// 			if tt.wantErr {
// 				require.Error(t, err)
// 				st, ok := status.FromError(err)
// 				require.True(t, ok)
// 				assert.Equal(t, tt.errCode, st.Code())
// 				return
// 			}

// 			require.NoError(t, err)
// 			if tt.verify != nil {
// 				tt.verify(t, resp)
// 			}
// 		})
// 	}
// }

// func TestServer_UpdateTenantAPIKey(t *testing.T) {
// 	testCtx := initializeTenantAPIKeyTestContext(t)
// 	defer testCtx.Cleanup()

// 	// Create a test tenant API key first
// 	createResp, err := MockServer.CreateTenantAPIKey(context.Background(), &proto.CreateTenantAPIKeyRequest{
// 		Name:     "Test Tenant API Key",
// 		TenantId: testCtx.TenantId,
// 		Scopes:   []string{"read:leads", "write:leads"},
// 	})
// 	require.NoError(t, err)
// 	require.NotNil(t, createResp)
// 	require.NotNil(t, createResp.ApiKey)

// 	tests := []struct {
// 		name    string
// 		req     *proto.UpdateTenantAPIKeyRequest
// 		wantErr bool
// 		errCode codes.Code
// 		verify  func(t *testing.T, resp *proto.UpdateTenantAPIKeyResponse)
// 	}{
// 		{
// 			name: "success",
// 			req: &proto.UpdateTenantAPIKeyRequest{
// 				KeyId:    createResp.ApiKey.Id,
// 				TenantId: testCtx.TenantId,
// 				Name:     "Updated Tenant API Key",
// 				Scopes:   []string{"read:leads"},
// 			},
// 			wantErr: false,
// 			verify: func(t *testing.T, resp *proto.UpdateTenantAPIKeyResponse) {
// 				require.NotNil(t, resp)
// 				require.NotNil(t, resp.ApiKey)
// 				assert.Equal(t, createResp.ApiKey.Id, resp.ApiKey.Id)
// 				assert.Equal(t, "Updated Tenant API Key", resp.ApiKey.Name)
// 				assert.Equal(t, []string{"read:leads"}, resp.ApiKey.Scopes)
// 			},
// 		},
// 		{
// 			name:    "error - nil request",
// 			req:     nil,
// 			wantErr: true,
// 			errCode: codes.InvalidArgument,
// 		},
// 		{
// 			name: "error - API key not found",
// 			req: &proto.UpdateTenantAPIKeyRequest{
// 				KeyId:    "non-existent",
// 				TenantId: testCtx.TenantId,
// 				Name:     "Updated Tenant API Key",
// 				Scopes:   []string{"read:leads"},
// 			},
// 			wantErr: true,
// 			errCode: codes.NotFound,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			resp, err := MockServer.UpdateTenantAPIKey(context.Background(), tt.req)
// 			if tt.wantErr {
// 				require.Error(t, err)
// 				st, ok := status.FromError(err)
// 				require.True(t, ok)
// 				assert.Equal(t, tt.errCode, st.Code())
// 				return
// 			}

// 			require.NoError(t, err)
// 			if tt.verify != nil {
// 				tt.verify(t, resp)
// 			}
// 		})
// 	}
// }

// func TestServer_DeleteTenantAPIKey(t *testing.T) {
// 	testCtx := initializeTenantAPIKeyTestContext(t)
// 	defer testCtx.Cleanup()

// 	// Create a test tenant API key first
// 	createResp, err := MockServer.CreateTenantAPIKey(context.Background(), &proto.CreateTenantAPIKeyRequest{
// 		Name:     "Test Tenant API Key",
// 		TenantId: testCtx.TenantId,
// 		Scopes:   []string{"read:leads", "write:leads"},
// 	})
// 	require.NoError(t, err)
// 	require.NotNil(t, createResp)
// 	require.NotNil(t, createResp.ApiKey)

// 	tests := []struct {
// 		name    string
// 		req     *proto.DeleteTenantAPIKeyRequest
// 		wantErr bool
// 		errCode codes.Code
// 	}{
// 		{
// 			name: "success",
// 			req: &proto.DeleteTenantAPIKeyRequest{
// 				KeyId:    createResp.ApiKey.Id,
// 				TenantId: testCtx.TenantId,
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name:    "error - nil request",
// 			req:     nil,
// 			wantErr: true,
// 			errCode: codes.InvalidArgument,
// 		},
// 		{
// 			name: "error - API key not found",
// 			req: &proto.DeleteTenantAPIKeyRequest{
// 				KeyId:    "non-existent",
// 				TenantId: testCtx.TenantId,
// 			},
// 			wantErr: true,
// 			errCode: codes.NotFound,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			resp, err := MockServer.DeleteTenantAPIKey(context.Background(), tt.req)
// 			if tt.wantErr {
// 				require.Error(t, err)
// 				st, ok := status.FromError(err)
// 				require.True(t, ok)
// 				assert.Equal(t, tt.errCode, st.Code())
// 				return
// 			}

// 			require.NoError(t, err)
// 			require.NotNil(t, resp)
// 		})
// 	}
// }

// func TestServer_ListTenantAPIKeys(t *testing.T) {
// 	testCtx := initializeTenantAPIKeyTestContext(t)
// 	defer testCtx.Cleanup()

// 	// Create some test tenant API keys first
// 	numKeys := 5
// 	apiKeys := make([]*proto.APIKey, 0, numKeys)
// 	for i := 0; i < numKeys; i++ {
// 		createResp, err := MockServer.CreateTenantAPIKey(context.Background(), &proto.CreateTenantAPIKeyRequest{
// 			Name:     "Test Tenant API Key",
// 			TenantId: testCtx.TenantId,
// 			Scopes:   []string{"read:leads", "write:leads"},
// 		})
// 		require.NoError(t, err)
// 		require.NotNil(t, createResp)
// 		require.NotNil(t, createResp.ApiKey)
// 		apiKeys = append(apiKeys, createResp.ApiKey)
// 	}

// 	tests := []struct {
// 		name    string
// 		req     *proto.ListTenantAPIKeysRequest
// 		wantErr bool
// 		errCode codes.Code
// 		verify  func(t *testing.T, resp *proto.ListTenantAPIKeysResponse)
// 	}{
// 		{
// 			name: "success - list all API keys",
// 			req: &proto.ListTenantAPIKeysRequest{
// 				TenantId:   testCtx.TenantId,
// 				PageSize:   10,
// 				PageNumber: 1,
// 			},
// 			wantErr: false,
// 			verify: func(t *testing.T, resp *proto.ListTenantAPIKeysResponse) {
// 				require.NotNil(t, resp)
// 				assert.Len(t, resp.ApiKeys, numKeys)
// 				assert.Equal(t, int32(0), resp.NextPageNumber) // No more pages
// 			},
// 		},
// 		{
// 			name: "success - pagination",
// 			req: &proto.ListTenantAPIKeysRequest{
// 				TenantId:   testCtx.TenantId,
// 				PageSize:   2,
// 				PageNumber: 1,
// 			},
// 			wantErr: false,
// 			verify: func(t *testing.T, resp *proto.ListTenantAPIKeysResponse) {
// 				require.NotNil(t, resp)
// 				assert.Len(t, resp.ApiKeys, 2)
// 				assert.Equal(t, int32(2), resp.NextPageNumber)
// 			},
// 		},
// 		{
// 			name:    "error - nil request",
// 			req:     nil,
// 			wantErr: true,
// 			errCode: codes.InvalidArgument,
// 		},
// 		{
// 			name: "error - invalid page size",
// 			req: &proto.ListTenantAPIKeysRequest{
// 				TenantId:   testCtx.TenantId,
// 				PageSize:   0,
// 				PageNumber: 1,
// 			},
// 			wantErr: true,
// 			errCode: codes.InvalidArgument,
// 		},
// 		{
// 			name: "error - invalid page number",
// 			req: &proto.ListTenantAPIKeysRequest{
// 				TenantId:   testCtx.TenantId,
// 				PageSize:   10,
// 				PageNumber: 0,
// 			},
// 			wantErr: true,
// 			errCode: codes.InvalidArgument,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			resp, err := MockServer.ListTenantAPIKeys(context.Background(), tt.req)
// 			if tt.wantErr {
// 				require.Error(t, err)
// 				st, ok := status.FromError(err)
// 				require.True(t, ok)
// 				assert.Equal(t, tt.errCode, st.Code())
// 				return
// 			}

// 			require.NoError(t, err)
// 			if tt.verify != nil {
// 				tt.verify(t, resp)
// 			}
// 		})
// 	}
// }

// func TestServer_RotateTenantAPIKey(t *testing.T) {
// 	testCtx := initializeTenantAPIKeyTestContext(t)
// 	defer testCtx.Cleanup()

// 	// Create a test tenant API key first
// 	createResp, err := MockServer.CreateTenantAPIKey(context.Background(), &proto.CreateTenantAPIKeyRequest{
// 		Name:     "Test Tenant API Key",
// 		TenantId: testCtx.TenantId,
// 		Scopes:   []string{"read:leads", "write:leads"},
// 	})
// 	require.NoError(t, err)
// 	require.NotNil(t, createResp)
// 	require.NotNil(t, createResp.ApiKey)

// 	tests := []struct {
// 		name    string
// 		req     *proto.RotateTenantAPIKeyRequest
// 		wantErr bool
// 		errCode codes.Code
// 		verify  func(t *testing.T, resp *proto.RotateTenantAPIKeyResponse)
// 	}{
// 		{
// 			name: "success",
// 			req: &proto.RotateTenantAPIKeyRequest{
// 				KeyId:    createResp.ApiKey.Id,
// 				TenantId: testCtx.TenantId,
// 			},
// 			wantErr: false,
// 			verify: func(t *testing.T, resp *proto.RotateTenantAPIKeyResponse) {
// 				require.NotNil(t, resp)
// 				require.NotNil(t, resp.ApiKey)
// 				assert.NotEqual(t, createResp.ApiKey.Id, resp.ApiKey.Id)
// 				assert.Equal(t, createResp.ApiKey.Name, resp.ApiKey.Name)
// 				assert.Equal(t, createResp.ApiKey.Scopes, resp.ApiKey.Scopes)
// 			},
// 		},
// 		{
// 			name:    "error - nil request",
// 			req:     nil,
// 			wantErr: true,
// 			errCode: codes.InvalidArgument,
// 		},
// 		{
// 			name: "error - API key not found",
// 			req: &proto.RotateTenantAPIKeyRequest{
// 				KeyId:    "non-existent",
// 				TenantId: testCtx.TenantId,
// 			},
// 			wantErr: true,
// 			errCode: codes.NotFound,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			resp, err := MockServer.RotateTenantAPIKey(context.Background(), tt.req)
// 			if tt.wantErr {
// 				require.Error(t, err)
// 				st, ok := status.FromError(err)
// 				require.True(t, ok)
// 				assert.Equal(t, tt.errCode, st.Code())
// 				return
// 			}

// 			require.NoError(t, err)
// 			if tt.verify != nil {
// 				tt.verify(t, resp)
// 			}
// 		})
// 	}
// }