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

func TestServer_ListWebhooks(t *testing.T) {
	testCtx := initializeWebhookTestContext(t)
	defer testCtx.Cleanup()

	// Create some test webhooks first
	numWebhooks := 5
	webhooks := make([]*proto.WebhookConfig, 0, numWebhooks)
	for i := 0; i < numWebhooks; i++ {
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
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
				WorkspaceId:    testCtx.Workspace.Id,
				PageSize:       10,
				PageNumber:     1,
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.ListWebhooksResponse) {
				require.NotNil(t, resp)
				assert.Equal(t, int32(0), resp.NextPageNumber) // No more pages
			},
		},
		{
			name: "success - pagination",
			req: &proto.ListWebhooksRequest{
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
				WorkspaceId:    testCtx.Workspace.Id,
				PageSize:       2,
				PageNumber:     1,
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.ListWebhooksResponse) {
				require.NotNil(t, resp)
				assert.Len(t, resp.Webhooks, 2)
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
