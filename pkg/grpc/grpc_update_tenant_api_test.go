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

func TestServer_UpdateTenant(t *testing.T) {
	// Create a test tenant first
	tc := initializeTestContext(t)
	defer tc.Cleanup()

	createResp, err := MockServer.CreateTenant(context.Background(), &proto.CreateTenantRequest{
		Tenant:         testutils.GenerateRandomizedTenant(),
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
