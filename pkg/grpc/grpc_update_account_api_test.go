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

	tests := []struct {
		name    string
		req     *proto.UpdateAccountRequest
		wantErr bool
		errCode codes.Code
		setup   func() *proto.Account // Modified to return the account to use
	}{
		{
			name: "success - update email",
			setup: func() *proto.Account {
				// Create a new account for this test
				createResp, err := MockServer.CreateAccount(context.Background(), &proto.CreateAccountRequest{
					Account:              testutils.GenerateRandomizedAccount(),
					OrganizationId:       tc.Organization.Id,
					TenantId:             tc.TenantId,
					InitialWorkspaceName: "Test Workspace Email",
				})
				require.NoError(t, err)
				require.NotNil(t, createResp)
				require.NotNil(t, createResp.Account)
				return createResp.Account
			},
			wantErr: false,
		},
		{
			name: "success - update monthly job limit",
			setup: func() *proto.Account {
				// Create a new account for this test
				createResp, err := MockServer.CreateAccount(context.Background(), &proto.CreateAccountRequest{
					Account:              testutils.GenerateRandomizedAccount(),
					OrganizationId:       tc.Organization.Id,
					TenantId:             tc.TenantId,
					InitialWorkspaceName: "Test Workspace Monthly",
				})
				require.NoError(t, err)
				require.NotNil(t, createResp)
				require.NotNil(t, createResp.Account)
				return createResp.Account
			},
			wantErr: false,
		},
		{
			name: "success - update multiple fields",
			setup: func() *proto.Account {
				// Create a new account for this test
				createResp, err := MockServer.CreateAccount(context.Background(), &proto.CreateAccountRequest{
					Account:              testutils.GenerateRandomizedAccount(),
					OrganizationId:       tc.Organization.Id,
					TenantId:             tc.TenantId,
					InitialWorkspaceName: "Test Workspace Multiple",
				})
				require.NoError(t, err)
				require.NotNil(t, createResp)
				require.NotNil(t, createResp.Account)
				return createResp.Account
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
			name: "nil payload",
			req: &proto.UpdateAccountRequest{
				Payload: nil,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "nil account",
			req: &proto.UpdateAccountRequest{
				Payload: &proto.UpdateAccountRequestPayload{
					OrganizationId: tc.Organization.Id,
					TenantId:       tc.TenantId,
					Account:        nil,
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing organization ID",
			req: &proto.UpdateAccountRequest{
				Payload: &proto.UpdateAccountRequestPayload{
					TenantId: tc.TenantId,
					Account: &proto.Account{
						Id:    999, // Will be replaced with testAccount.Id
						Email: "missing-org-id@example.com",
					},
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "account not found",
			req: &proto.UpdateAccountRequest{
				Payload: &proto.UpdateAccountRequestPayload{
					OrganizationId: tc.Organization.Id,
					TenantId:       tc.TenantId,
					Account: &proto.Account{
						Id:                 999999, // Non-existent account ID
						Email:              "non-existent@example.com",
						AuthPlatformUserId: "auth0|123456789abcdef",
						MonthlyJobLimit:    10,
						ConcurrentJobLimit: 2,
					},
				},
			},
			wantErr: true,
			errCode: codes.Internal,
		},
		{
			name: "invalid request - missing account ID",
			req: &proto.UpdateAccountRequest{
				Payload: &proto.UpdateAccountRequestPayload{
					OrganizationId: tc.Organization.Id,
					TenantId:       tc.TenantId,
					Account: &proto.Account{
						Email: "missing-id@example.com",
					},
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "non-existent organization",
			req: &proto.UpdateAccountRequest{
				Payload: &proto.UpdateAccountRequestPayload{
					OrganizationId: 999999, // Non-existent org ID
					TenantId:       tc.TenantId,
					Account: &proto.Account{
						Id:                 999, // Will be replaced with testAccount.Id
						Email:              "non-existent-org@example.com",
						AuthPlatformUserId: "auth0|123456789abcdef",
						MonthlyJobLimit:    10,
						ConcurrentJobLimit: 2,
					},
				},
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "non-existent tenant",
			req: &proto.UpdateAccountRequest{
				Payload: &proto.UpdateAccountRequestPayload{
					OrganizationId: tc.Organization.Id,
					TenantId:       999999, // Non-existent tenant ID
					Account: &proto.Account{
						Id:                 999, // Will be replaced with testAccount.Id
						Email:              "non-existent-tenant@example.com",
						AuthPlatformUserId: "auth0|123456789abcdef",
						MonthlyJobLimit:    10,
						ConcurrentJobLimit: 2,
					},
				},
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testAccount *proto.Account
			if tt.setup != nil {
				testAccount = tt.setup()
			} else {
				testAccount = account // Use the default account for error cases
			}

			// Build the request based on the test case
			var req *proto.UpdateAccountRequest
			if tt.req != nil {
				req = tt.req
				// If the test case has a payload with an account, update it with the test account
				if req.Payload != nil && req.Payload.Account != nil && req.Payload.Account.Id != 0 && req.Payload.Account.Id != 999999 {
					// Update the account ID to use the test account
					req.Payload.Account.Id = testAccount.Id
				}
			} else if !tt.wantErr {
				// For success cases, create a new request
				if tt.name == "success - update email" {
					req = &proto.UpdateAccountRequest{
						Payload: &proto.UpdateAccountRequestPayload{
							OrganizationId: tc.Organization.Id,
							TenantId:       tc.TenantId,
							Account: &proto.Account{
								Id:                 testAccount.Id,
								Email:              "updated-email@example.com",
								AuthPlatformUserId: testAccount.AuthPlatformUserId,
								MonthlyJobLimit:    testAccount.MonthlyJobLimit,
								ConcurrentJobLimit: testAccount.ConcurrentJobLimit,
							},
						},
					}
				} else if tt.name == "success - update monthly job limit" {
					req = &proto.UpdateAccountRequest{
						Payload: &proto.UpdateAccountRequestPayload{
							OrganizationId: tc.Organization.Id,
							TenantId:       tc.TenantId,
							Account: &proto.Account{
								Id:                 testAccount.Id,
								Email:              testAccount.Email,
								AuthPlatformUserId: testAccount.AuthPlatformUserId,
								MonthlyJobLimit:    50,
								ConcurrentJobLimit: testAccount.ConcurrentJobLimit,
							},
						},
					}
				} else if tt.name == "success - update multiple fields" {
					req = &proto.UpdateAccountRequest{
						Payload: &proto.UpdateAccountRequestPayload{
							OrganizationId: tc.Organization.Id,
							TenantId:       tc.TenantId,
							Account: &proto.Account{
								Id:                 testAccount.Id,
								Email:              "multiple-update@example.com",
								AuthPlatformUserId: testAccount.AuthPlatformUserId,
								MonthlyJobLimit:    75,
								ConcurrentJobLimit: 5,
							},
						},
					}
				}
			}

			resp, err := MockServer.UpdateAccount(context.Background(), req)
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

			// Debug: Print the response account details
			t.Logf("Response account: ID=%d, Email=%s, MonthlyJobLimit=%d, ConcurrentJobLimit=%d",
				resp.Account.Id, resp.Account.Email, resp.Account.MonthlyJobLimit, resp.Account.ConcurrentJobLimit)

			// Debug: Print the request account details
			t.Logf("Request account: ID=%d, Email=%s, MonthlyJobLimit=%d, ConcurrentJobLimit=%d",
				req.Payload.Account.Id, req.Payload.Account.Email, req.Payload.Account.MonthlyJobLimit, req.Payload.Account.ConcurrentJobLimit)

			// Verify the update was applied correctly in the response
			if req.Payload.Account.Email != "" {
				assert.Equal(t, req.Payload.Account.Email, resp.Account.Email, "Email in response doesn't match request")
			}
			if req.Payload.Account.MonthlyJobLimit != 0 {
				assert.Equal(t, req.Payload.Account.MonthlyJobLimit, resp.Account.MonthlyJobLimit, "MonthlyJobLimit in response doesn't match request")
			}
			if req.Payload.Account.ConcurrentJobLimit != 0 {
				assert.Equal(t, req.Payload.Account.ConcurrentJobLimit, resp.Account.ConcurrentJobLimit, "ConcurrentJobLimit in response doesn't match request")
			}

			// Verify the account ID remains unchanged
			assert.Equal(t, req.Payload.Account.Id, resp.Account.Id)

			// Get the updated account
			getResp, err := MockServer.GetAccount(context.Background(), &proto.GetAccountRequest{
				Id:             resp.Account.Id,
				OrganizationId: req.Payload.OrganizationId,
				TenantId:       req.Payload.TenantId,
			})
			require.NoError(t, err)
			require.NotNil(t, getResp)
			require.NotNil(t, getResp.Account)

			retrievedAccount := getResp.Account
			t.Logf("Retrieved account: ID=%d, Email=%s, MonthlyJobLimit=%d, ConcurrentJobLimit=%d",
				retrievedAccount.Id, retrievedAccount.Email, retrievedAccount.MonthlyJobLimit, retrievedAccount.ConcurrentJobLimit)

			// Verify the account was updated correctly
			assert.Equal(t, resp.Account.Id, retrievedAccount.Id)
			assert.Equal(t, resp.Account.Email, retrievedAccount.Email)
			assert.Equal(t, resp.Account.MonthlyJobLimit, retrievedAccount.MonthlyJobLimit)
			assert.Equal(t, resp.Account.ConcurrentJobLimit, retrievedAccount.ConcurrentJobLimit)
		})
	}
}
