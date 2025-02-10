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

func TestServer_CreateOrganization(t *testing.T) {
	tests := []struct {
		name    string
		req     *proto.CreateOrganizationRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.CreateOrganizationRequest{
				Organization: testutils.GenerateRandomizedOrganization(),
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
			name: "empty organization name",
			req: &proto.CreateOrganizationRequest{
				Organization: &proto.Organization{
					Name:        "",
					Description: "A test organization",
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.CreateOrganization(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

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
