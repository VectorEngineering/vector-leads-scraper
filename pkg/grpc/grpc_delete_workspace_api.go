package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeleteWorkspace permanently removes a workspace and all associated resources.
// This action cannot be undone.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workspace ID to delete
//
// Returns:
//   - DeleteWorkspaceResponse: Confirmation of deletion
//   - error: Any error encountered during deletion
//
// Example:
//
//	resp, err := server.DeleteWorkspace(ctx, &DeleteWorkspaceRequest{
//	    Id: 123,
//	})
func (s *Server) DeleteWorkspace(ctx context.Context, req *proto.DeleteWorkspaceRequest) (*proto.DeleteWorkspaceResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "delete-workspace")
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

	logger.Info("deleting workspace", zap.Uint64("workspace_id", req.Id))

	// Delete the workspace using the database client
	err := s.db.DeleteWorkspace(ctx, req.Id)
	if err != nil {
		logger.Error("failed to delete workspace", zap.Error(err))
		if err == database.ErrInvalidInput {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		return nil, status.Error(codes.Internal, "failed to delete workspace")
	}

	return &proto.DeleteWorkspaceResponse{}, nil
} 