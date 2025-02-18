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

func TestServer_CreateAccount(t *testing.T) {
	// create an organization and tenant first
	testCtx := initializeTestContext(t)
	defer testCtx.Cleanup()

	tests := []struct {
		name    string
		req     *proto.CreateAccountRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.CreateAccountRequest{
				Account:              testutils.GenerateRandomizedAccount(),
				OrganizationId:       testCtx.Organization.Id,
				TenantId:             testCtx.TenantId,
				InitialWorkspaceName: "Test Workspace",
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
			name: "invalid organization id",
			req: &proto.CreateAccountRequest{
				Account:              testutils.GenerateRandomizedAccount(),
				OrganizationId:       0,
				TenantId:             testCtx.TenantId,
				InitialWorkspaceName: "Test Workspace",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid tenant id",
			req: &proto.CreateAccountRequest{
				Account:              testutils.GenerateRandomizedAccount(),
				OrganizationId:       testCtx.Organization.Id,
				TenantId:             0,
				InitialWorkspaceName: "Test Workspace",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.CreateAccount(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Account)
		})
	}
}
