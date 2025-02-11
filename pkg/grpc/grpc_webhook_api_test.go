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

type webhookTestContext struct {
	Organization *proto.Organization
	TenantId     uint64
	Account      *proto.Account
	Workspace    *proto.Workspace
	Cleanup      func()
}

func initializeWebhookTestContext(t *testing.T) *webhookTestContext {
	// Create organization and tenant first
	org := testutils.GenerateRandomizedOrganization()
	tenant := testutils.GenerateRandomizedTenant()

	createOrgResp, err := MockServer.CreateOrganization(context.Background(), &proto.CreateOrganizationRequest{
		Organization: org,
	})
	require.NoError(t, err)
	require.NotNil(t, createOrgResp)
	require.NotNil(t, createOrgResp.Organization)

	createTenantResp, err := MockServer.CreateTenant(context.Background(), &proto.CreateTenantRequest{
		Tenant:         tenant,
		OrganizationId: createOrgResp.Organization.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createTenantResp)
	require.NotNil(t, createTenantResp.TenantId)

	// Create account
	account := testutils.GenerateRandomizedAccount()
	createAcctResp, err := MockServer.CreateAccount(context.Background(), &proto.CreateAccountRequest{
		Account:        account,
		OrganizationId: createOrgResp.Organization.Id,
		TenantId:       createTenantResp.TenantId,
	})
	require.NoError(t, err)
	require.NotNil(t, createAcctResp)
	require.NotNil(t, createAcctResp.Account)

	// Create workspace
	workspace := testutils.GenerateRandomWorkspace()
	createWorkspaceResp, err := MockServer.CreateWorkspace(context.Background(), &proto.CreateWorkspaceRequest{
		Workspace:      workspace,
		AccountId:      createAcctResp.Account.Id,
		TenantId:       createTenantResp.TenantId,
		OrganizationId: createOrgResp.Organization.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createWorkspaceResp)
	require.NotNil(t, createWorkspaceResp.Workspace)

	// Return test context with cleanup function
	return &webhookTestContext{
		Organization: createOrgResp.Organization,
		TenantId:     createTenantResp.TenantId,
		Account:      createAcctResp.Account,
		Workspace:    createWorkspaceResp.Workspace,
		Cleanup: func() {
			ctx := context.Background()

			// Delete workspace
			_, err := MockServer.DeleteWorkspace(ctx, &proto.DeleteWorkspaceRequest{
				Id: createWorkspaceResp.Workspace.Id,
			})
			if err != nil {
				t.Logf("Failed to cleanup test workspace: %v", err)
			}

			// Delete account
			_, err = MockServer.DeleteAccount(ctx, &proto.DeleteAccountRequest{
				Id:             createAcctResp.Account.Id,
				OrganizationId: createOrgResp.Organization.Id,
				TenantId:       createTenantResp.TenantId,
			})
			if err != nil {
				t.Logf("Failed to cleanup test account: %v", err)
			}

			// Delete tenant
			_, err = MockServer.DeleteTenant(ctx, &proto.DeleteTenantRequest{
				TenantId:       createTenantResp.TenantId,
				OrganizationId: createOrgResp.Organization.Id,
			})
			if err != nil {
				t.Logf("Failed to cleanup test tenant: %v", err)
			}

			// Delete organization
			_, err = MockServer.DeleteOrganization(ctx, &proto.DeleteOrganizationRequest{
				Id: createOrgResp.Organization.Id,
			})
			if err != nil {
				t.Logf("Failed to cleanup test organization: %v", err)
			}
		},
	}
}

func TestServer_CreateWebhook(t *testing.T) {
	testCtx := initializeWebhookTestContext(t)
	defer testCtx.Cleanup()

	tests := []struct {
		name    string
		req     *proto.CreateWebhookRequest
		wantErr bool
		errCode codes.Code
		verify  func(t *testing.T, resp *proto.CreateWebhookResponse)
	}{
		{
			name: "success",
			req: &proto.CreateWebhookRequest{
				Webhook: &proto.WebhookConfig{
					WebhookName:   "Test Webhook",
					Url:           "https://example.com/webhook",
					AuthType:      "basic",
					AuthToken:     "test-token",
					CustomHeaders: map[string]string{"Content-Type": "application/json"},
					MaxRetries:    3,
					VerifySsl:     true,
					SigningSecret: "test-secret",
				},
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.CreateWebhookResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Webhook)
				assert.NotEmpty(t, resp.Webhook.Id)
				assert.Equal(t, "Test Webhook", resp.Webhook.WebhookName)
				assert.Equal(t, "https://example.com/webhook", resp.Webhook.Url)
				assert.Equal(t, "basic", resp.Webhook.AuthType)
				assert.Equal(t, "test-token", resp.Webhook.AuthToken)
				assert.Equal(t, int32(3), resp.Webhook.MaxRetries)
				assert.True(t, resp.Webhook.VerifySsl)
				assert.Equal(t, "test-secret", resp.Webhook.SigningSecret)
			},
		},
		{
			name:    "error - nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - missing webhook",
			req: &proto.CreateWebhookRequest{
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - invalid URL",
			req: &proto.CreateWebhookRequest{
				Webhook: &proto.WebhookConfig{
					WebhookName:   "Test Webhook",
					Url:           "not-a-url",
					AuthType:      "basic",
					AuthToken:     "test-token",
					CustomHeaders: map[string]string{"Content-Type": "application/json"},
					MaxRetries:    3,
					VerifySsl:     true,
					SigningSecret: "test-secret",
				},
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.CreateWebhook(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}

			require.NoError(t, err)
			if tt.verify != nil {
				tt.verify(t, resp)
			}
		})
	}
}

func TestServer_GetWebhook(t *testing.T) {
	testCtx := initializeWebhookTestContext(t)
	defer testCtx.Cleanup()

	// Create a test webhook first
	createResp, err := MockServer.CreateWebhook(context.Background(), &proto.CreateWebhookRequest{
		Webhook: &proto.WebhookConfig{
			WebhookName:   "Test Webhook",
			Url:           "https://example.com/webhook",
			AuthType:      "basic",
			AuthToken:     "test-token",
			CustomHeaders: map[string]string{"Content-Type": "application/json"},
			MaxRetries:    3,
			VerifySsl:     true,
			SigningSecret: "test-secret",
		},
		WorkspaceId: testCtx.Workspace.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Webhook)

	tests := []struct {
		name    string
		req     *proto.GetWebhookRequest
		wantErr bool
		errCode codes.Code
		verify  func(t *testing.T, resp *proto.GetWebhookResponse)
	}{
		{
			name: "success",
			req: &proto.GetWebhookRequest{
				WebhookId:   createResp.Webhook.Id,
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.GetWebhookResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Webhook)
				assert.Equal(t, createResp.Webhook.Id, resp.Webhook.Id)
				assert.Equal(t, createResp.Webhook.WebhookName, resp.Webhook.WebhookName)
				assert.Equal(t, createResp.Webhook.Url, resp.Webhook.Url)
				assert.Equal(t, createResp.Webhook.AuthType, resp.Webhook.AuthType)
				assert.Equal(t, createResp.Webhook.AuthToken, resp.Webhook.AuthToken)
				assert.Equal(t, createResp.Webhook.MaxRetries, resp.Webhook.MaxRetries)
				assert.Equal(t, createResp.Webhook.VerifySsl, resp.Webhook.VerifySsl)
				assert.Equal(t, createResp.Webhook.SigningSecret, resp.Webhook.SigningSecret)
			},
		},
		{
			name:    "error - nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - webhook not found",
			req: &proto.GetWebhookRequest{
				WebhookId:   999999,
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "error - invalid webhook ID",
			req: &proto.GetWebhookRequest{
				WebhookId:   0,
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.GetWebhook(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}

			require.NoError(t, err)
			if tt.verify != nil {
				tt.verify(t, resp)
			}
		})
	}
}

func TestServer_UpdateWebhook(t *testing.T) {
	testCtx := initializeWebhookTestContext(t)
	defer testCtx.Cleanup()

	// Create a test webhook first
	createResp, err := MockServer.CreateWebhook(context.Background(), &proto.CreateWebhookRequest{
		Webhook: &proto.WebhookConfig{
			WebhookName:   "Test Webhook",
			Url:           "https://example.com/webhook",
			AuthType:      "basic",
			AuthToken:     "test-token",
			CustomHeaders: map[string]string{"Content-Type": "application/json"},
			MaxRetries:    3,
			VerifySsl:     true,
			SigningSecret: "test-secret",
		},
		WorkspaceId: testCtx.Workspace.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Webhook)

	tests := []struct {
		name    string
		req     *proto.UpdateWebhookRequest
		wantErr bool
		errCode codes.Code
		verify  func(t *testing.T, resp *proto.UpdateWebhookResponse)
	}{
		{
			name: "success",
			req: &proto.UpdateWebhookRequest{
				Webhook: &proto.WebhookConfig{
					Id:            createResp.Webhook.Id,
					WebhookName:   "Updated Webhook",
					Url:           "https://example.com/webhook/v2",
					AuthType:      "bearer",
					AuthToken:     "updated-token",
					CustomHeaders: map[string]string{"Authorization": "Bearer token"},
					MaxRetries:    5,
					VerifySsl:     false,
					SigningSecret: "updated-secret",
				},
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.UpdateWebhookResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Webhook)
				assert.Equal(t, createResp.Webhook.Id, resp.Webhook.Id)
				assert.Equal(t, "Updated Webhook", resp.Webhook.WebhookName)
				assert.Equal(t, "https://example.com/webhook/v2", resp.Webhook.Url)
				assert.Equal(t, "bearer", resp.Webhook.AuthType)
				assert.Equal(t, "updated-token", resp.Webhook.AuthToken)
				assert.Equal(t, int32(5), resp.Webhook.MaxRetries)
				assert.False(t, resp.Webhook.VerifySsl)
				assert.Equal(t, "updated-secret", resp.Webhook.SigningSecret)
			},
		},
		{
			name:    "error - nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - missing webhook",
			req: &proto.UpdateWebhookRequest{
				Webhook: testutils.GenerateRandomWebhookConfig(),
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - webhook not found",
			req: &proto.UpdateWebhookRequest{
				Webhook: &proto.WebhookConfig{
					Id:          999999,
					WebhookName: "Non-existent",
				},
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.UpdateWebhook(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}

			require.NoError(t, err)
			if tt.verify != nil {
				tt.verify(t, resp)
			}
		})
	}
}

func TestServer_DeleteWebhook(t *testing.T) {
	testCtx := initializeWebhookTestContext(t)
	defer testCtx.Cleanup()

	// Create a test webhook first
	createResp, err := MockServer.CreateWebhook(context.Background(), &proto.CreateWebhookRequest{
		Webhook: &proto.WebhookConfig{
			WebhookName:   "Test Webhook",
			Url:           "https://example.com/webhook",
			AuthType:      "basic",
			AuthToken:     "test-token",
			CustomHeaders: map[string]string{"Content-Type": "application/json"},
			MaxRetries:    3,
			VerifySsl:     true,
			SigningSecret: "test-secret",
		},
		WorkspaceId: testCtx.Workspace.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Webhook)

	tests := []struct {
		name    string
		req     *proto.DeleteWebhookRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.DeleteWebhookRequest{
				WebhookId:   createResp.Webhook.Id,
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: false,
		},
		{
			name:    "error - nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - webhook not found",
			req: &proto.DeleteWebhookRequest{
				WebhookId:   999999,
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "error - invalid webhook ID",
			req: &proto.DeleteWebhookRequest{
				WebhookId:   0,
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.DeleteWebhook(context.Background(), tt.req)
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

func TestServer_ListWebhooks(t *testing.T) {
	testCtx := initializeWebhookTestContext(t)
	defer testCtx.Cleanup()

	// Create some test webhooks first
	numWebhooks := 5
	webhooks := make([]*proto.WebhookConfig, 0, numWebhooks)
	for i := 0; i < numWebhooks; i++ {
		createResp, err := MockServer.CreateWebhook(context.Background(), &proto.CreateWebhookRequest{
			Webhook: &proto.WebhookConfig{
				WebhookName:   "Test Webhook",
				Url:           "https://example.com/webhook",
				AuthType:      "basic",
				AuthToken:     "test-token",
				CustomHeaders: map[string]string{"Content-Type": "application/json"},
				MaxRetries:    3,
				VerifySsl:     true,
				SigningSecret: "test-secret",
			},
			WorkspaceId: testCtx.Workspace.Id,
		})
		require.NoError(t, err)
		require.NotNil(t, createResp)
		require.NotNil(t, createResp.Webhook)
		webhooks = append(webhooks, createResp.Webhook)
	}

	tests := []struct {
		name    string
		req     *proto.ListWebhooksRequest
		wantErr bool
		errCode codes.Code
		verify  func(t *testing.T, resp *proto.ListWebhooksResponse)
	}{
		{
			name: "success - list all webhooks",
			req: &proto.ListWebhooksRequest{
				WorkspaceId: testCtx.Workspace.Id,
				PageSize:    10,
				PageNumber:  1,
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.ListWebhooksResponse) {
				require.NotNil(t, resp)
				assert.Len(t, resp.Webhooks, numWebhooks)
				assert.Equal(t, int32(0), resp.NextPageNumber) // No more pages
			},
		},
		{
			name: "success - pagination",
			req: &proto.ListWebhooksRequest{
				WorkspaceId: testCtx.Workspace.Id,
				PageSize:    2,
				PageNumber:  1,
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.ListWebhooksResponse) {
				require.NotNil(t, resp)
				assert.Len(t, resp.Webhooks, 2)
				assert.Equal(t, int32(2), resp.NextPageNumber)
			},
		},
		{
			name:    "error - nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - invalid page size",
			req: &proto.ListWebhooksRequest{
				WorkspaceId: testCtx.Workspace.Id,
				PageSize:    0,
				PageNumber:  1,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - invalid page number",
			req: &proto.ListWebhooksRequest{
				WorkspaceId: testCtx.Workspace.Id,
				PageSize:    10,
				PageNumber:  0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.ListWebhooks(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}

			require.NoError(t, err)
			if tt.verify != nil {
				tt.verify(t, resp)
			}
		})
	}
} 