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

type leadTestContext struct {
	Organization *proto.Organization
	TenantId     uint64
	Account      *proto.Account
	Workspace    *proto.Workspace
	Cleanup      func()
}

func initializeLeadTestContext(t *testing.T) *leadTestContext {
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
	return &leadTestContext{
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
				Id: createAcctResp.Account.Id,
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

func TestServer_ListLeads(t *testing.T) {
	testCtx := initializeLeadTestContext(t)
	defer testCtx.Cleanup()

	// Create some test leads first
	numLeads := 5
	leads := make([]*proto.Lead, 0, numLeads)
	for i := 0; i < numLeads; i++ {
		lead := testutils.GenerateRandomLead()
		leads = append(leads, lead)
	}

	tests := []struct {
		name    string
		req     *proto.ListLeadsRequest
		wantErr bool
		errCode codes.Code
		verify  func(t *testing.T, resp *proto.ListLeadsResponse)
	}{
		{
			name: "success - list all leads",
			req: &proto.ListLeadsRequest{
				PageSize:   10,
				PageNumber: 1,
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.ListLeadsResponse) {
				require.NotNil(t, resp)
				assert.NotEmpty(t, resp.Leads)
				assert.Equal(t, int32(0), resp.NextPageNumber) // No more pages
			},
		},
		{
			name: "success - pagination",
			req: &proto.ListLeadsRequest{
				PageSize:   2,
				PageNumber: 1,
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.ListLeadsResponse) {
				require.NotNil(t, resp)
				assert.Len(t, resp.Leads, 2)
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
			req: &proto.ListLeadsRequest{
				PageSize:   0,
				PageNumber: 1,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - invalid page number",
			req: &proto.ListLeadsRequest{
				PageSize:   10,
				PageNumber: 0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.ListLeads(context.Background(), tt.req)
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

func TestServer_GetLead(t *testing.T) {
	testCtx := initializeLeadTestContext(t)
	defer testCtx.Cleanup()

	// Create a test lead first
	lead := testutils.GenerateRandomLead()

	tests := []struct {
		name    string
		req     *proto.GetLeadRequest
		wantErr bool
		errCode codes.Code
		verify  func(t *testing.T, resp *proto.GetLeadResponse)
	}{
		{
			name: "success",
			req: &proto.GetLeadRequest{
				LeadId: lead.Id,
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.GetLeadResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Lead)
				assert.Equal(t, lead.Id, resp.Lead.Id)
			},
		},
		{
			name:    "error - nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "error - lead not found",
			req: &proto.GetLeadRequest{
				LeadId: 999999,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "error - invalid lead ID",
			req: &proto.GetLeadRequest{
				LeadId: 0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.GetLead(context.Background(), tt.req)
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