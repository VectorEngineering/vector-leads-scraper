package grpc

import (
	"context"
	"reflect"
	"testing"

	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPauseWorkflow(t *testing.T) {
	t.Run("successful pause", func(t *testing.T) {
		// Initialize test context
		ctx := context.Background()

		// First create a workflow to pause
		createResp, err := leadScraperClient.CreateWorkflow(ctx, &proto.CreateWorkflowRequest{
			WorkspaceId: 1,
			Workflow: &proto.ScrapingWorkflow{
				Name:              "Test Workflow",
				CronExpression:    "0 0 * * *",
				MaxRetries:        3,
				GeoFencingLat:     37.7749,
				GeoFencingLon:     -122.4194,
				GeoFencingZoomMin: 10,
				GeoFencingZoomMax: 15,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, createResp)
		require.NotNil(t, createResp.Workflow)

		// Now pause the workflow
		req := &proto.PauseWorkflowRequest{
			Id:          createResp.Workflow.Id,
			WorkspaceId: 1,
		}

		resp, err := leadScraperClient.PauseWorkflow(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Workflow)
		assert.Equal(t, proto.WorkflowStatus_WORKFLOW_STATUS_PAUSED, resp.Workflow.Status)
	})

	t.Run("invalid request - nil", func(t *testing.T) {
		ctx := context.Background()
		resp, err := leadScraperClient.PauseWorkflow(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("invalid request - workflow not found", func(t *testing.T) {
		ctx := context.Background()
		req := &proto.PauseWorkflowRequest{
			Id:          999999, // Non-existent workflow
			WorkspaceId: 1,
		}

		resp, err := leadScraperClient.PauseWorkflow(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("invalid request - missing workspace ID", func(t *testing.T) {
		ctx := context.Background()
		req := &proto.PauseWorkflowRequest{
			Id: 1,
		}

		resp, err := leadScraperClient.PauseWorkflow(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("invalid request - missing workflow ID", func(t *testing.T) {
		ctx := context.Background()
		req := &proto.PauseWorkflowRequest{
			WorkspaceId: 1,
		}

		resp, err := leadScraperClient.PauseWorkflow(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestServer_PauseWorkflow(t *testing.T) {
	type args struct {
		ctx context.Context
		req *proto.PauseWorkflowRequest
	}
	tests := []struct {
		name    string
		s       *Server
		args    args
		want    *proto.PauseWorkflowResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.PauseWorkflow(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.PauseWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.PauseWorkflow() = %v, want %v", got, tt.want)
			}
		})
	}
}
