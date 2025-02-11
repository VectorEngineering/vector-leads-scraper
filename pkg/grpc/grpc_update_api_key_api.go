package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateAPIKey modifies an existing API key's properties.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the API key ID and fields to update
//
// Returns:
//   - UpdateAPIKeyResponse: Contains the updated API key
//   - error: Any error encountered during update
//
// Required permissions:
//   - update:api_key
//
// Example:
//
//	resp, err := server.UpdateAPIKey(ctx, &UpdateAPIKeyRequest{
//	    ApiKey: &APIKey{
//	        Id: 123,
//	        Name: "Updated API Key Name",
//	        Scopes: []string{"read:leads"},
//	    },
//	})
func (s *Server) UpdateAPIKey(ctx context.Context, req *proto.UpdateAPIKeyRequest) (*proto.UpdateAPIKeyResponse, error) {
	ctx, logger, cleanup := s.setupRequest(ctx, "update-api-key")
	defer cleanup()

	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	apiKey := req.GetApiKey()
	if apiKey == nil {
		logger.Error("API key is nil")
		return nil, status.Error(codes.InvalidArgument, "API key is required")
	}

	logger.Info("updating API key", zap.Uint64("key_id", apiKey.GetId()))

	result, err := s.db.UpdateAPIKey(ctx, apiKey)
	if err != nil {
		logger.Error("failed to update API key", zap.Error(err))
		if err == database.ErrNotFound {
			return nil, status.Error(codes.NotFound, "API key not found")
		}
		return nil, status.Error(codes.Internal, "failed to update API key")
	}

	return &proto.UpdateAPIKeyResponse{
		ApiKey: result,
	}, nil
} 