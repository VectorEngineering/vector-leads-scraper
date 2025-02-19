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

func TestServer_GetLead(t *testing.T) {
	testCtx := initializeLeadTestContext(t)
	defer testCtx.Cleanup()

	// Create a test lead first
	lead := testutils.GenerateRandomLead()

	// create a job and tie it to a workspace
	job := testutils.GenerateRandomizedScrapingJob()
	createdJob, err := MockServer.db.CreateScrapingJob(context.Background(), testCtx.Workspace.Id, job)
	require.NoError(t, err)
	require.NotNil(t, createdJob)

	// create a test lead with the same id as the lead
	createdLead, err := MockServer.db.CreateLead(context.Background(), createdJob.Id, lead)
	require.NoError(t, err)
	require.NotNil(t, createdLead)

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
				WorkspaceId:    testCtx.Workspace.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
				LeadId:         createdLead.Id,
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.GetLeadResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Lead)
				assert.Equal(t, createdLead.Id, resp.Lead.Id)
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
				WorkspaceId:    testCtx.Workspace.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
				LeadId:         999999,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "error - invalid lead ID",
			req: &proto.GetLeadRequest{
				WorkspaceId:    testCtx.Workspace.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
				LeadId:         0,
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
