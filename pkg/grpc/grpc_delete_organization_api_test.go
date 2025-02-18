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

func TestServer_DeleteOrganization(t *testing.T) {
	// Create a test organization first
	createResp, err := MockServer.CreateOrganization(context.Background(), &proto.CreateOrganizationRequest{
		Organization: testutils.GenerateRandomizedOrganization(),
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Organization)

	tests := []struct {
		name    string
		req     *proto.DeleteOrganizationRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.DeleteOrganizationRequest{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.DeleteOrganization(context.Background(), tt.req)
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

			// Verify organization is actually deleted
			getResp, err := MockServer.GetOrganization(context.Background(), &proto.GetOrganizationRequest{
				Id: tt.req.Id,
			})
			require.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.Internal, st.Code())
			assert.Nil(t, getResp)
		})
	}
}
