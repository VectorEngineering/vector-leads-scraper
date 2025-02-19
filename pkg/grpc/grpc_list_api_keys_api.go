package grpc

import (
	"context"

	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListAPIKeys retrieves all API keys for a workspace.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains pagination parameters
//
// Returns:
//   - ListAPIKeysResponse: List of API keys
//   - error: Any error encountered during listing
//
// Required permissions:
//   - list:api_key
//
// Example:
//
//	resp, err := server.ListAPIKeys(ctx, &ListAPIKeysRequest{
//	    WorkspaceId: 123,
//	    PageSize: 10,
//	    PageNumber: 1,
//	})
func (s *Server) ListAPIKeys(ctx context.Context, req *proto.ListAPIKeysRequest) (*proto.ListAPIKeysResponse, error) {
	ctx, logger, cleanup := s.setupRequest(ctx, "list-api-keys")
	defer cleanup()

	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	logger.Info("listing API keys")

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 50 // Default page size
	}

	pageNumber := req.GetPageNumber()
	if pageNumber < 1 {
		pageNumber = 1
	}
	offset := pageSize * (int(pageNumber) - 1)

	apiKeys, err := s.db.ListAPIKeys(ctx, int(pageSize), offset)
	if err != nil {
		logger.Error("failed to list API keys", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list API keys")
	}

	var nextPageNumber int32
	if len(apiKeys) == pageSize {
		nextPageNumber = pageNumber + 1
	}

	return &proto.ListAPIKeysResponse{
		ApiKeys:        apiKeys,
		NextPageNumber: nextPageNumber,
	}, nil
}
