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

func TestServer_DeleteTenant(t *testing.T) {
	// Create test organization first
	org := testutils.GenerateRandomizedOrganization()
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

	// Create a test tenant
	createTenantResp, err := MockServer.CreateTenant(context.Background(), &proto.CreateTenantRequest{
		Tenant: &proto.Tenant{
			Name:        "Test Tenant",
			Description: "A test tenant",
			ApiBaseUrl:  "https://api.example.com",
		},
		OrganizationId: createOrgResp.Organization.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createTenantResp)
	require.NotZero(t, createTenantResp.TenantId)

	tests := []struct {
		name    string
		req     *proto.DeleteTenantRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.DeleteTenantRequest{
				TenantId:       createTenantResp.TenantId,
				OrganizationId: createOrgResp.Organization.Id,
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
				TenantId:       999999,
				OrganizationId: createOrgResp.Organization.Id,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "zero tenant ID",
			req: &proto.DeleteTenantRequest{
				TenantId:       0,
				OrganizationId: createOrgResp.Organization.Id,
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

			// Verify tenant is deleted
			_, err = MockServer.GetTenant(context.Background(), &proto.GetTenantRequest{
				TenantId:       tt.req.TenantId,
				OrganizationId: tt.req.OrganizationId,
			})
			require.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.NotFound, st.Code())
		})
	}

	// Cleanup
	MockServer.DeleteOrganization(context.Background(), &proto.DeleteOrganizationRequest{
		Id: createOrgResp.Organization.Id,
	})
}
