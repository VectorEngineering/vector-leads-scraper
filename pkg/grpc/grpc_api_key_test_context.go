package grpc

import (
	"context"
	"testing"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/require"
)

type apiKeyTestContext struct {
	Organization *proto.Organization
	TenantId     uint64
	Account      *proto.Account
	Workspace    *proto.Workspace
	Cleanup      func()
}

func initializeAPIKeyTestContext(t *testing.T) *apiKeyTestContext {
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
		Account:              account,
		OrganizationId:       createOrgResp.Organization.Id,
		TenantId:             createTenantResp.TenantId,
		InitialWorkspaceName: testutils.GenerateRandomString(10, true, true),
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
	return &apiKeyTestContext{
		Organization: createOrgResp.Organization,
		TenantId:     createTenantResp.TenantId,
		Account:      createAcctResp.Account,
		Workspace:    createWorkspaceResp.Workspace,
		Cleanup: func() {
			ctx := context.Background()

			// Delete in reverse order of dependencies
			// First delete API keys as they depend on workspaces
			apiKeysResp, err := MockServer.ListAPIKeys(ctx, &proto.ListAPIKeysRequest{
				WorkspaceId:    createWorkspaceResp.Workspace.Id,
				OrganizationId: createOrgResp.Organization.Id,
				TenantId:       createTenantResp.TenantId,
				AccountId:      createAcctResp.Account.Id,
				PageSize:       100,
				PageNumber:     1,
			})
			if err == nil && apiKeysResp != nil && len(apiKeysResp.ApiKeys) > 0 {
				for _, apiKey := range apiKeysResp.ApiKeys {
					_, err = MockServer.DeleteAPIKey(ctx, &proto.DeleteAPIKeyRequest{
						KeyId:          apiKey.Id,
						WorkspaceId:    createWorkspaceResp.Workspace.Id,
						OrganizationId: createOrgResp.Organization.Id,
						TenantId:       createTenantResp.TenantId,
						AccountId:      createAcctResp.Account.Id,
					})
					if err != nil {
						t.Logf("Failed to cleanup test API key %d: %v", apiKey.Id, err)
					}
				}
			}

			// Then delete webhooks as they depend on workspaces
			webhooksResp, err := MockServer.ListWebhooks(ctx, &proto.ListWebhooksRequest{
				WorkspaceId:    createWorkspaceResp.Workspace.Id,
				OrganizationId: createOrgResp.Organization.Id,
				TenantId:       createTenantResp.TenantId,
				AccountId:      createAcctResp.Account.Id,
				PageSize:       100,
				PageNumber:     1,
			})
			if err == nil && webhooksResp != nil && len(webhooksResp.Webhooks) > 0 {
				for _, webhook := range webhooksResp.Webhooks {
					_, err = MockServer.DeleteWebhook(ctx, &proto.DeleteWebhookRequest{
						WebhookId:      webhook.Id,
						WorkspaceId:    createWorkspaceResp.Workspace.Id,
						OrganizationId: createOrgResp.Organization.Id,
						TenantId:       createTenantResp.TenantId,
						AccountId:      createAcctResp.Account.Id,
					})
					if err != nil {
						t.Logf("Failed to cleanup test webhook %d: %v", webhook.Id, err)
					}
				}
			}

			// Then delete workspaces as they depend on accounts
			_, err = MockServer.DeleteWorkspace(ctx, &proto.DeleteWorkspaceRequest{
				Id: createWorkspaceResp.Workspace.Id,
			})
			if err != nil {
				t.Logf("Failed to cleanup test workspace: %v", err)
			}

			// Then delete accounts as they depend on tenants
			_, err = MockServer.DeleteAccount(ctx, &proto.DeleteAccountRequest{
				Id:             createAcctResp.Account.Id,
				OrganizationId: createOrgResp.Organization.Id,
				TenantId:       createTenantResp.TenantId,
			})
			if err != nil {
				t.Logf("Failed to cleanup test account: %v", err)
			}

			// Then delete tenants as they depend on organizations
			_, err = MockServer.DeleteTenant(ctx, &proto.DeleteTenantRequest{
				TenantId:       createTenantResp.TenantId,
				OrganizationId: createOrgResp.Organization.Id,
			})
			if err != nil {
				t.Logf("Failed to cleanup test tenant: %v", err)
			}

			// Finally delete organization as it's the root resource
			_, err = MockServer.DeleteOrganization(ctx, &proto.DeleteOrganizationRequest{
				Id: createOrgResp.Organization.Id,
			})
			if err != nil {
				t.Logf("Failed to cleanup test organization: %v", err)
			}
		},
	}
}
