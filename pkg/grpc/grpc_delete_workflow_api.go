package grpc

import (
	"context"
	"strings"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeleteWorkflow permanently removes a workflow configuration.
// This action cannot be undone.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workflow ID to delete
//
// Returns:
//   - DeleteWorkflowResponse: Confirmation of deletion
//   - error: Any error encountered during deletion
//
// Required permissions:
//   - delete:workflow
//
// Example:
//
//	resp, err := server.DeleteWorkflow(ctx, &DeleteWorkflowRequest{
//	    Id: 123,
//	    WorkspaceId: 456,
//	})
func (s *Server) DeleteWorkflow(ctx context.Context, req *proto.DeleteWorkflowRequest) (*proto.DeleteWorkflowResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "delete-workflow")
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

	logger.Info("deleting workflow", zap.Uint64("workflow_id", req.Id))

	// Delete the workflow using the database client
	err := s.db.DeleteScrapingWorkflow(ctx, req.Id)
	if err != nil {
		logger.Error("failed to delete workflow", zap.Error(err))
		if err == database.ErrInvalidInput {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		if strings.Contains(err.Error(), "record not found") {
			return nil, status.Error(codes.NotFound, "workflow not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete workflow")
	}

	return &proto.DeleteWorkflowResponse{}, nil
}
