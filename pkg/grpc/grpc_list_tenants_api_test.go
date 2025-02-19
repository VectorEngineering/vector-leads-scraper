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
		name             string
		req              *proto.ListTenantsRequest
		wantErr          bool
		errCode          codes.Code
		expectedCount    int
		expectedNextPage int32
	}{
		{
			name: "success - first page",
			req: &proto.ListTenantsRequest{
				PageSize:       2,
				PageNumber:     1,
				OrganizationId: tc.Organization.Id,
			},
			wantErr:          false,
			expectedCount:    2,
			expectedNextPage: 2,
		},
		{
			name: "success - second page",
			req: &proto.ListTenantsRequest{
				PageSize:       2,
				PageNumber:     2,
				OrganizationId: tc.Organization.Id,
			},
			wantErr:          false,
			expectedCount:    2,
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
