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

func TestServer_DeleteAccount(t *testing.T) {
	tc := initializeTestContext(t)
	defer tc.Cleanup()

	// Create an account first
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

	// Create a second account for additional tests
	createResp2, err := MockServer.CreateAccount(context.Background(), &proto.CreateAccountRequest{
		Account:              testutils.GenerateRandomizedAccount(),
		OrganizationId:       tc.Organization.Id,
		TenantId:             tc.TenantId,
		InitialWorkspaceName: "Test Workspace 2",
	})
	require.NoError(t, err)
	require.NotNil(t, createResp2)
	require.NotNil(t, createResp2.Account)

	account2 := createResp2.Account

	tests := []struct {
		name    string
		req     *proto.DeleteAccountRequest
		wantErr bool
		errCode codes.Code
		setup   func() // Optional setup function
	}{
		{
			name: "success",
			req: &proto.DeleteAccountRequest{
				Id:             account.Id,
				OrganizationId: tc.Organization.Id,
				TenantId:       tc.TenantId,
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
			name: "missing account ID",
			req: &proto.DeleteAccountRequest{
				OrganizationId: tc.Organization.Id,
				TenantId:       tc.TenantId,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing organization ID",
			req: &proto.DeleteAccountRequest{
				Id:       account2.Id,
				TenantId: tc.TenantId,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing tenant ID",
			req: &proto.DeleteAccountRequest{
				Id:             account2.Id,
				OrganizationId: tc.Organization.Id,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "account not found",
			req: &proto.DeleteAccountRequest{
				Id:             999999, // Non-existent account ID
				OrganizationId: tc.Organization.Id,
				TenantId:       tc.TenantId,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "non-existent organization",
			req: &proto.DeleteAccountRequest{
				Id:             account2.Id,
				OrganizationId: 999999, // Non-existent org ID
				TenantId:       tc.TenantId,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "non-existent tenant",
			req: &proto.DeleteAccountRequest{
				Id:             account2.Id,
				OrganizationId: tc.Organization.Id,
				TenantId:       999999, // Non-existent tenant ID
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "already deleted account",
			req: &proto.DeleteAccountRequest{
				Id:             account.Id, // Will be deleted in the first test case
				OrganizationId: tc.Organization.Id,
				TenantId:       tc.TenantId,
			},
			wantErr: true,
			errCode: codes.NotFound,
			setup: func() {
				// Ensure the first account is deleted before this test runs
				_, err := MockServer.DeleteAccount(context.Background(), &proto.DeleteAccountRequest{
					Id:             account.Id,
					OrganizationId: tc.Organization.Id,
					TenantId:       tc.TenantId,
				})
				
				// Don't fail the test if this setup call fails - the actual test will handle it
				// Just log the error if there is one
				if err != nil {
					t.Logf("Setup for already_deleted_account test: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			resp, err := MockServer.DeleteAccount(context.Background(), tt.req)
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

			// Verify the account is deleted by trying to get it
			getResp, err := MockServer.GetAccount(context.Background(), &proto.GetAccountRequest{
				Id:             tt.req.Id,
				OrganizationId: tt.req.OrganizationId,
				TenantId:       tt.req.TenantId,
			})
			require.Error(t, err, "Expected error when getting deleted account")
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.NotFound, st.Code(), "Expected NotFound error when getting deleted account")
			assert.Nil(t, getResp)
		})
	}
}
