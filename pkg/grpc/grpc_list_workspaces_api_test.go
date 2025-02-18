package grpc

import (
	"context"
	"reflect"
	"testing"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type listWorkspacesTestContext struct {
	Organization *proto.Organization
	TenantId     uint64
	Account      *proto.Account
	Workspace    *proto.Workspace
	Cleanup      func()
}

func initializeListWorkspacesTestContext(t *testing.T) *listWorkspacesTestContext {
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

	return &listWorkspacesTestContext{
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

func createTestWorkspaces(t *testing.T, testCtx *listWorkspacesTestContext, count int) []*proto.Workspace {
	workspaces := make([]*proto.Workspace, 0, count)
	for i := 0; i < count; i++ {
		workspace := testutils.GenerateRandomWorkspace()
		workspace.WorkspaceJobLimit = 100 // Set a valid job limit
		req := &proto.CreateWorkspaceRequest{
			Workspace:      workspace,
			AccountId:      testCtx.Account.Id,
			OrganizationId: testCtx.Organization.Id,
			TenantId:       testCtx.TenantId,
		}
		createResp, err := MockServer.CreateWorkspace(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, createResp)
		require.NotNil(t, createResp.Workspace)
		workspaces = append(workspaces, createResp.Workspace)
	}
	return workspaces
}

func TestListWorkspaces_Success(t *testing.T) {
	testCtx := initializeListWorkspacesTestContext(t)
	defer testCtx.Cleanup()

	// Create test workspaces
	workspaces := createTestWorkspaces(t, testCtx, 5)
	require.Len(t, workspaces, 5)

	tests := []struct {
		name string
		req  *proto.ListWorkspacesRequest
	}{
		{
			name: "successful list with default pagination",
			req: &proto.ListWorkspacesRequest{
				PageSize:       50,
				PageNumber:     1,
				AccountId:      testCtx.Account.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
			},
		},
		{
			name: "successful list with custom pagination",
			req: &proto.ListWorkspacesRequest{
				PageSize:       2,
				PageNumber:     1,
				AccountId:      testCtx.Account.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.ListWorkspaces(context.Background(), tt.req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.NotNil(t, resp.Workspaces)
			assert.Greater(t, resp.NextPageNumber, int32(0))
		})
	}
}

func TestListWorkspaces_InvalidRequest(t *testing.T) {
	testCtx := initializeListWorkspacesTestContext(t)
	defer testCtx.Cleanup()

	t.Run("nil request", func(t *testing.T) {
		resp, err := MockServer.ListWorkspaces(context.Background(), nil)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp)
	})

	t.Run("missing account ID", func(t *testing.T) {
		req := &proto.ListWorkspacesRequest{
			OrganizationId: testCtx.Organization.Id,
			TenantId:       testCtx.TenantId,
			PageSize:       50,
			PageNumber:     1,
		}

		resp, err := MockServer.ListWorkspaces(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp)
	})

	t.Run("missing organization ID", func(t *testing.T) {
		req := &proto.ListWorkspacesRequest{
			AccountId:  testCtx.Account.Id,
			TenantId:   testCtx.TenantId,
			PageSize:   50,
			PageNumber: 1,
		}

		resp, err := MockServer.ListWorkspaces(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp)
	})

	t.Run("missing tenant ID", func(t *testing.T) {
		req := &proto.ListWorkspacesRequest{
			AccountId:      testCtx.Account.Id,
			OrganizationId: testCtx.Organization.Id,
			PageSize:       50,
			PageNumber:     1,
		}

		resp, err := MockServer.ListWorkspaces(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp)
	})

	t.Run("invalid page size", func(t *testing.T) {
		req := &proto.ListWorkspacesRequest{
			AccountId:      testCtx.Account.Id,
			OrganizationId: testCtx.Organization.Id,
			TenantId:       testCtx.TenantId,
			PageSize:       -1,
			PageNumber:     1,
		}

		resp, err := MockServer.ListWorkspaces(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp)
	})

	t.Run("invalid page number", func(t *testing.T) {
		req := &proto.ListWorkspacesRequest{
			AccountId:      testCtx.Account.Id,
			OrganizationId: testCtx.Organization.Id,
			TenantId:       testCtx.TenantId,
			PageSize:       50,
			PageNumber:     -1,
		}

		resp, err := MockServer.ListWorkspaces(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp)
	})
}

func TestServer_ListWorkspaces(t *testing.T) {
	type args struct {
		ctx context.Context
		req *proto.ListWorkspacesRequest
	}
	tests := []struct {
		name    string
		s       *Server
		args    args
		want    *proto.ListWorkspacesResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.ListWorkspaces(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.ListWorkspaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.ListWorkspaces() = %v, want %v", got, tt.want)
			}
		})
	}
}
