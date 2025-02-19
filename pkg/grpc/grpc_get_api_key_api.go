package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetAPIKey retrieves information about a specific API key.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the API key ID to retrieve
//
// Returns:
//   - GetAPIKeyResponse: Contains the API key details
//   - error: Any error encountered during retrieval
//
// Required permissions:
//   - read:api_key
//
// Example:
//
//	resp, err := server.GetAPIKey(ctx, &GetAPIKeyRequest{
//	    WorkspaceId: 123,
//	    KeyId: 456,
//	})
func (s *Server) GetAPIKey(ctx context.Context, req *proto.GetAPIKeyRequest) (*proto.GetAPIKeyResponse, error) {
	ctx, logger, cleanup := s.setupRequest(ctx, "get-api-key")
	defer cleanup()

	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	logger.Info("getting API key", zap.Uint64("key_id", req.GetKeyId()))

	result, err := s.db.GetAPIKey(ctx, req.GetKeyId())
	if err != nil {
		logger.Error("failed to get API key", zap.Error(err))
		if err == database.ErrNotFound {
			return nil, status.Error(codes.NotFound, "API key not found")
		}
		return nil, status.Error(codes.Internal, "failed to get API key")
	}

	return &proto.GetAPIKeyResponse{
		ApiKey: result,
	}, nil
}
