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

func TestServer_GetWorkflow(t *testing.T) {
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
		req     *proto.GetWorkflowRequest
		wantErr bool
		errCode codes.Code
		verify  func(t *testing.T, resp *proto.GetWorkflowResponse)
	}{
		{
			name: "success",
			req: &proto.GetWorkflowRequest{
				WorkspaceId: testCtx.Workspace.Id,
				Id:          createResp.Workflow.Id,
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.GetWorkflowResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Workflow)
				assert.Equal(t, createResp.Workflow.Id, resp.Workflow.Id)
				assert.Equal(t, createResp.Workflow.CronExpression, resp.Workflow.CronExpression)
				assert.Equal(t, createResp.Workflow.Status, resp.Workflow.Status)
				assert.Equal(t, createResp.Workflow.MaxRetries, resp.Workflow.MaxRetries)
				assert.Equal(t, createResp.Workflow.AlertEmails, resp.Workflow.AlertEmails)
				assert.Equal(t, createResp.Workflow.GeoFencingRadius, resp.Workflow.GeoFencingRadius)
				assert.Equal(t, createResp.Workflow.GeoFencingLat, resp.Workflow.GeoFencingLat)
				assert.Equal(t, createResp.Workflow.GeoFencingLon, resp.Workflow.GeoFencingLon)
				assert.Equal(t, createResp.Workflow.GeoFencingZoomMin, resp.Workflow.GeoFencingZoomMin)
				assert.Equal(t, createResp.Workflow.GeoFencingZoomMax, resp.Workflow.GeoFencingZoomMax)
				assert.Equal(t, createResp.Workflow.IncludeReviews, resp.Workflow.IncludeReviews)
				assert.Equal(t, createResp.Workflow.IncludePhotos, resp.Workflow.IncludePhotos)
				assert.Equal(t, createResp.Workflow.IncludeBusinessHours, resp.Workflow.IncludeBusinessHours)
				assert.Equal(t, createResp.Workflow.MaxReviewsPerBusiness, resp.Workflow.MaxReviewsPerBusiness)
				assert.Equal(t, createResp.Workflow.OutputFormat, resp.Workflow.OutputFormat)
				assert.Equal(t, createResp.Workflow.OutputDestination, resp.Workflow.OutputDestination)
				assert.Equal(t, createResp.Workflow.AnonymizePii, resp.Workflow.AnonymizePii)
				assert.Equal(t, createResp.Workflow.NotificationSlackChannel, resp.Workflow.NotificationSlackChannel)
				assert.Equal(t, createResp.Workflow.NotificationEmailGroup, resp.Workflow.NotificationEmailGroup)
				assert.Equal(t, createResp.Workflow.QosMaxConcurrentRequests, resp.Workflow.QosMaxConcurrentRequests)
				assert.Equal(t, createResp.Workflow.QosMaxRetries, resp.Workflow.QosMaxRetries)
				assert.Equal(t, createResp.Workflow.RespectRobotsTxt, resp.Workflow.RespectRobotsTxt)
				assert.Equal(t, createResp.Workflow.AcceptTermsOfService, resp.Workflow.AcceptTermsOfService)
			},
		},
		{
			name: "workflow_not_found",
			req: &proto.GetWorkflowRequest{
				Id:          999999,
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "empty workflow id",
			req: &proto.GetWorkflowRequest{
				WorkspaceId: testCtx.Workspace.Id,
				Id:          0,
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
			name: "missing workspace id",
			req: &proto.GetWorkflowRequest{
				Id: createResp.Workflow.Id,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid workspace id",
			req: &proto.GetWorkflowRequest{
				Id:          createResp.Workflow.Id,
				WorkspaceId: 0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.GetWorkflow(context.Background(), tt.req)
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
