package grpc

import (
	"context"
	"fmt"
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

func TestServer_ListTenants(t *testing.T) {
	tc := initializeTestContext(t)
	defer tc.Cleanup()

	// Create 3 test tenants
	var tenantIDs []uint64
	for i := 0; i < 3; i++ {
		tenant := testutils.GenerateRandomizedTenant()
		tenant.Name = fmt.Sprintf("Test Tenant %d", i)
		tenant.Description = fmt.Sprintf("Test tenant description %d", i)
		
		createResp, err := MockServer.CreateTenant(context.Background(), &proto.CreateTenantRequest{
			Tenant:         tenant,
			OrganizationId: tc.Organization.Id,
		})
		require.NoError(t, err)
		require.NotNil(t, createResp)
		require.NotZero(t, createResp.TenantId)
		tenantIDs = append(tenantIDs, createResp.TenantId)
	}

	tests := []struct {
		name           string
		req            *proto.ListTenantsRequest
		wantErr        bool
		errCode        codes.Code
		expectedCount  int
		expectedNextPage int32
	}{
		{
			name: "success - first page",
			req: &proto.ListTenantsRequest{
				PageSize:       2,
				PageNumber:     1,
				OrganizationId: tc.Organization.Id,
			},
			wantErr:        false,
			expectedCount:  2,
			expectedNextPage: 2,
		},
		{
			name: "success - second page",
			req: &proto.ListTenantsRequest{
				PageSize:       2,
				PageNumber:     2,
				OrganizationId: tc.Organization.Id,
			},
			wantErr:        false,
			expectedCount:  2,
			expectedNextPage: 3,
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
				PageSize:   0,
				PageNumber: 1,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
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
			assert.Len(t, resp.Tenants, tt.expectedCount)
			assert.Equal(t, tt.expectedNextPage, resp.NextPageNumber)
		})
	}
} 