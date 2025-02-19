package grpc

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/alecthomas/assert/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_UpdateWorkspace(t *testing.T) {
	testCtx := initializeWorkspaceTestContext(t)
	defer testCtx.Cleanup()

	req := &proto.CreateWorkspaceRequest{
		Workspace:      testutils.GenerateRandomWorkspace(),
		AccountId:      testCtx.Account.Id,
		OrganizationId: testCtx.Organization.Id,
		TenantId:       testCtx.TenantId,
	}

	// Create a test workspace first
	createResp, err := MockServer.CreateWorkspace(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Workspace)

	workspace := createResp.Workspace
	workspace.Name = "Updated Workspace Name"
	workspace.Industry = "Technology"
	workspace.Domain = "updated-domain.com"
	workspace.GdprCompliant = true
	workspace.HipaaCompliant = true
	workspace.Soc2Compliant = true
	workspace.StorageQuota = 1000000
	workspace.UsedStorage = 500000
	workspace.WorkspaceJobLimit = 100

	tests := []struct {
		name     string
		req      *proto.UpdateWorkspaceRequest
		wantErr  bool
		errCode  codes.Code
		validate func(t *testing.T, resp *proto.UpdateWorkspaceResponse)
	}{
		{
			name: "success - update all fields",
			req: &proto.UpdateWorkspaceRequest{
				Workspace: workspace,
			},
			wantErr: false,
			validate: func(t *testing.T, resp *proto.UpdateWorkspaceResponse) {
				assert.Equal(t, workspace.Id, resp.Workspace.Id)
				assert.Equal(t, "Updated Workspace Name", resp.Workspace.Name)
				assert.Equal(t, "Technology", resp.Workspace.Industry)
				assert.Equal(t, "updated-domain.com", resp.Workspace.Domain)
				assert.True(t, resp.Workspace.GdprCompliant)
				assert.True(t, resp.Workspace.HipaaCompliant)
				assert.True(t, resp.Workspace.Soc2Compliant)
				assert.Equal(t, int64(1000000), resp.Workspace.StorageQuota)
				assert.Equal(t, int64(500000), resp.Workspace.UsedStorage)
				assert.Equal(t, int32(100), resp.Workspace.WorkspaceJobLimit)
			},
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "nil workspace",
			req: &proto.UpdateWorkspaceRequest{
				Workspace: nil,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid workspace ID",
			req: &proto.UpdateWorkspaceRequest{
				Workspace: &proto.Workspace{
					Id:   0,
					Name: "Invalid ID Workspace",
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "workspace not found",
			req: &proto.UpdateWorkspaceRequest{
				Workspace: &proto.Workspace{
					Id:                999999,
					Name:              "Non-existent Workspace",
					WorkspaceJobLimit: 100,
					StorageQuota:      1000000,
					UsedStorage:       0,
				},
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "invalid storage quota",
			req: &proto.UpdateWorkspaceRequest{
				Workspace: &proto.Workspace{
					Id:           workspace.Id,
					Name:         "Invalid Storage Quota",
					StorageQuota: -1,
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid used storage",
			req: &proto.UpdateWorkspaceRequest{
				Workspace: &proto.Workspace{
					Id:          workspace.Id,
					Name:        "Invalid Used Storage",
					UsedStorage: -1,
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid job limit",
			req: &proto.UpdateWorkspaceRequest{
				Workspace: &proto.Workspace{
					Id:                workspace.Id,
					Name:              "Invalid Job Limit",
					WorkspaceJobLimit: -1,
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.UpdateWorkspace(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Workspace)
			if tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}

	// Test concurrent updates
	t.Run("concurrent updates", func(t *testing.T) {
		numUpdates := 5
		var wg sync.WaitGroup
		errors := make(chan error, numUpdates)
		results := make(chan *proto.UpdateWorkspaceResponse, numUpdates)

		for i := 0; i < numUpdates; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				updateWorkspace := &proto.Workspace{
					Id:                workspace.Id,
					Name:              fmt.Sprintf("Concurrent Update %d", index),
					Industry:          fmt.Sprintf("Industry %d", index),
					Domain:            fmt.Sprintf("domain-%d.com", index),
					GdprCompliant:     true,
					HipaaCompliant:    true,
					Soc2Compliant:     true,
					StorageQuota:      int64(1000000 + index),
					UsedStorage:       int64(500000 + index),
					WorkspaceJobLimit: int32(100 + index),
				}
				resp, err := MockServer.UpdateWorkspace(context.Background(), &proto.UpdateWorkspaceRequest{
					Workspace: updateWorkspace,
				})
				if err != nil {
					errors <- err
					return
				}
				results <- resp
			}(i)
		}

		wg.Wait()
		close(errors)
		close(results)

		// Check for errors
		var errs []error
		for err := range errors {
			errs = append(errs, err)
		}
		require.Empty(t, errs, "Expected no errors during concurrent updates")

		// Verify all updates were successful
		var updates []*proto.UpdateWorkspaceResponse
		for result := range results {
			updates = append(updates, result)
		}
		require.Len(t, updates, numUpdates, "Expected %d successful updates", numUpdates)

		// Verify final state
		getResp, err := MockServer.GetWorkspace(context.Background(), &proto.GetWorkspaceRequest{
			Id:             workspace.Id,
			AccountId:      testCtx.Account.Id,
			OrganizationId: testCtx.Organization.Id,
			TenantId:       testCtx.TenantId,
		})
		require.NoError(t, err)
		require.NotNil(t, getResp)
		require.NotNil(t, getResp.Workspace)
		assert.Contains(t, getResp.Workspace.Name, "Concurrent Update")
		assert.Contains(t, getResp.Workspace.Industry, "Industry")
		assert.Contains(t, getResp.Workspace.Domain, "domain-")
		assert.True(t, getResp.Workspace.GdprCompliant)
		assert.True(t, getResp.Workspace.HipaaCompliant)
		assert.True(t, getResp.Workspace.Soc2Compliant)
	})
}
