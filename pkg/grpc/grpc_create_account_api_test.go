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

	// Create a valid account for duplicate email test
	validAccount := testutils.GenerateRandomizedAccount()
	createResp, err := MockServer.CreateAccount(context.Background(), &proto.CreateAccountRequest{
		Account:              validAccount,
		OrganizationId:       testCtx.Organization.Id,
		TenantId:             testCtx.TenantId,
		InitialWorkspaceName: "Test Workspace",
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Account)

	tests := []struct {
		name    string
		req     *proto.CreateAccountRequest
		wantErr bool
		errCode codes.Code
		setup   func() // Optional setup function
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
			name: "success with custom monthly job limit",
			req: &proto.CreateAccountRequest{
				Account: &proto.Account{
					Email:              testutils.GenerateRandomEmail(10),
					AuthPlatformUserId: "auth0|" + testutils.GenerateRandomString(24, true, false),
					MonthlyJobLimit:    50,
					ConcurrentJobLimit: 5,
					AccountStatus:      proto.Account_ACCOUNT_STATUS_ACTIVE,
				},
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
			name: "nil account",
			req: &proto.CreateAccountRequest{
				Account:              nil,
				OrganizationId:       testCtx.Organization.Id,
				TenantId:             testCtx.TenantId,
				InitialWorkspaceName: "Test Workspace",
			},
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
		{
			name: "missing initial workspace name",
			req: &proto.CreateAccountRequest{
				Account:        testutils.GenerateRandomizedAccount(),
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing email",
			req: &proto.CreateAccountRequest{
				Account: &proto.Account{
					AuthPlatformUserId: "auth0|" + testutils.GenerateRandomString(24, true, false),
					MonthlyJobLimit:    10,
					ConcurrentJobLimit: 2,
					AccountStatus:      proto.Account_ACCOUNT_STATUS_ACTIVE,
				},
				OrganizationId:       testCtx.Organization.Id,
				TenantId:             testCtx.TenantId,
				InitialWorkspaceName: "Test Workspace",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing auth platform user id",
			req: &proto.CreateAccountRequest{
				Account: &proto.Account{
					Email:              testutils.GenerateRandomEmail(10),
					MonthlyJobLimit:    10,
					ConcurrentJobLimit: 2,
					AccountStatus:      proto.Account_ACCOUNT_STATUS_ACTIVE,
				},
				OrganizationId:       testCtx.Organization.Id,
				TenantId:             testCtx.TenantId,
				InitialWorkspaceName: "Test Workspace",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid monthly job limit",
			req: &proto.CreateAccountRequest{
				Account: &proto.Account{
					Email:              testutils.GenerateRandomEmail(10),
					AuthPlatformUserId: "auth0|" + testutils.GenerateRandomString(24, true, false),
					MonthlyJobLimit:    -1, // Invalid: must be > 0
					ConcurrentJobLimit: 2,
					AccountStatus:      proto.Account_ACCOUNT_STATUS_ACTIVE,
				},
				OrganizationId:       testCtx.Organization.Id,
				TenantId:             testCtx.TenantId,
				InitialWorkspaceName: "Test Workspace",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid concurrent job limit",
			req: &proto.CreateAccountRequest{
				Account: &proto.Account{
					Email:              testutils.GenerateRandomEmail(10),
					AuthPlatformUserId: "auth0|" + testutils.GenerateRandomString(24, true, false),
					MonthlyJobLimit:    10,
					ConcurrentJobLimit: -1, // Invalid: must be > 0
					AccountStatus:      proto.Account_ACCOUNT_STATUS_ACTIVE,
				},
				OrganizationId:       testCtx.Organization.Id,
				TenantId:             testCtx.TenantId,
				InitialWorkspaceName: "Test Workspace",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "duplicate email",
			req: &proto.CreateAccountRequest{
				Account: &proto.Account{
					Email:              validAccount.Email, // Using the same email as the already created account
					AuthPlatformUserId: "auth0|" + testutils.GenerateRandomString(24, true, false),
					MonthlyJobLimit:    10,
					ConcurrentJobLimit: 2,
					AccountStatus:      proto.Account_ACCOUNT_STATUS_ACTIVE,
				},
				OrganizationId:       testCtx.Organization.Id,
				TenantId:             testCtx.TenantId,
				InitialWorkspaceName: "Test Workspace",
			},
			wantErr: true,
			errCode: codes.AlreadyExists,
		},
		{
			name: "non-existent organization",
			req: &proto.CreateAccountRequest{
				Account:              testutils.GenerateRandomizedAccount(),
				OrganizationId:       999999, // Non-existent org ID
				TenantId:             testCtx.TenantId,
				InitialWorkspaceName: "Test Workspace",
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "non-existent tenant",
			req: &proto.CreateAccountRequest{
				Account:              testutils.GenerateRandomizedAccount(),
				OrganizationId:       testCtx.Organization.Id,
				TenantId:             999999, // Non-existent tenant ID
				InitialWorkspaceName: "Test Workspace",
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

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

			// Verify the account was created with the correct properties
			assert.NotZero(t, resp.Account.Id)
			if tt.req.Account.Email != "" {
				assert.Equal(t, tt.req.Account.Email, resp.Account.Email)
			}
			if tt.req.Account.AuthPlatformUserId != "" {
				assert.Equal(t, tt.req.Account.AuthPlatformUserId, resp.Account.AuthPlatformUserId)
			}
			if tt.req.Account.MonthlyJobLimit != 0 {
				assert.Equal(t, tt.req.Account.MonthlyJobLimit, resp.Account.MonthlyJobLimit)
			}
			if tt.req.Account.ConcurrentJobLimit != 0 {
				assert.Equal(t, tt.req.Account.ConcurrentJobLimit, resp.Account.ConcurrentJobLimit)
			}

			// Verify the workspace was created
			assert.NotEmpty(t, resp.Account.Workspaces)
			assert.Equal(t, tt.req.InitialWorkspaceName, resp.Account.Workspaces[0].Name)

			// Verify the account can be retrieved
			getResp, err := MockServer.GetAccount(context.Background(), &proto.GetAccountRequest{
				Id:             resp.Account.Id,
				OrganizationId: tt.req.OrganizationId,
				TenantId:       tt.req.TenantId,
			})
			require.NoError(t, err)
			require.NotNil(t, getResp)
			require.NotNil(t, getResp.Account)
			assert.Equal(t, resp.Account.Id, getResp.Account.Id)
		})
	}
}
