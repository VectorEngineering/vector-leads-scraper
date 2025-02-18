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
					GeoFencingZoomMax:        20,
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
					GeoFencingZoomMax: 15,
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
					GeoFencingZoomMin: 15,    // Min > Max
					GeoFencingZoomMax: 10,
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
					GeoFencingZoomMax:        20,
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
			GeoFencingZoomMax:        20,
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
			GeoFencingZoomMax:        20,
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
	}{
		{
			name: "success",
			req: &proto.UpdateWorkflowRequest{
				Workflow: &proto.ScrapingWorkflow{
					Id:                       createResp.Workflow.Id,
					CronExpression:           "0 0 * * *",
					Status:                   proto.WorkflowStatus_WORKFLOW_STATUS_ACTIVE,
					MaxRetries:               3,
					AlertEmails:              "alerts@example.com",
					GeoFencingRadius:         1000.0,
					GeoFencingLat:            37.7749,
					GeoFencingLon:            -122.4194,
					GeoFencingZoomMin:        10,
					GeoFencingZoomMax:        20,
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
			name:    "workflow_not_found",
			req:     &proto.UpdateWorkflowRequest{Workflow: &proto.ScrapingWorkflow{Id: 999999}},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "empty workflow id",
			req: &proto.UpdateWorkflowRequest{
				Workflow: &proto.ScrapingWorkflow{
					Id: 0,
				},
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
			name: "nil workflow",
			req: &proto.UpdateWorkflowRequest{
				Workflow: nil,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid cron expression",
			req: &proto.UpdateWorkflowRequest{
				Workflow: &proto.ScrapingWorkflow{
					Id:             createResp.Workflow.Id,
					CronExpression: "invalid",
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
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Workflow)

			// Verify the updates were applied
			getResp, err := MockServer.GetWorkflow(context.Background(), &proto.GetWorkflowRequest{
				WorkspaceId: testCtx.Workspace.Id,
				Id:          createResp.Workflow.Id,
			})
			require.NoError(t, err)
			require.NotNil(t, getResp)
			require.NotNil(t, getResp.Workflow)

			if tt.req.Workflow != nil {
				assert.Equal(t, tt.req.Workflow.CronExpression, getResp.Workflow.CronExpression)
				assert.Equal(t, tt.req.Workflow.Status, getResp.Workflow.Status)
				assert.Equal(t, tt.req.Workflow.MaxRetries, getResp.Workflow.MaxRetries)
				assert.Equal(t, tt.req.Workflow.AlertEmails, getResp.Workflow.AlertEmails)
				assert.Equal(t, tt.req.Workflow.OutputFormat, getResp.Workflow.OutputFormat)
				assert.Equal(t, tt.req.Workflow.OutputDestination, getResp.Workflow.OutputDestination)
			}
		})
	}
}

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
			GeoFencingZoomMax:        20,
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
			name:    "workflow_not_found",
			req:     &proto.DeleteWorkflowRequest{
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
				GeoFencingZoomMax:        20,
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
				PageNumber:      1,
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

// func TestServer_TriggerWorkflow(t *testing.T) {
// 	testCtx := initializeAPIKeyTestContext(t)
// 	defer testCtx.Cleanup()

// 	// Create a test workflow first
// 	createResp, err := MockServer.CreateWorkflow(context.Background(), &proto.CreateWorkflowRequest{
// 		WorkspaceId: testCtx.Workspace.Id,
// 		Workflow: &proto.ScrapingWorkflow{
// 			CronExpression:           "0 0 * * *",
// 			Status:                   proto.WorkflowStatus_WORKFLOW_STATUS_ACTIVE,
// 			MaxRetries:               3,
// 			AlertEmails:              "alerts@example.com",
// 			GeoFencingRadius:         1000.0,
// 			GeoFencingLat:            37.7749,
// 			GeoFencingLon:            -122.4194,
// 			GeoFencingZoomMin:        10,
// 			GeoFencingZoomMax:        20,
// 			IncludeReviews:           true,
// 			IncludePhotos:            true,
// 			IncludeBusinessHours:     true,
// 			MaxReviewsPerBusiness:    100,
// 			OutputFormat:             proto.ScrapingWorkflow_OUTPUT_FORMAT_JSON,
// 			OutputDestination:        "s3://bucket/path",
// 			AnonymizePii:             true,
// 			NotificationSlackChannel: "#alerts",
// 			NotificationEmailGroup:   "team@example.com",
// 			QosMaxConcurrentRequests: 5,
// 			QosMaxRetries:            3,
// 			RespectRobotsTxt:         true,
// 			AcceptTermsOfService:     true,
// 		},
// 	})
// 	require.NoError(t, err)
// 	require.NotNil(t, createResp)
// 	require.NotNil(t, createResp.Workflow)

// 	tests := []struct {
// 		name    string
// 		req     *proto.TriggerWorkflowRequest
// 		wantErr bool
// 		errCode codes.Code
// 	}{
// 		{
// 			name: "success",
// 			req: &proto.TriggerWorkflowRequest{
// 				Id:          createResp.Workflow.Id,
// 				WorkspaceId: testCtx.Workspace.Id,
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "workflow not found",
// 			req: &proto.TriggerWorkflowRequest{
// 				Id:          999999,
// 				WorkspaceId: testCtx.Workspace.Id,
// 			},
// 			wantErr: true,
// 			errCode: codes.NotFound,
// 		},
// 		{
// 			name: "empty workflow id",
// 			req: &proto.TriggerWorkflowRequest{
// 				Id:          0,
// 				WorkspaceId: testCtx.Workspace.Id,
// 			},
// 			wantErr: true,
// 			errCode: codes.InvalidArgument,
// 		},
// 		{
// 			name:    "nil request",
// 			req:     nil,
// 			wantErr: true,
// 			errCode: codes.InvalidArgument,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			resp, err := MockServer.TriggerWorkflow(context.Background(), tt.req)
// 			if tt.wantErr {
// 				require.Error(t, err)
// 				st, ok := status.FromError(err)
// 				require.True(t, ok)
// 				assert.Equal(t, tt.errCode, st.Code())
// 				return
// 			}
// 			require.NoError(t, err)
// 			require.NotNil(t, resp)
// 		})
// 	}
// }
