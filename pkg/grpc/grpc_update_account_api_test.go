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

func TestServer_UpdateAccount(t *testing.T) {
	tc := initializeTestContext(t)
	defer tc.Cleanup()

	// create an account first
	createResp, err := MockServer.CreateAccount(context.Background(), &proto.CreateAccountRequest{
		Account:              testutils.GenerateRandomizedAccount(),
		OrganizationId:       tc.Organization.Id,
		TenantId:             tc.TenantId,
		InitialWorkspaceName: "Test Workspace",
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Account)

	account := createResp.Account

	tests := []struct {
		name    string
		req     *proto.UpdateAccountRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.UpdateAccountRequest{
				Payload: &proto.UpdateAccountRequestPayload{
					OrganizationId: tc.Organization.Id,
					TenantId:       tc.TenantId,
					Account:        account,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.UpdateAccount(context.Background(), tt.req)
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
