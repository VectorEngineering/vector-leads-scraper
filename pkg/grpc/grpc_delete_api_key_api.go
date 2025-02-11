package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeleteAPIKey permanently removes an API key.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the API key ID to delete
//
// Returns:
//   - DeleteAPIKeyResponse: Confirmation of deletion
//   - error: Any error encountered during deletion
//
// Required permissions:
//   - delete:api_key
//
// Example:
//
//	resp, err := server.DeleteAPIKey(ctx, &DeleteAPIKeyRequest{
//	    WorkspaceId: 123,
//	    KeyId: 456,
//	})
func (s *Server) DeleteAPIKey(ctx context.Context, req *proto.DeleteAPIKeyRequest) (*proto.DeleteAPIKeyResponse, error) {
	ctx, logger, cleanup := s.setupRequest(ctx, "delete-api-key")
	defer cleanup()

	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	logger.Info("deleting API key", zap.Uint64("key_id", req.GetKeyId()))

	err := s.db.DeleteAPIKey(ctx, req.GetWorkspaceId())
	if err != nil {
		logger.Error("failed to delete API key", zap.Error(err))
		if err == database.ErrNotFound {
			return nil, status.Error(codes.NotFound, "API key not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete API key")
	}

	return &proto.DeleteAPIKeyResponse{}, nil
} 