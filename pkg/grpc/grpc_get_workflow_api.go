package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetWorkflow retrieves detailed information about a specific workflow.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workflow ID to retrieve
//
// Returns:
//   - GetWorkflowResponse: Detailed workflow information
//   - error: Any error encountered during retrieval
//
// Required permissions:
//   - read:workflow
//
// Example:
//
//	resp, err := server.GetWorkflow(ctx, &GetWorkflowRequest{
//	    Id: 123,
//	    WorkspaceId: 456,
//	})
func (s *Server) GetWorkflow(ctx context.Context, req *proto.GetWorkflowRequest) (*proto.GetWorkflowResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "get-workflow")
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

	logger.Info("getting workflow", zap.Uint64("workflow_id", req.Id))

	// Get the workflow using the database client
	workflow, err := s.db.GetScrapingWorkflow(ctx, req.Id)
	if err != nil {
		logger.Error("failed to get workflow", zap.Error(err))
		if err == database.ErrInvalidInput {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		return nil, status.Error(codes.Internal, "failed to get workflow")
	}

	return &proto.GetWorkflowResponse{
		Workflow: workflow,
	}, nil
} 