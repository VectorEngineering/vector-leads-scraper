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

func TestServer_UpdateWorkflow(t *testing.T) {
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
		req     *proto.UpdateWorkflowRequest
		wantErr bool
		errCode codes.Code
		verify  func(t *testing.T, resp *proto.UpdateWorkflowResponse)
	}{
		{
			name: "success - update all fields",
			req: &proto.UpdateWorkflowRequest{
				Workflow: &proto.ScrapingWorkflow{
					Id:                       createResp.Workflow.Id,
					CronExpression:           "*/5 * * * *",
					Status:                   proto.WorkflowStatus_WORKFLOW_STATUS_PAUSED,
					MaxRetries:               5,
					AlertEmails:              "newalerts@example.com",
					GeoFencingRadius:         2000.0,
					GeoFencingLat:            38.7749,
					GeoFencingLon:            -123.4194,
					GeoFencingZoomMin:        12,
					GeoFencingZoomMax:        15,
					IncludeReviews:           true,
					IncludePhotos:            true,
					IncludeBusinessHours:     true,
					MaxReviewsPerBusiness:    200,
					OutputFormat:             proto.ScrapingWorkflow_OUTPUT_FORMAT_CSV,
					OutputDestination:        "s3://new-bucket/path",
					AnonymizePii:             true,
					NotificationSlackChannel: "#new-alerts",
					NotificationEmailGroup:   "newteam@example.com",
					QosMaxConcurrentRequests: 10,
					QosMaxRetries:            6,
					RespectRobotsTxt:         true,
					AcceptTermsOfService:     true,
				},
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.UpdateWorkflowResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Workflow)
				assert.Equal(t, "*/5 * * * *", resp.Workflow.CronExpression)
				assert.Equal(t, proto.WorkflowStatus_WORKFLOW_STATUS_PAUSED, resp.Workflow.Status)
				assert.Equal(t, int32(5), resp.Workflow.MaxRetries)
				assert.Equal(t, "newalerts@example.com", resp.Workflow.AlertEmails)
				assert.Equal(t, float32(2000.0), resp.Workflow.GeoFencingRadius)
				assert.Equal(t, float64(38.7749), resp.Workflow.GeoFencingLat)
				assert.Equal(t, float64(-123.4194), resp.Workflow.GeoFencingLon)
				assert.Equal(t, int32(12), resp.Workflow.GeoFencingZoomMin)
				assert.Equal(t, int32(15), resp.Workflow.GeoFencingZoomMax)
				assert.Equal(t, true, resp.Workflow.IncludeReviews)
				assert.Equal(t, true, resp.Workflow.IncludePhotos)
				assert.Equal(t, true, resp.Workflow.IncludeBusinessHours)
				assert.Equal(t, int32(200), resp.Workflow.MaxReviewsPerBusiness)
				assert.Equal(t, proto.ScrapingWorkflow_OUTPUT_FORMAT_CSV, resp.Workflow.OutputFormat)
				assert.Equal(t, "s3://new-bucket/path", resp.Workflow.OutputDestination)
				assert.Equal(t, true, resp.Workflow.AnonymizePii)
				assert.Equal(t, "#new-alerts", resp.Workflow.NotificationSlackChannel)
				assert.Equal(t, "newteam@example.com", resp.Workflow.NotificationEmailGroup)
				assert.Equal(t, int32(10), resp.Workflow.QosMaxConcurrentRequests)
				assert.Equal(t, int32(6), resp.Workflow.QosMaxRetries)
				assert.Equal(t, true, resp.Workflow.RespectRobotsTxt)
				assert.Equal(t, true, resp.Workflow.AcceptTermsOfService)
			},
		},
		{
			name: "success - partial update",
			req: &proto.UpdateWorkflowRequest{
				Workflow: &proto.ScrapingWorkflow{
					Id:                createResp.Workflow.Id,
					CronExpression:    "0 */2 * * *",
					Status:            proto.WorkflowStatus_WORKFLOW_STATUS_PAUSED,
					MaxRetries:        3,
					GeoFencingZoomMin: 10,
					GeoFencingZoomMax: 15,
				},
			},
			wantErr: false,
			verify: func(t *testing.T, resp *proto.UpdateWorkflowResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Workflow)
				assert.Equal(t, "0 */2 * * *", resp.Workflow.CronExpression)
				assert.Equal(t, proto.WorkflowStatus_WORKFLOW_STATUS_PAUSED, resp.Workflow.Status)
				assert.Equal(t, int32(3), resp.Workflow.MaxRetries)
				assert.Equal(t, int32(10), resp.Workflow.GeoFencingZoomMin)
				assert.Equal(t, int32(15), resp.Workflow.GeoFencingZoomMax)
			},
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "nil workflow",
			req: &proto.UpdateWorkflowRequest{
				Workflow: nil,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "workflow not found",
			req: &proto.UpdateWorkflowRequest{
				Workflow: &proto.ScrapingWorkflow{
					Id:             999999,
					CronExpression: "0 0 * * *",
					Status:         proto.WorkflowStatus_WORKFLOW_STATUS_ACTIVE,
				},
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "invalid cron expression",
			req: &proto.UpdateWorkflowRequest{
				Workflow: &proto.ScrapingWorkflow{
					Id:             createResp.Workflow.Id,
					CronExpression: "invalid",
					Status:         proto.WorkflowStatus_WORKFLOW_STATUS_ACTIVE,
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid geo coordinates",
			req: &proto.UpdateWorkflowRequest{
				Workflow: &proto.ScrapingWorkflow{
					Id:             createResp.Workflow.Id,
					CronExpression: "0 0 * * *",
					Status:         proto.WorkflowStatus_WORKFLOW_STATUS_ACTIVE,
					GeoFencingLat:  91.0,  // Invalid latitude (>90)
					GeoFencingLon:  180.1, // Invalid longitude (>180)
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid zoom range",
			req: &proto.UpdateWorkflowRequest{
				Workflow: &proto.ScrapingWorkflow{
					Id:                createResp.Workflow.Id,
					CronExpression:    "0 0 * * *",
					Status:            proto.WorkflowStatus_WORKFLOW_STATUS_ACTIVE,
					GeoFencingLat:     37.7749,
					GeoFencingLon:     -122.4194,
					GeoFencingZoomMin: 15, // Min > Max
					GeoFencingZoomMax: 10,
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid max retries",
			req: &proto.UpdateWorkflowRequest{
				Workflow: &proto.ScrapingWorkflow{
					Id:             createResp.Workflow.Id,
					CronExpression: "0 0 * * *",
					Status:         proto.WorkflowStatus_WORKFLOW_STATUS_ACTIVE,
					MaxRetries:     -1, // Invalid negative value
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid QoS settings",
			req: &proto.UpdateWorkflowRequest{
				Workflow: &proto.ScrapingWorkflow{
					Id:                       createResp.Workflow.Id,
					CronExpression:           "0 0 * * *",
					Status:                   proto.WorkflowStatus_WORKFLOW_STATUS_ACTIVE,
					QosMaxConcurrentRequests: -1, // Invalid negative value
					QosMaxRetries:            -1, // Invalid negative value
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.UpdateWorkflow(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Workflow)

			if tt.verify != nil {
				tt.verify(t, resp)
			}

			// Verify the updates were applied by getting the workflow
			getResp, err := MockServer.GetWorkflow(context.Background(), &proto.GetWorkflowRequest{
				WorkspaceId: testCtx.Workspace.Id,
				Id:          createResp.Workflow.Id,
			})
			require.NoError(t, err)
			require.NotNil(t, getResp)
			require.NotNil(t, getResp.Workflow)

			if tt.verify != nil {
				tt.verify(t, &proto.UpdateWorkflowResponse{Workflow: getResp.Workflow})
			}
		})
	}
}
