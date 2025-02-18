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

func TestServer_ListOrganizations(t *testing.T) {
	// Create multiple test organizations
	for i := 0; i < 3; i++ {
		createResp, err := MockServer.CreateOrganization(context.Background(), &proto.CreateOrganizationRequest{
			Organization: testutils.GenerateRandomizedOrganization(),
		})
		require.NoError(t, err)
		require.NotNil(t, createResp)
		require.NotNil(t, createResp.Organization)
	}

	tests := []struct {
		name    string
		req     *proto.ListOrganizationsRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success - first page",
			req: &proto.ListOrganizationsRequest{
				PageSize:   2,
				PageNumber: 1,
			},
			wantErr: false,
		},
		{
			name: "success - second page",
			req: &proto.ListOrganizationsRequest{
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
			req: &proto.ListOrganizationsRequest{
				PageSize:   -1,
				PageNumber: 0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.ListOrganizations(context.Background(), tt.req)
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
				assert.LessOrEqual(t, len(resp.Organizations), int(tt.req.PageSize))
			}

			for _, org := range resp.Organizations {
				assert.NotEmpty(t, org.Name)
				assert.NotEmpty(t, org.Description)
			}
		})
	}
}
