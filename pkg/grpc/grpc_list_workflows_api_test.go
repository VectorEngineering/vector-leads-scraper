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

func TestServer_ListWorkflows(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Create multiple test workflows
	for i := 0; i < 3; i++ {
		createResp, err := MockServer.CreateWorkflow(context.Background(), &proto.CreateWorkflowRequest{
			WorkspaceId: testCtx.Workspace.Id,
			Workflow: &proto.ScrapingWorkflow{
				CronExpression:           "0 0 * * *",
				Status:                   proto.WorkflowStatus_WORKFLOW_STATUS_ACTIVE,
				MaxRetries:               3,
				AlertEmails:              "alerts@example.com",
				GeoFencingRadius:         1000.0,
				GeoFencingLat:            37.7749,
				GeoFencingLon:            -122.4194,
				GeoFencingZoomMin:        10,
				GeoFencingZoomMax:        12,
				IncludeReviews:           true,
				IncludePhotos:            true,
				IncludeBusinessHours:     true,
				MaxReviewsPerBusiness:    100,
				OutputFormat:             proto.ScrapingWorkflow_OUTPUT_FORMAT_JSON,
				OutputDestination:        "s3://bucket/path",
				AnonymizePii:             true,
				NotificationSlackChannel: "#alerts",
				NotificationEmailGroup:   "team@example.com",
				QosMaxConcurrentRequests: 5,
				QosMaxRetries:            3,
				RespectRobotsTxt:         true,
				AcceptTermsOfService:     true,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, createResp)
		require.NotNil(t, createResp.Workflow)
	}

	tests := []struct {
		name    string
		req     *proto.ListWorkflowsRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success - no filters",
			req: &proto.ListWorkflowsRequest{
				WorkspaceId:    testCtx.Workspace.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
				PageSize:       10,
				PageNumber:     1,
			},
			wantErr: false,
		},
		{
			name: "success - with pagination",
			req: &proto.ListWorkflowsRequest{
				WorkspaceId:    testCtx.Workspace.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
				PageSize:       2,
				PageNumber:     1,
			},
			wantErr: false,
		},
		{
			name: "success - with default page size",
			req: &proto.ListWorkflowsRequest{
				WorkspaceId:    testCtx.Workspace.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
				PageSize:       0,
				PageNumber:     1,
			},
			wantErr: false,
		},
		{
			name: "success - second page",
			req: &proto.ListWorkflowsRequest{
				WorkspaceId:    testCtx.Workspace.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
				PageSize:       2,
				PageNumber:     2,
			},
			wantErr: false,
		},
		{
			name: "invalid page size",
			req: &proto.ListWorkflowsRequest{
				WorkspaceId:    testCtx.Workspace.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
				PageSize:       -1,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid page number",
			req: &proto.ListWorkflowsRequest{
				WorkspaceId:    testCtx.Workspace.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
				PageSize:       10,
				PageNumber:     -1,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing workspace ID",
			req: &proto.ListWorkflowsRequest{
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				AccountId:      testCtx.Account.Id,
				PageSize:       10,
				PageNumber:     1,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing organization ID",
			req: &proto.ListWorkflowsRequest{
				WorkspaceId: testCtx.Workspace.Id,
				TenantId:    testCtx.TenantId,
				AccountId:   testCtx.Account.Id,
				PageSize:    10,
				PageNumber:  1,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing tenant ID",
			req: &proto.ListWorkflowsRequest{
				WorkspaceId:    testCtx.Workspace.Id,
				OrganizationId: testCtx.Organization.Id,
				AccountId:      testCtx.Account.Id,
				PageSize:       10,
				PageNumber:     1,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing account ID",
			req: &proto.ListWorkflowsRequest{
				WorkspaceId:    testCtx.Workspace.Id,
				OrganizationId: testCtx.Organization.Id,
				TenantId:       testCtx.TenantId,
				PageSize:       10,
				PageNumber:     1,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.ListWorkflows(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Workflows)

			if tt.req.PageSize > 0 {
				assert.LessOrEqual(t, len(resp.Workflows), int(tt.req.PageSize))
			} else {
				// For default page size
				assert.LessOrEqual(t, len(resp.Workflows), 50)
			}

			// Verify pagination
			if tt.req.PageNumber > 1 {
				assert.Equal(t, tt.req.PageNumber+1, resp.NextPageNumber)
			}
		})
	}
}
