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
