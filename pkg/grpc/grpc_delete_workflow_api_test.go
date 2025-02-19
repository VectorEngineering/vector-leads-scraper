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

func TestServer_DeleteWorkflow(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Create a test workflow first
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

	tests := []struct {
		name    string
		req     *proto.DeleteWorkflowRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.DeleteWorkflowRequest{
				Id:          createResp.Workflow.Id,
				WorkspaceId: testCtx.Workspace.Id,
				OrgId:       testCtx.Organization.Id,
				TenantId:    testCtx.TenantId,
				AccountId:   testCtx.Account.Id,
			},
			wantErr: false,
		},
		{
			name: "workflow_not_found",
			req: &proto.DeleteWorkflowRequest{
				Id:          999999,
				WorkspaceId: testCtx.Workspace.Id,
				OrgId:       testCtx.Organization.Id,
				TenantId:    testCtx.TenantId,
				AccountId:   testCtx.Account.Id,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.DeleteWorkflow(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}
