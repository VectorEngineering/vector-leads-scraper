package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetWorkspace retrieves detailed information about a specific workspace.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workspace ID to retrieve
//
// Returns:
//   - GetWorkspaceResponse: Detailed workspace information
//   - error: Any error encountered during retrieval
//
// Example:
//
//	resp, err := server.GetWorkspace(ctx, &GetWorkspaceRequest{
//	    Id: 123,
//	})
func (s *Server) GetWorkspace(ctx context.Context, req *proto.GetWorkspaceRequest) (*proto.GetWorkspaceResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "get-workspace")
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

	logger.Info("getting workspace", zap.Uint64("workspace_id", req.Id))

	// Get the workspace using the database client
	workspace, err := s.db.GetWorkspace(ctx, req.Id)
	if err != nil {
		logger.Error("failed to get workspace", zap.Error(err))
		if err == database.ErrInvalidInput {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		return nil, status.Error(codes.Internal, "failed to get workspace")
	}

	return &proto.GetWorkspaceResponse{
		Workspace: workspace,
	}, nil
}
