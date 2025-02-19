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

func TestServer_CreateWorkflow(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	tests := []struct {
		name    string
		req     *proto.CreateWorkflowRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.CreateWorkflowRequest{
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
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "nil workflow",
			req: &proto.CreateWorkflowRequest{
				Workflow: nil,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid cron expression",
			req: &proto.CreateWorkflowRequest{
				WorkspaceId: testCtx.Workspace.Id,
				Workflow: &proto.ScrapingWorkflow{
					CronExpression: "invalid cron",
					Status:         proto.WorkflowStatus_WORKFLOW_STATUS_ACTIVE,
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing workspace ID",
			req: &proto.CreateWorkflowRequest{
				Workflow: &proto.ScrapingWorkflow{
					CronExpression: "0 0 * * *",
					Status:         proto.WorkflowStatus_WORKFLOW_STATUS_ACTIVE,
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid geo coordinates",
			req: &proto.CreateWorkflowRequest{
				WorkspaceId: testCtx.Workspace.Id,
				Workflow: &proto.ScrapingWorkflow{
					CronExpression:    "0 0 * * *",
					Status:            proto.WorkflowStatus_WORKFLOW_STATUS_ACTIVE,
					GeoFencingLat:     91.0,  // Invalid latitude (>90)
					GeoFencingLon:     180.1, // Invalid longitude (>180)
					GeoFencingZoomMin: 10,
					GeoFencingZoomMax: 12,
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid zoom range",
			req: &proto.CreateWorkflowRequest{
				WorkspaceId: testCtx.Workspace.Id,
				Workflow: &proto.ScrapingWorkflow{
					CronExpression:    "0 0 * * *",
					Status:            proto.WorkflowStatus_WORKFLOW_STATUS_ACTIVE,
					GeoFencingLat:     37.7749,
					GeoFencingLon:     -122.4194,
					GeoFencingZoomMin: 15, // Min > Max
					GeoFencingZoomMax: 12,
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid max retries",
			req: &proto.CreateWorkflowRequest{
				WorkspaceId: testCtx.Workspace.Id,
				Workflow: &proto.ScrapingWorkflow{
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
			req: &proto.CreateWorkflowRequest{
				WorkspaceId: testCtx.Workspace.Id,
				Workflow: &proto.ScrapingWorkflow{
					CronExpression:           "0 0 * * *",
					Status:                   proto.WorkflowStatus_WORKFLOW_STATUS_ACTIVE,
					QosMaxConcurrentRequests: -1, // Invalid negative value
					QosMaxRetries:            -1, // Invalid negative value
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "database error",
			req: &proto.CreateWorkflowRequest{
				WorkspaceId: 999999, // Non-existent workspace
				Workflow: &proto.ScrapingWorkflow{
					CronExpression:           "0 0 * * *",
					Status:                   proto.WorkflowStatus_WORKFLOW_STATUS_ACTIVE,
					MaxRetries:               3,
					GeoFencingRadius:         1000.0,
					GeoFencingLat:            37.7749,
					GeoFencingLon:            -122.4194,
					GeoFencingZoomMin:        10,
					GeoFencingZoomMax:        12,
					OutputFormat:             proto.ScrapingWorkflow_OUTPUT_FORMAT_JSON,
					OutputDestination:        "s3://bucket/path",
					QosMaxConcurrentRequests: 5,
					QosMaxRetries:            3,
					RespectRobotsTxt:         true,
					AcceptTermsOfService:     true,
				},
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.CreateWorkflow(context.Background(), tt.req)
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
			assert.NotEmpty(t, resp.Workflow.Id)

			if tt.req.Workflow != nil {
				workflow := tt.req.Workflow
				assert.Equal(t, workflow.CronExpression, resp.Workflow.CronExpression)
				assert.Equal(t, workflow.Status, resp.Workflow.Status)
				assert.Equal(t, workflow.MaxRetries, resp.Workflow.MaxRetries)
				assert.Equal(t, workflow.AlertEmails, resp.Workflow.AlertEmails)
				assert.Equal(t, workflow.GeoFencingRadius, resp.Workflow.GeoFencingRadius)
				assert.Equal(t, workflow.GeoFencingLat, resp.Workflow.GeoFencingLat)
				assert.Equal(t, workflow.GeoFencingLon, resp.Workflow.GeoFencingLon)
				assert.Equal(t, workflow.OutputFormat, resp.Workflow.OutputFormat)
				assert.Equal(t, workflow.OutputDestination, resp.Workflow.OutputDestination)
			}
		})
	}
}

func Test_isValidCronExpression(t *testing.T) {
	type args struct {
		expr string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidCronExpression(tt.args.expr); got != tt.want {
				t.Errorf("isValidCronExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}
