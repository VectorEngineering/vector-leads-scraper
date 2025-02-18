package grpc

import (
	"context"
	"testing"

	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_GetWebhook(t *testing.T) {
	testCtx := initializeWebhookTestContext(t)
	defer testCtx.Cleanup()

	// Create a test webhook first
	createResp, err := MockServer.CreateWebhook(context.Background(), &proto.CreateWebhookRequest{
		OrganizationId: testCtx.Organization.Id,
		TenantId:       testCtx.TenantId,
		AccountId:      testCtx.Account.Id,
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
				WebhookId:      createResp.Webhook.Id,
				WorkspaceId:    testCtx.Workspace.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
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
				WebhookId:      999999,
				WorkspaceId:    testCtx.Workspace.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
			},
			wantErr: true,
			errCode: codes.Internal,
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
