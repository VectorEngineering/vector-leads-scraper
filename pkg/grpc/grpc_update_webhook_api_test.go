package grpc

import (
	"context"
	"testing"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_UpdateWebhook(t *testing.T) {
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
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.UpdateWebhookResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Webhook)
				require.Equal(t, createResp.Webhook.Id, resp.Webhook.Id)
				require.Equal(t, "Updated Webhook", resp.Webhook.WebhookName)
				require.Equal(t, "https://example.com/webhook/v2", resp.Webhook.Url)
				require.Equal(t, "bearer", resp.Webhook.AuthType)
				require.Equal(t, "updated-token", resp.Webhook.AuthToken)
				require.Equal(t, int32(5), resp.Webhook.MaxRetries)
				require.False(t, resp.Webhook.VerifySsl)
				require.Equal(t, "updated-secret", resp.Webhook.SigningSecret)
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
			errCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.UpdateWebhook(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.errCode, st.Code())
				return
			}

			require.NoError(t, err)
			if tt.verify != nil {
				tt.verify(t, resp)
			}
		})
	}
}
