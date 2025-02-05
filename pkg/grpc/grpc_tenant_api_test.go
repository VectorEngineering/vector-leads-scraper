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

func TestServer_CreateTenant(t *testing.T) {
	tc := initializeTestContext(t)
	defer tc.Cleanup()

	tests := []struct {
		name    string
		req     *proto.CreateTenantRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.CreateTenantRequest{
				Tenant: testutils.GenerateRandomizedTenant(),
				OrganizationId: tc.Organization.Id,
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
			name: "nil tenant",
			req: &proto.CreateTenantRequest{
				Tenant: nil,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "empty tenant name",
			req: &proto.CreateTenantRequest{
				Tenant: &proto.Tenant{
					Name:        "",
					Description: "A test tenant",
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.CreateTenant(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.NotZero(t, resp.TenantId)
		})
	}
}

func TestServer_GetTenant(t *testing.T) {
	// Create a test tenant first
	tc := initializeTestContext(t)
	defer tc.Cleanup()

	tests := []struct {
		name    string
		req     *proto.GetTenantRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.GetTenantRequest{
				TenantId: tc.TenantId,
				OrganizationId: tc.Organization.Id,
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
			name: "zero tenant ID",
			req: &proto.GetTenantRequest{
				TenantId: 0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.GetTenant(context.Background(), tt.req)
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

func TestServer_UpdateTenant(t *testing.T) {
	// Create a test tenant first
	tc := initializeTestContext(t)
	defer tc.Cleanup()

	createResp, err := MockServer.CreateTenant(context.Background(), &proto.CreateTenantRequest{
		Tenant: testutils.GenerateRandomizedTenant(),
		OrganizationId: tc.Organization.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotZero(t, createResp.TenantId)

	tests := []struct {
		name    string
		req     *proto.UpdateTenantRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.UpdateTenantRequest{
				Tenant: &proto.Tenant{
					Id:          createResp.TenantId,
					Name:        "Updated Tenant",
					Description: "An updated test tenant",
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
			name: "nil tenant",
			req: &proto.UpdateTenantRequest{
				Tenant: nil,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "tenant not found",
			req: &proto.UpdateTenantRequest{
				Tenant: &proto.Tenant{
					Id:          999999,
					Name:        "Updated Tenant",
					Description: "An updated test tenant",
				},
			},
			wantErr: true,
			errCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.UpdateTenant(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Tenant)
			assert.Equal(t, tt.req.Tenant.Id, resp.Tenant.Id)
			assert.Equal(t, tt.req.Tenant.Name, resp.Tenant.Name)
			assert.Equal(t, tt.req.Tenant.Description, resp.Tenant.Description)
		})
	}
}

func TestServer_DeleteTenant(t *testing.T) {
	// Create a test tenant first
	createResp, err := MockServer.CreateTenant(context.Background(), &proto.CreateTenantRequest{
		Tenant: &proto.Tenant{
			Name:        "Test Tenant",
			Description: "A test tenant",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotZero(t, createResp.TenantId)

	tests := []struct {
		name    string
		req     *proto.DeleteTenantRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.DeleteTenantRequest{
				TenantId: createResp.TenantId,
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
			name: "tenant not found",
			req: &proto.DeleteTenantRequest{
				TenantId: 999999,
			},
			wantErr: true,
			errCode: codes.Internal,
		},
		{
			name: "zero tenant ID",
			req: &proto.DeleteTenantRequest{
				TenantId: 0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.DeleteTenant(context.Background(), tt.req)
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

			// Verify tenant is actually deleted
			getResp, err := MockServer.GetTenant(context.Background(), &proto.GetTenantRequest{
				TenantId: tt.req.TenantId,
			})
			require.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.Internal, st.Code())
			assert.Nil(t, getResp)
		})
	}
}

func TestServer_ListTenants(t *testing.T) {
	// Create multiple test tenants
	for i := 0; i < 3; i++ {
		createResp, err := MockServer.CreateTenant(context.Background(), &proto.CreateTenantRequest{
			Tenant: &proto.Tenant{
				Name:        testutils.GenerateRandomString(10, false, false),
				Description: testutils.GenerateRandomString(20, true, false),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, createResp)
		require.NotZero(t, createResp.TenantId)
	}

	tests := []struct {
		name    string
		req     *proto.ListTenantsRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success - first page",
			req: &proto.ListTenantsRequest{
				PageSize:   2,
				PageNumber: 0,
			},
			wantErr: false,
		},
		{
			name: "success - second page",
			req: &proto.ListTenantsRequest{
				PageSize:   2,
				PageNumber: 1,
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
			name: "invalid page size",
			req: &proto.ListTenantsRequest{
				PageSize:   -1,
				PageNumber: 0,
			},
			wantErr: false, // Should use default page size
		},
		{
			name: "filter by organization ID",
			req: &proto.ListTenantsRequest{
				PageSize:       10,
				PageNumber:     0,
				OrganizationId: 123, // Using a test org ID
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.ListTenants(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)

			if tt.req != nil && tt.req.PageSize > 0 {
				assert.LessOrEqual(t, len(resp.Tenants), int(tt.req.PageSize))
			}

			for _, tenant := range resp.Tenants {
				assert.NotEmpty(t, tenant.Name)
				assert.NotEmpty(t, tenant.Description)
				assert.NotZero(t, tenant.Id)
			}
		})
	}
} 