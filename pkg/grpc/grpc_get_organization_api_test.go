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

func TestServer_GetOrganization(t *testing.T) {
	// Create a test organization first
	createResp, err := MockServer.CreateOrganization(context.Background(), &proto.CreateOrganizationRequest{
		Organization: testutils.GenerateRandomizedOrganization(),
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Organization)

	tests := []struct {
		name    string
		req     *proto.GetOrganizationRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.GetOrganizationRequest{
				Id: createResp.Organization.Id,
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
			name: "organization not found",
			req: &proto.GetOrganizationRequest{
				Id: 999999,
			},
			wantErr: true,
			errCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.GetOrganization(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Organization)
			assert.Equal(t, createResp.Organization.Id, resp.Organization.Id)
			assert.Equal(t, createResp.Organization.Name, resp.Organization.Name)
			assert.Equal(t, createResp.Organization.Description, resp.Organization.Description)
		})
	}
}
