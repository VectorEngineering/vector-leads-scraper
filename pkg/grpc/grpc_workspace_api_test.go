package grpc

import (
	"context"
	"testing"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/require"
)

type workspaceTestContext struct {
	Organization *proto.Organization
	TenantId     uint64
	Account      *proto.Account
	Cleanup      func()
}

func initializeWorkspaceTestContext(t *testing.T) *workspaceTestContext {
	// create an organization and tenant first
	org := testutils.GenerateRandomizedOrganization()
	tenant := testutils.GenerateRandomizedTenant()

	// Ensure required fields are set
	if org.Name == "" {
		org.Name = "Test Organization"
	}
	if org.BillingEmail == "" {
		org.BillingEmail = "billing@example.com"
	}
	if org.TechnicalEmail == "" {
		org.TechnicalEmail = "tech@example.com"
	}

	createOrgResp, err := MockServer.CreateOrganization(context.Background(), &proto.CreateOrganizationRequest{
		Organization: org,
	})
	require.NoError(t, err)
	require.NotNil(t, createOrgResp)
	require.NotNil(t, createOrgResp.Organization)

	// Ensure required fields are set for tenant
	if tenant.Name == "" {
		tenant.Name = "Test Tenant"
	}
	if tenant.ApiBaseUrl == "" {
		tenant.ApiBaseUrl = "https://api.example.com"
	}

	createTenantResp, err := MockServer.CreateTenant(context.Background(), &proto.CreateTenantRequest{
		Tenant:         tenant,
		OrganizationId: createOrgResp.Organization.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createTenantResp)
	require.NotNil(t, createTenantResp.TenantId)

	// Create a test account
	createAcctResp, err := MockServer.CreateAccount(context.Background(), &proto.CreateAccountRequest{
		Account:              testutils.GenerateRandomizedAccount(),
		OrganizationId:       createOrgResp.Organization.Id,
		TenantId:             createTenantResp.TenantId,
		InitialWorkspaceName: "Default Workspace",
	})
	require.NoError(t, err)
	require.NotNil(t, createAcctResp)
	require.NotNil(t, createAcctResp.Account)

	cleanup := func() {
		ctx := context.Background()

		// First get all accounts for this tenant
		listAcctsResp, err := MockServer.ListAccounts(ctx, &proto.ListAccountsRequest{
			OrganizationId: createOrgResp.Organization.Id,
			TenantId:       createTenantResp.TenantId,
			PageSize:       100,
			PageNumber:     1,
		})
		if err == nil && listAcctsResp != nil {
			// For each account, delete its workspaces first
			for _, acct := range listAcctsResp.Accounts {
				// Get account details to access workspaces
				getAcctResp, err := MockServer.GetAccount(ctx, &proto.GetAccountRequest{
					Id:             acct.Id,
					OrganizationId: createOrgResp.Organization.Id,
					TenantId:       createTenantResp.TenantId,
				})
				if err == nil && getAcctResp != nil && getAcctResp.Account != nil {
					// Delete all workspaces for this account
					for _, workspace := range getAcctResp.Account.Workspaces {
						_, err := MockServer.DeleteWorkspace(ctx, &proto.DeleteWorkspaceRequest{
							Id: workspace.Id,
						})
						if err != nil {
							t.Logf("Failed to delete workspace %d: %v", workspace.Id, err)
						}
					}
				}

				// Now delete the account
				_, err = MockServer.DeleteAccount(ctx, &proto.DeleteAccountRequest{
					Id:             acct.Id,
					OrganizationId: createOrgResp.Organization.Id,
					TenantId:       createTenantResp.TenantId,
				})
				if err != nil {
					t.Logf("Failed to delete account %d: %v", acct.Id, err)
				}
			}
		}

		// Delete tenant
		_, err = MockServer.DeleteTenant(ctx, &proto.DeleteTenantRequest{
			TenantId:       createTenantResp.TenantId,
			OrganizationId: createOrgResp.Organization.Id,
		})
		if err != nil {
			t.Logf("Failed to delete tenant %d: %v", createTenantResp.TenantId, err)
		}

		// Delete organization
		_, err = MockServer.DeleteOrganization(ctx, &proto.DeleteOrganizationRequest{
			Id: createOrgResp.Organization.Id,
		})
		if err != nil {
			t.Logf("Failed to delete organization %d: %v", createOrgResp.Organization.Id, err)
		}
	}

	return &workspaceTestContext{
		Organization: createOrgResp.Organization,
		TenantId:     createTenantResp.TenantId,
		Account:      createAcctResp.Account,
		Cleanup:      cleanup,
	}
}
