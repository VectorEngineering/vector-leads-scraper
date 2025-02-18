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

func TestServer_GetAccount(t *testing.T) {
	// Create a test account first
	testCtx := initializeTestContext(t)
	defer testCtx.Cleanup()

	// create an account first
	createResp, err := MockServer.CreateAccount(context.Background(), &proto.CreateAccountRequest{
		Account:              testutils.GenerateRandomizedAccount(),
		OrganizationId:       testCtx.Organization.Id,
		TenantId:             testCtx.TenantId,
		InitialWorkspaceName: "Test Workspace",
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Account)

	tests := []struct {
		name    string
		req     *proto.GetAccountRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.GetAccountRequest{
				Id:             createResp.Account.Id,
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
			name: "account not found",
			req: &proto.GetAccountRequest{
				Id:             999999,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "invalid organization id",
			req: &proto.GetAccountRequest{
				Id:             createResp.Account.Id,
				OrganizationId: 0,
				TenantId:       testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid tenant id",
			req: &proto.GetAccountRequest{
				Id:             createResp.Account.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.GetAccount(context.Background(), tt.req)
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
			assert.Equal(t, createResp.Account.Id, resp.Account.Id)
			assert.Equal(t, createResp.Account.Email, resp.Account.Email)
		})
	}
}
