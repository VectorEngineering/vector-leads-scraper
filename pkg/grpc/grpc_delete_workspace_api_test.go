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

func TestServer_DeleteWorkspace(t *testing.T) {
	testCtx := initializeWorkspaceTestContext(t)
	defer testCtx.Cleanup()

	// Create a test workspace first
	createResp, err := MockServer.CreateWorkspace(context.Background(), &proto.CreateWorkspaceRequest{
		Workspace:      testutils.GenerateRandomWorkspace(),
		AccountId:      testCtx.Account.Id,
		OrganizationId: testCtx.Organization.Id,
		TenantId:       testCtx.TenantId,
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Workspace)

	tests := []struct {
		name    string
		req     *proto.DeleteWorkspaceRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.DeleteWorkspaceRequest{
				Id: createResp.Workspace.Id,
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
			name: "workspace not found",
			req: &proto.DeleteWorkspaceRequest{
				Id: 999999,
			},
			wantErr: true,
			errCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.DeleteWorkspace(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)

			// Verify workspace is actually deleted
			getResp, err := MockServer.GetWorkspace(context.Background(), &proto.GetWorkspaceRequest{
				Id:             tt.req.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
			})
			require.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.Internal, st.Code())
			assert.Nil(t, getResp)
		})
	}
}
