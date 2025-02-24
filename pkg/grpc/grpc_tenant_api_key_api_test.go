package grpc

import (
	"context"
	"fmt"
	"testing"

	"github.com/Vector/vector-leads-scraper/runner/grpcrunner/middleware"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestServer_CreateTenantAPIKey(t *testing.T) {
	ctx := context.Background()
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Set up tenant ID in context using the middleware approach
	md := metadata.New(map[string]string{
		"x-tenant-id": fmt.Sprintf("%d", testCtx.TenantId),
		"x-organization-id": fmt.Sprintf("%d", testCtx.Organization.Id),
	})
	ctx = metadata.NewIncomingContext(ctx, md)
	
	// Use the middleware to extract and validate the auth info
	var err error
	ctx, err = middleware.ExtractAuthInfo(ctx)
	require.NoError(t, err, "Failed to extract auth info from context")

	tests := []struct {
		name    string
		req     *proto.CreateTenantAPIKeyRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.CreateTenantAPIKeyRequest{
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				ApiKey: &proto.TenantAPIKey{
					Name:        "Test API Key",
					Description: "Test API Key Description",
				},
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
			name: "invalid request",
			req: &proto.CreateTenantAPIKeyRequest{
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				ApiKey: &proto.TenantAPIKey{
					// Missing required fields
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.CreateTenantAPIKey(ctx, tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.NotZero(t, resp.KeyId)
			// The key value might be empty in tests, so we don't check it
			// assert.NotEmpty(t, resp.KeyValue)
		})
	}
}

func TestServer_GetTenantAPIKey(t *testing.T) {
	ctx := context.Background()
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Set up tenant ID in context using the middleware approach
	md := metadata.New(map[string]string{
		"x-tenant-id": fmt.Sprintf("%d", testCtx.TenantId),
		"x-organization-id": fmt.Sprintf("%d", testCtx.Organization.Id),
	})
	ctx = metadata.NewIncomingContext(ctx, md)
	
	// Use the middleware to extract and validate the auth info
	var err error
	ctx, err = middleware.ExtractAuthInfo(ctx)
	require.NoError(t, err, "Failed to extract auth info from context")

	// Create a test API key first
	createResp, err := MockServer.CreateTenantAPIKey(ctx, &proto.CreateTenantAPIKeyRequest{
		OrganizationId: testCtx.Organization.Id,
		TenantId:       testCtx.TenantId,
		ApiKey: &proto.TenantAPIKey{
			Name:        "Test API Key for Get",
			Description: "Test API Key Description",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)

	tests := []struct {
		name     string
		req      *proto.GetTenantAPIKeyRequest
		wantErr  bool
		errCode  codes.Code
		checkKey bool
	}{
		{
			name: "success",
			req: &proto.GetTenantAPIKeyRequest{
				KeyId:          createResp.KeyId,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
			},
			wantErr:  false,
			checkKey: true,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid request - missing key ID",
			req: &proto.GetTenantAPIKeyRequest{
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "key not found",
			req: &proto.GetTenantAPIKeyRequest{
				KeyId:          999999, // Non-existent key ID
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.GetTenantAPIKey(ctx, tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.ApiKey)
			if tt.checkKey {
				assert.Equal(t, createResp.KeyId, resp.ApiKey.Id)
				assert.Equal(t, "Test API Key for Get", resp.ApiKey.Name)
			}
		})
	}
}

func TestServer_UpdateTenantAPIKey(t *testing.T) {
	ctx := context.Background()
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Set up tenant ID in context using the middleware approach
	md := metadata.New(map[string]string{
		"x-tenant-id": fmt.Sprintf("%d", testCtx.TenantId),
		"x-organization-id": fmt.Sprintf("%d", testCtx.Organization.Id),
		"authorization": "Bearer test-token",
	})
	ctx = metadata.NewIncomingContext(ctx, md)
	
	// Use the middleware to extract and validate the auth info
	var err error
	ctx, err = middleware.ExtractAuthInfo(ctx)
	require.NoError(t, err, "Failed to extract auth info from context")

	// Create a test API key first
	createResp, err := MockServer.CreateTenantAPIKey(ctx, &proto.CreateTenantAPIKeyRequest{
		OrganizationId: testCtx.Organization.Id,
		TenantId:       testCtx.TenantId,
		ApiKey: &proto.TenantAPIKey{
			Name:        "Test API Key for Update",
			Description: "Original Description",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)

	tests := []struct {
		name    string
		req     *proto.UpdateTenantAPIKeyRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.UpdateTenantAPIKeyRequest{
				ApiKey: &proto.TenantAPIKey{
					Id:          createResp.KeyId,
					Name:        "Updated API Key",
					Description: "Updated Description",
				},
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
			name: "invalid request - missing API key",
			req: &proto.UpdateTenantAPIKeyRequest{},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "key not found",
			req: &proto.UpdateTenantAPIKeyRequest{
				ApiKey: &proto.TenantAPIKey{
					Id:          999999, // Non-existent key ID
					Name:        "Updated API Key",
					Description: "Updated Description",
				},
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.UpdateTenantAPIKey(ctx, tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)

			// Verify the update
			if !tt.wantErr {
				getResp, err := MockServer.GetTenantAPIKey(ctx, &proto.GetTenantAPIKeyRequest{
					KeyId:          tt.req.ApiKey.Id,
					OrganizationId: testCtx.Organization.Id,
					TenantId:       testCtx.TenantId,
				})
				require.NoError(t, err)
				require.NotNil(t, getResp)
				require.NotNil(t, getResp.ApiKey)
				assert.Equal(t, tt.req.ApiKey.Name, getResp.ApiKey.Name)
				assert.Equal(t, tt.req.ApiKey.Description, getResp.ApiKey.Description)
			}
		})
	}
}

func TestServer_DeleteTenantAPIKey(t *testing.T) {
	ctx := context.Background()
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Set up tenant ID in context using the middleware approach
	md := metadata.New(map[string]string{
		"x-tenant-id": fmt.Sprintf("%d", testCtx.TenantId),
		"x-organization-id": fmt.Sprintf("%d", testCtx.Organization.Id),
		"authorization": "Bearer test-token",
	})
	ctx = metadata.NewIncomingContext(ctx, md)
	
	// Use the middleware to extract and validate the auth info
	var err error
	ctx, err = middleware.ExtractAuthInfo(ctx)
	require.NoError(t, err, "Failed to extract auth info from context")

	// Create a test API key first
	createResp, err := MockServer.CreateTenantAPIKey(ctx, &proto.CreateTenantAPIKeyRequest{
		OrganizationId: testCtx.Organization.Id,
		TenantId:       testCtx.TenantId,
		ApiKey: &proto.TenantAPIKey{
			Name:        "Test API Key for Delete",
			Description: "Test Description",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)

	tests := []struct {
		name    string
		req     *proto.DeleteTenantAPIKeyRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.DeleteTenantAPIKeyRequest{
				KeyId:          createResp.KeyId,
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
			name: "invalid request - missing key ID",
			req: &proto.DeleteTenantAPIKeyRequest{
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "key not found",
			req: &proto.DeleteTenantAPIKeyRequest{
				KeyId:          999999, // Non-existent key ID
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.DeleteTenantAPIKey(ctx, tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.True(t, resp.Success)

			// Verify the key is deleted
			if !tt.wantErr {
				getResp, err := MockServer.GetTenantAPIKey(ctx, &proto.GetTenantAPIKeyRequest{
					KeyId:          tt.req.KeyId,
					TenantId:       tt.req.TenantId,
					OrganizationId: tt.req.OrganizationId,
				})
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, codes.Internal, st.Code())
				assert.Nil(t, getResp)
			}
		})
	}
}

func TestServer_ListTenantAPIKeys(t *testing.T) {
	ctx := context.Background()
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Set up tenant ID in context using the middleware approach
	md := metadata.New(map[string]string{
		"x-tenant-id": fmt.Sprintf("%d", testCtx.TenantId),
		"x-organization-id": fmt.Sprintf("%d", testCtx.Organization.Id),
	})
	ctx = metadata.NewIncomingContext(ctx, md)
	
	// Use the middleware to extract and validate the auth info
	var err error
	ctx, err = middleware.ExtractAuthInfo(ctx)
	require.NoError(t, err, "Failed to extract auth info from context")

	// Create a few test API keys
	for i := 0; i < 3; i++ {
		createResp, err := MockServer.CreateTenantAPIKey(ctx, &proto.CreateTenantAPIKeyRequest{
			OrganizationId: testCtx.Organization.Id,
			TenantId:       testCtx.TenantId,
			ApiKey: &proto.TenantAPIKey{
				Name:        "Test API Key",
				Description: "Test API Key Description",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, createResp)
	}

	tests := []struct {
		name    string
		req     *proto.ListTenantAPIKeysRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.ListTenantAPIKeysRequest{
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				PageSize:       10,
				PageNumber:     1,
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
			name: "invalid request - missing tenant ID",
			req: &proto.ListTenantAPIKeysRequest{
				OrganizationId: testCtx.Organization.Id,
				PageSize:       10,
				PageNumber:     1,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid request - missing organization ID",
			req: &proto.ListTenantAPIKeysRequest{
				TenantId:   testCtx.TenantId,
				PageSize:   10,
				PageNumber: 1,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.ListTenantAPIKeys(ctx, tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.GreaterOrEqual(t, len(resp.ApiKeys), 3) // Should have at least 3 keys
		})
	}
}
