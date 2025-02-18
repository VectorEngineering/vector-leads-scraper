package grpc

import (
	"context"
	"testing"

	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_GetTenant(t *testing.T) {
	// Create a test tenant first
	tc := initializeTestContext(t)
	defer tc.Cleanup()

	// Create a test tenant
	createResp, err := MockServer.CreateTenant(context.Background(), &proto.CreateTenantRequest{
		Tenant: &proto.Tenant{
			Name:        "Test Tenant",
			Description: "A test tenant",
			ApiBaseUrl:  "https://api.example.com",
		},
		OrganizationId: tc.Organization.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotZero(t, createResp.TenantId)

	tests := []struct {
		name    string
		req     *proto.GetTenantRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.GetTenantRequest{
				TenantId:       createResp.TenantId,
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
				TenantId:       0,
				OrganizationId: tc.Organization.Id,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "zero organization ID",
			req: &proto.GetTenantRequest{
				TenantId:       createResp.TenantId,
				OrganizationId: 0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "tenant not found",
			req: &proto.GetTenantRequest{
				TenantId:       999999,
				OrganizationId: tc.Organization.Id,
			},
			wantErr: true,
			errCode: codes.NotFound,
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
			require.NotNil(t, resp.Tenant)
			assert.Equal(t, createResp.TenantId, resp.Tenant.Id)
		})
	}
}
