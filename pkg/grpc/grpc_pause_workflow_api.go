package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PauseWorkflow temporarily stops a workflow from executing.
// The workflow can be resumed later.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workflow ID to pause
//
// Returns:
//   - PauseWorkflowResponse: Confirmation of pause
//   - error: Any error encountered during pausing
//
// Required permissions:
//   - pause:workflow
//
// Example:
//
//	resp, err := server.PauseWorkflow(ctx, &PauseWorkflowRequest{
//	    Id: 123,
//	    WorkspaceId: 456,
//	})
func (s *Server) PauseWorkflow(ctx context.Context, req *proto.PauseWorkflowRequest) (*proto.PauseWorkflowResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "pause-workflow")
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

	logger.Info("pausing workflow", zap.Uint64("workflow_id", req.Id))

	// get the scraping workflow by id
	scrapingWorkflow, err := s.db.GetScrapingWorkflow(ctx, req.Id)
	if err != nil {
		logger.Error("failed to get scraping workflow", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get scraping workflow")
	}

	// set the status to paused
	scrapingWorkflow.Status = proto.WorkflowStatus_WORKFLOW_STATUS_PAUSED
	
	// Pause the workflow using the database client
	updatedScrapingWorkflow, err := s.db.UpdateScrapingWorkflow(ctx, scrapingWorkflow)
	if err != nil {
		logger.Error("failed to pause workflow", zap.Error(err))
		if err == database.ErrInvalidInput {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		return nil, status.Error(codes.Internal, "failed to pause workflow")
	}

	return &proto.PauseWorkflowResponse{
		Workflow: updatedScrapingWorkflow,
	}, nil
} 