package database

import (
	"context"
	"testing"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDb_GetWorkspace(t *testing.T) {
	ctx := context.Background()
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create a test workspace
	workspace, err := conn.CreateWorkspace(ctx, &CreateWorkspaceInput{
		Workspace:      testutils.GenerateRandomWorkspace(),
		AccountID:      tc.Account.Id,
		TenantID:       tc.Tenant.Id,
		OrganizationID: tc.Organization.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, workspace)

	// Clean up workspace after test
	defer func() {
		err := conn.DeleteWorkspace(ctx, workspace.Id)
		require.NoError(t, err)
	}()

	tests := []struct {
		name    string
		id      uint64
		want    *lead_scraper_servicev1.Workspace
		wantErr bool
		errType error
	}{
		{
			name:    "success - get existing workspace",
			id:      workspace.Id,
			want:    workspace,
			wantErr: false,
		},
		{
			name:    "error - workspace does not exist",
			id:      999999,
			want:    nil,
			wantErr: true,
			errType: ErrWorkspaceDoesNotExist,
		},
		{
			name:    "error - invalid workspace ID",
			id:      0,
			want:    nil,
			wantErr: true,
			errType: ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conn.GetWorkspace(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tt.want.Id, got.Id)
			assert.Equal(t, tt.want.Name, got.Name)
		})
	}
}

func TestDb_ListWorkspaces(t *testing.T) {
	ctx := context.Background()
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Clean up any existing workspaces first
	result := conn.Client.Engine.Exec("DELETE FROM workspaces")
	require.NoError(t, result.Error)

	// Create multiple test workspaces
	numWorkspaces := 5
	createdWorkspaces := make([]*lead_scraper_servicev1.Workspace, 0, numWorkspaces)
	for i := 0; i < numWorkspaces; i++ {
		workspace, err := conn.CreateWorkspace(ctx, &CreateWorkspaceInput{
			Workspace:      testutils.GenerateRandomWorkspace(),
			AccountID:      tc.Account.Id,
			TenantID:       tc.Tenant.Id,
			OrganizationID: tc.Organization.Id,
		})
		require.NoError(t, err)
		createdWorkspaces = append(createdWorkspaces, workspace)
	}

	// Clean up workspaces after test
	defer func() {
		for _, workspace := range createdWorkspaces {
			err := conn.DeleteWorkspace(ctx, workspace.Id)
			require.NoError(t, err)
		}
	}()

	tests := []struct {
		name    string
		limit   int
		offset  int
		want    []*lead_scraper_servicev1.Workspace
		wantErr bool
		errType error
	}{
		{
			name:    "success - get all workspaces",
			limit:   numWorkspaces,
			offset:  0,
			want:    createdWorkspaces,
			wantErr: false,
		},
		{
			name:    "success - get first 2 workspaces",
			limit:   2,
			offset:  0,
			want:    createdWorkspaces[:2],
			wantErr: false,
		},
		{
			name:    "success - get last 2 workspaces",
			limit:   2,
			offset:  numWorkspaces - 2,
			want:    createdWorkspaces[numWorkspaces-2:],
			wantErr: false,
		},
		{
			name:    "error - invalid limit",
			limit:   0,
			offset:  0,
			want:    nil,
			wantErr: true,
			errType: ErrInvalidInput,
		},
		{
			name:    "error - negative limit",
			limit:   -1,
			offset:  0,
			want:    nil,
			wantErr: true,
			errType: ErrInvalidInput,
		},
		{
			name:    "error - negative offset",
			limit:   1,
			offset:  -1,
			want:    nil,
			wantErr: true,
			errType: ErrInvalidInput,
		},
		{
			name:    "success - offset beyond available workspaces",
			limit:   1,
			offset:  numWorkspaces + 1,
			want:    []*lead_scraper_servicev1.Workspace{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conn.ListWorkspaces(ctx, tc.Account.Id, tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, len(tt.want))

			if len(tt.want) > 0 {
				// Verify workspaces are returned in the correct order
				for i := 0; i < len(tt.want); i++ {
					expectedWorkspace := tt.want[i]
					actualWorkspace := got[i]
					assert.Equal(t, expectedWorkspace.Id, actualWorkspace.Id)
					assert.Equal(t, expectedWorkspace.Name, actualWorkspace.Name)
				}
			}
		})
	}
}

func TestListWorkspaces(t *testing.T) {
	ctx := context.Background()
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Clean up any existing workspaces first
	result := conn.Client.Engine.Exec("DELETE FROM workspaces")
	require.NoError(t, result.Error)

	// Create multiple test workspaces
	numWorkspaces := 5
	createdWorkspaces := make([]*lead_scraper_servicev1.Workspace, 0, numWorkspaces)
	for i := 0; i < numWorkspaces; i++ {
		workspace, err := conn.CreateWorkspace(ctx, &CreateWorkspaceInput{
			Workspace:      testutils.GenerateRandomWorkspace(),
			AccountID:      tc.Account.Id,
			TenantID:       tc.Tenant.Id,
			OrganizationID: tc.Organization.Id,
		})
		require.NoError(t, err)
		createdWorkspaces = append(createdWorkspaces, workspace)
	}

	// Clean up workspaces after test
	defer func() {
		for _, workspace := range createdWorkspaces {
			err := conn.DeleteWorkspace(ctx, workspace.Id)
			require.NoError(t, err)
		}
	}()

	tests := []struct {
		name    string
		limit   int
		offset  int
		want    int
		wantErr bool
	}{
		{
			name:    "success - get all workspaces",
			limit:   numWorkspaces,
			offset:  0,
			want:    numWorkspaces,
			wantErr: false,
		},
		{
			name:    "success - get first 2 workspaces",
			limit:   2,
			offset:  0,
			want:    2,
			wantErr: false,
		},
		{
			name:    "success - get last 2 workspaces",
			limit:   2,
			offset:  numWorkspaces - 2,
			want:    2,
			wantErr: false,
		},
		{
			name:    "error - invalid limit",
			limit:   0,
			offset:  0,
			want:    0,
			wantErr: true,
		},
		{
			name:    "error - negative limit",
			limit:   -1,
			offset:  0,
			want:    0,
			wantErr: true,
		},
		{
			name:    "error - negative offset",
			limit:   1,
			offset:  -1,
			want:    0,
			wantErr: true,
		},
		{
			name:    "success - offset beyond available workspaces",
			limit:   1,
			offset:  numWorkspaces + 1,
			want:    0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspaces, err := conn.ListWorkspaces(ctx, tc.Account.Id, tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, workspaces, tt.want)

			if tt.want > 0 {
				// Verify workspaces are returned in the correct order
				for i := 0; i < tt.want; i++ {
					expectedWorkspace := createdWorkspaces[tt.offset+i]
					actualWorkspace := workspaces[i]
					assert.Equal(t, expectedWorkspace.Id, actualWorkspace.Id)
					assert.Equal(t, expectedWorkspace.Name, actualWorkspace.Name)
				}
			}
		})
	}
}
