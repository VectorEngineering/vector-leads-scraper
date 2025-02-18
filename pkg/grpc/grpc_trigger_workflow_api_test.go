package grpc

import (
	"context"
	"testing"

	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTriggerWorkflow(t *testing.T) {
	t.Run("successful trigger", func(t *testing.T) {
		// Initialize test context
		ctx := context.Background()

		// First create a workflow to trigger
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

		// Now trigger the workflow
		req := &proto.TriggerWorkflowRequest{
			Id:          createResp.Workflow.Id,
			WorkspaceId: 1,
		}

		resp, err := leadScraperClient.TriggerWorkflow(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED, resp.Status)
	})

	t.Run("invalid request - nil", func(t *testing.T) {
		ctx := context.Background()
		resp, err := leadScraperClient.TriggerWorkflow(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("invalid request - workflow not found", func(t *testing.T) {
		ctx := context.Background()
		req := &proto.TriggerWorkflowRequest{
			Id:          999999, // Non-existent workflow
			WorkspaceId: 1,
		}

		resp, err := leadScraperClient.TriggerWorkflow(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("invalid request - missing workspace ID", func(t *testing.T) {
		ctx := context.Background()
		req := &proto.TriggerWorkflowRequest{
			Id: 1,
		}

		resp, err := leadScraperClient.TriggerWorkflow(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("invalid request - missing workflow ID", func(t *testing.T) {
		ctx := context.Background()
		req := &proto.TriggerWorkflowRequest{
			WorkspaceId: 1,
		}

		resp, err := leadScraperClient.TriggerWorkflow(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("trigger paused workflow", func(t *testing.T) {
		ctx := context.Background()

		// Create and pause a workflow
		createResp, err := leadScraperClient.CreateWorkflow(ctx, &proto.CreateWorkflowRequest{
			WorkspaceId: 1,
			Workflow: &proto.ScrapingWorkflow{
				Name:              "Paused Workflow",
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

		// Pause the workflow
		pauseResp, err := leadScraperClient.PauseWorkflow(ctx, &proto.PauseWorkflowRequest{
			Id:          createResp.Workflow.Id,
			WorkspaceId: 1,
		})
		require.NoError(t, err)
		require.NotNil(t, pauseResp)

		// Try to trigger the paused workflow
		req := &proto.TriggerWorkflowRequest{
			Id:          createResp.Workflow.Id,
			WorkspaceId: 1,
		}

		resp, err := leadScraperClient.TriggerWorkflow(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED, resp.Status)
	})
} 