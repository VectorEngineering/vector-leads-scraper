package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateAPIKey creates a new API key for accessing the service.
func (s *Server) CreateAPIKey(ctx context.Context, req *proto.CreateAPIKeyRequest) (*proto.CreateAPIKeyResponse, error) {
	ctx, logger, cleanup := s.setupRequest(ctx, "create-api-key")
	defer cleanup()

	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	logger.Info("creating API key")

	apiKey, err := s.db.CreateAPIKey(ctx, req.GetWorkspaceId(), req.GetName(), req.GetScopes())
	if err != nil {
		logger.Error("failed to create API key", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create API key")
	}

	return &proto.CreateAPIKeyResponse{
		ApiKey: apiKey,
	}, nil
}

// GetAPIKey retrieves information about a specific API key.
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

	logger.Info("getting API key", zap.String("key_id", req.GetKeyId()))

	apiKey, err := s.db.GetAPIKey(ctx, req.GetWorkspaceId(), req.GetKeyId())
	if err != nil {
		logger.Error("failed to get API key", zap.Error(err))
		if err == database.ErrNotFound {
			return nil, status.Error(codes.NotFound, "API key not found")
		}
		return nil, status.Error(codes.Internal, "failed to get API key")
	}

	return &proto.GetAPIKeyResponse{
		ApiKey: apiKey,
	}, nil
}

// UpdateAPIKey modifies an existing API key's properties.
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

	logger.Info("updating API key", zap.String("key_id", req.GetKeyId()))

	apiKey, err := s.db.UpdateAPIKey(ctx, req.GetWorkspaceId(), req.GetKeyId(), req.GetName(), req.GetScopes())
	if err != nil {
		logger.Error("failed to update API key", zap.Error(err))
		if err == database.ErrNotFound {
			return nil, status.Error(codes.NotFound, "API key not found")
		}
		return nil, status.Error(codes.Internal, "failed to update API key")
	}

	return &proto.UpdateAPIKeyResponse{
		ApiKey: apiKey,
	}, nil
}

// DeleteAPIKey permanently removes an API key.
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

	logger.Info("deleting API key", zap.String("key_id", req.GetKeyId()))

	err := s.db.DeleteAPIKey(ctx, req.GetWorkspaceId(), req.GetKeyId())
	if err != nil {
		logger.Error("failed to delete API key", zap.Error(err))
		if err == database.ErrNotFound {
			return nil, status.Error(codes.NotFound, "API key not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete API key")
	}

	return &proto.DeleteAPIKeyResponse{}, nil
}

// ListAPIKeys retrieves all API keys for a workspace.
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

	apiKeys, err := s.db.ListAPIKeys(ctx, req.GetWorkspaceId(), pageSize, offset)
	if err != nil {
		logger.Error("failed to list API keys", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list API keys")
	}

	var nextPageNumber int32
	if len(apiKeys) == pageSize {
		nextPageNumber = pageNumber + 1
	}

	return &proto.ListAPIKeysResponse{
		ApiKeys:       apiKeys,
		NextPageNumber: nextPageNumber,
	}, nil
}

// RotateAPIKey generates a new API key while invalidating the old one.
func (s *Server) RotateAPIKey(ctx context.Context, req *proto.RotateAPIKeyRequest) (*proto.RotateAPIKeyResponse, error) {
	ctx, logger, cleanup := s.setupRequest(ctx, "rotate-api-key")
	defer cleanup()

	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	logger.Info("rotating API key", zap.String("key_id", req.GetKeyId()))

	newApiKey, err := s.db.RotateAPIKey(ctx, req.GetWorkspaceId(), req.GetKeyId())
	if err != nil {
		logger.Error("failed to rotate API key", zap.Error(err))
		if err == database.ErrNotFound {
			return nil, status.Error(codes.NotFound, "API key not found")
		}
		return nil, status.Error(codes.Internal, "failed to rotate API key")
	}

	return &proto.RotateAPIKeyResponse{
		ApiKey: newApiKey,
	}, nil
} 