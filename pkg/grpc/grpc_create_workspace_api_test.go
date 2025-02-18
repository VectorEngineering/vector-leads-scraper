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

func TestServer_CreateWorkspace(t *testing.T) {
	testCtx := initializeWorkspaceTestContext(t)
	defer testCtx.Cleanup()

	tests := []struct {
		name    string
		req     *proto.CreateWorkspaceRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.CreateWorkspaceRequest{
				Workspace:      testutils.GenerateRandomWorkspace(),
				AccountId:      testCtx.Account.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
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
			name: "invalid account id",
			req: &proto.CreateWorkspaceRequest{
				Workspace:      testutils.GenerateRandomWorkspace(),
				AccountId:      0,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid organization id",
			req: &proto.CreateWorkspaceRequest{
				Workspace:      testutils.GenerateRandomWorkspace(),
				AccountId:      testCtx.Account.Id,
				OrganizationId: 0,
				TenantId:       testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid tenant id",
			req: &proto.CreateWorkspaceRequest{
				Workspace:      testutils.GenerateRandomWorkspace(),
				AccountId:      testCtx.Account.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.CreateWorkspace(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Workspace)
			assert.NotEmpty(t, resp.Workspace.Id)
			assert.Equal(t, tt.req.Workspace.Name, resp.Workspace.Name)
		})
	}
}
