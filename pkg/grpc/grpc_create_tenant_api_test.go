package grpc

import (
	"context"
	"testing"

	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
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
				Tenant: &proto.Tenant{
					Name:        "Test Tenant",
					Description: "A test tenant",
					ApiBaseUrl:  "https://api.example.com",
				},
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
				OrganizationId: tc.Organization.Id,
				Tenant:         nil,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "empty tenant name",
			req: &proto.CreateTenantRequest{
				OrganizationId: tc.Organization.Id,
				Tenant: &proto.Tenant{
					Name:        "",
					Description: "A test tenant",
					ApiBaseUrl:  "https://api.example.com",
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing organization id",
			req: &proto.CreateTenantRequest{
				Tenant: &proto.Tenant{
					Name:        "Test Tenant",
					Description: "A test tenant",
					ApiBaseUrl:  "https://api.example.com",
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
				require.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotZero(t, resp.TenantId)
		})
	}
}
