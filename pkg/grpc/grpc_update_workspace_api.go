package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateWorkspace modifies an existing workspace configuration.
// This can include changing the name, description, or other settings.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workspace ID and fields to update
//
// Returns:
//   - UpdateWorkspaceResponse: Updated workspace information
//   - error: Any error encountered during update
//
// Example:
//
//	resp, err := server.UpdateWorkspace(ctx, &UpdateWorkspaceRequest{
//	    Workspace: &Workspace{
//	        Id: 123,
//	        Name: "Updated Team Name",
//	        Description: "Updated description",
//	    },
//	})
func (s *Server) UpdateWorkspace(ctx context.Context, req *proto.UpdateWorkspaceRequest) (*proto.UpdateWorkspaceResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "update-workspace")
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

	// Extract workspace details
	workspace := req.GetWorkspace()
	if workspace == nil {
		logger.Error("workspace is nil")
		return nil, status.Error(codes.InvalidArgument, "workspace is required")
	}

	logger.Info("updating workspace",
		zap.Uint64("workspace_id", workspace.Id),
		zap.String("name", workspace.Name))

	// Update the workspace using the database client
	result, err := s.db.UpdateWorkspace(ctx, workspace)
	if err != nil {
		logger.Error("failed to update workspace", zap.Error(err))
		if err == database.ErrInvalidInput {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		if err == database.ErrWorkspaceDoesNotExist || err.Error() == "failed to get workspace: workspace does not exist" {
			return nil, status.Error(codes.NotFound, "workspace not found")
		}
		return nil, status.Error(codes.Internal, "failed to update workspace")
	}

	return &proto.UpdateWorkspaceResponse{
		Workspace: result,
	}, nil
}
