package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateWorkflow modifies an existing workflow configuration.
// This can include changing the schedule, steps, or other settings.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workflow ID and fields to update
//
// Returns:
//   - UpdateWorkflowResponse: Updated workflow information
//   - error: Any error encountered during update
//
// Required permissions:
//   - update:workflow
//
// Example:
//
//	resp, err := server.UpdateWorkflow(ctx, &UpdateWorkflowRequest{
//	    Workflow: &Workflow{
//	        Id: 123,
//	        Name: "Updated Workflow",
//	        Schedule: "0 0 * * *",
//	    },
//	    WorkspaceId: 456,
//	})
func (s *Server) UpdateWorkflow(ctx context.Context, req *proto.UpdateWorkflowRequest) (*proto.UpdateWorkflowResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "update-workflow")
	defer cleanup()

	// Check for nil request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validate the request
	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	// Extract workflow details
	workflow := req.GetWorkflow()
	if workflow == nil {
		logger.Error("workflow is nil")
		return nil, status.Error(codes.InvalidArgument, "workflow is required")
	}

	// Validate schedule if provided
	if workflow.CronExpression != "" && !isValidCronExpression(workflow.CronExpression) {
		logger.Error("invalid cron expression", zap.String("schedule", workflow.CronExpression))
		return nil, status.Error(codes.InvalidArgument, "invalid cron expression")
	}

	logger.Info("updating workflow",
		zap.Uint64("workflow_id", workflow.Id),
		zap.String("name", workflow.Name))

	// Update the workflow using the database client
	result, err := s.db.UpdateScrapingWorkflow(ctx, workflow)
	if err != nil {
		logger.Error("failed to update workflow", zap.Error(err))
		if err == database.ErrInvalidInput {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		return nil, status.Error(codes.Internal, "failed to update workflow")
	}

	return &proto.UpdateWorkflowResponse{
		Workflow: result,
	}, nil
}
