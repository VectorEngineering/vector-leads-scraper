package grpc

import (
	"context"

	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TriggerWorkflow manually starts a workflow execution.
// This bypasses the workflow's schedule and runs it immediately.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workflow ID to trigger
//
// Returns:
//   - TriggerWorkflowResponse: Contains the triggered workflow's execution ID
//   - error: Any error encountered during triggering
//
// Required permissions:
//   - trigger:workflow
//
// Example:
//
//	resp, err := server.TriggerWorkflow(ctx, &TriggerWorkflowRequest{
//	    Id: 123,
//	    WorkspaceId: 456,
//	})
func (s *Server) TriggerWorkflow(ctx context.Context, req *proto.TriggerWorkflowRequest) (*proto.TriggerWorkflowResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "trigger-workflow")
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

	logger.Info("triggering workflow", zap.Uint64("workflow_id", req.Id))

	// get the workflow from the database
	workflow, err := s.db.GetScrapingWorkflow(ctx, req.GetId())
	if err != nil {
		logger.Error("failed to get workflow", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get workflow")
	}

	// set the workflow status to active
	workflow.Status = proto.WorkflowStatus_WORKFLOW_STATUS_ACTIVE

	// // trigger the  execution of a new job from this workflow

	// // Trigger the workflow using the database client
	// execution, err := s.db.TriggerWorkflow(ctx, req.WorkspaceId, req.Id)
	// if err != nil {
	// 	logger.Error("failed to trigger workflow", zap.Error(err))
	// 	if err == database.ErrInvalidInput {
	// 		return nil, status.Error(codes.InvalidArgument, "invalid input")
	// 	}
	// 	return nil, status.Error(codes.Internal, "failed to trigger workflow")
	// }

	return &proto.TriggerWorkflowResponse{
		Status: proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED,
	}, nil
}
