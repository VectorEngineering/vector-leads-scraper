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

func TestServer_ListLeads(t *testing.T) {
	testCtx := initializeLeadTestContext(t)
	defer testCtx.Cleanup()

	// create a job and tie it to a workspace
	job := testutils.GenerateRandomizedScrapingJob()
	createdJob, err := MockServer.db.CreateScrapingJob(context.Background(), testCtx.Workspace.Id, job)
	require.NoError(t, err)
	require.NotNil(t, createdJob)

	// Create some test leads first
	numLeads := 5
	leads := make([]*proto.Lead, 0, numLeads)
	for i := 0; i < numLeads; i++ {
		lead := testutils.GenerateRandomLead()

		createdLead, err := MockServer.db.CreateLead(context.Background(), createdJob.Id, lead)
		require.NoError(t, err)
		require.NotNil(t, createdLead)

		leads = append(leads, createdLead)
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
				WorkspaceId:    testCtx.Workspace.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
				PageSize:       10,
				PageNumber:     1,
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.ListLeadsResponse) {
				require.NotNil(t, resp)
				assert.NotEmpty(t, resp.Leads)
			},
		},
		{
			name: "success - pagination",
			req: &proto.ListLeadsRequest{
				WorkspaceId:    testCtx.Workspace.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
				PageSize:       2,
				PageNumber:     1,
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
				WorkspaceId:    testCtx.Workspace.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
				PageSize:       0,
				PageNumber:     1,
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
