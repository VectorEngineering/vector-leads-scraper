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

func TestServer_UpdateOrganization(t *testing.T) {
	const testOrgName = "Updated Test Organization"
	sampleOrg := testutils.GenerateRandomizedOrganization()

	// Create a test organization first
	createResp, err := MockServer.CreateOrganization(context.Background(), &proto.CreateOrganizationRequest{
		Organization: sampleOrg,
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Organization)

	createdOrg := createResp.Organization
	// Update the organization name
	createdOrg.Name = testOrgName

	// sample org id should be updated
	createdOrg.Id = sampleOrg.Id

	tests := []struct {
		name    string
		req     *proto.UpdateOrganizationRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.UpdateOrganizationRequest{
				Organization: createdOrg,
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
			name: "nil organization",
			req: &proto.UpdateOrganizationRequest{
				Organization: nil,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.UpdateOrganization(context.Background(), tt.req)
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
			require.Equal(t, tt.req.Organization.Name, resp.Organization.Name)
		})
	}
}
