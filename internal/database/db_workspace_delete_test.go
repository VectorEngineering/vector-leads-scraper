package database

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDb_DeleteWorkspace(t *testing.T) {
	// Initialize test context
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	tests := []struct {
		name    string
		id      uint64
		setup   func(t *testing.T) uint64
		wantErr bool
		errType error
	}{
		{
			name: "[success] delete existing workspace",
			setup: func(t *testing.T) uint64 {
				// Create a new workspace to delete
				workspace, err := conn.CreateWorkspace(context.Background(), &CreateWorkspaceInput{
					Workspace:      testutils.GenerateRandomWorkspace(),
					AccountID:      tc.Account.Id,
					OrganizationID: tc.Organization.Id,
					TenantID:       tc.Tenant.Id,
				})
				require.NoError(t, err)
				require.NotNil(t, workspace)
				return workspace.Id
			},
			wantErr: false,
		},
		{
			name:    "[failure] delete with invalid ID",
			id:      0,
			wantErr: true,
			errType: ErrInvalidInput,
		},
		{
			name:    "[failure] delete non-existent workspace",
			id:      999999,
			wantErr: true,
			errType: ErrWorkspaceDoesNotExist,
		},
		{
			name: "[failure] delete already deleted workspace",
			setup: func(t *testing.T) uint64 {
				// Create a workspace first
				workspace, err := conn.CreateWorkspace(context.Background(), &CreateWorkspaceInput{
					Workspace:      testutils.GenerateRandomWorkspace(),
					AccountID:      tc.Account.Id,
					OrganizationID: tc.Organization.Id,
					TenantID:       tc.Tenant.Id,
				})
				require.NoError(t, err)
				require.NotNil(t, workspace)

				// Delete it once
				err = conn.DeleteWorkspace(context.Background(), workspace.Id)
				require.NoError(t, err)

				return workspace.Id
			},
			wantErr: true,
			errType: ErrWorkspaceDoesNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id uint64
			if tt.setup != nil {
				id = tt.setup(t)
			} else {
				id = tt.id
			}

			err := conn.DeleteWorkspace(context.Background(), id)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)

			// Verify the workspace was actually deleted
			_, err = conn.GetWorkspace(context.Background(), id)
			assert.ErrorIs(t, err, ErrWorkspaceDoesNotExist)
		})
	}
}

func TestDb_DeleteWorkspace_ConcurrentDeletions(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create multiple workspaces
	numWorkspaces := 5
	workspaces := make([]*lead_scraper_servicev1.Workspace, 0, numWorkspaces)

	for i := 0; i < numWorkspaces; i++ {
		workspace, err := conn.CreateWorkspace(context.Background(), &CreateWorkspaceInput{
			Workspace:      testutils.GenerateRandomWorkspace(),
			AccountID:      tc.Account.Id,
			OrganizationID: tc.Organization.Id,
			TenantID:       tc.Tenant.Id,
		})
		require.NoError(t, err)
		workspaces = append(workspaces, workspace)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(workspaces))

	// Delete workspaces concurrently
	for _, workspace := range workspaces {
		wg.Add(1)
		go func(id uint64) {
			defer wg.Done()
			err := conn.DeleteWorkspace(context.Background(), id)
			if err != nil && !errors.Is(err, ErrWorkspaceDoesNotExist) {
				errChan <- err
			}
		}(workspace.Id)
	}

	wg.Wait()
	close(errChan)

	// Collect any unexpected errors
	var unexpectedErrors []error
	for err := range errChan {
		unexpectedErrors = append(unexpectedErrors, err)
	}
	require.Empty(t, unexpectedErrors, "Got unexpected errors during concurrent deletions: %v", unexpectedErrors)

	// Verify all workspaces are deleted
	for _, workspace := range workspaces {
		_, err := conn.GetWorkspace(context.Background(), workspace.Id)
		require.Error(t, err)
	}
}
