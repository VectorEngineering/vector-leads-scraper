package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateTenantAPIKey creates a new API key for a tenant.
func (s *Server) CreateTenantAPIKey(ctx context.Context, req *proto.CreateTenantAPIKeyRequest) (*proto.CreateTenantAPIKeyResponse, error) {
	ctx, logger, cleanup := s.setupRequest(ctx, "create-tenant-api-key")
	defer cleanup()

	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	logger.Info("creating tenant API key")

	apiKey, err := s.db.CreateTenantAPIKey(ctx, req.GetTenantId(), req.GetName(), req.GetScopes())
	if err != nil {
		logger.Error("failed to create tenant API key", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create tenant API key")
	}

	return &proto.CreateTenantAPIKeyResponse{
		ApiKey: apiKey,
	}, nil
}

// GetTenantAPIKey retrieves information about a specific tenant API key.
func (s *Server) GetTenantAPIKey(ctx context.Context, req *proto.GetTenantAPIKeyRequest) (*proto.GetTenantAPIKeyResponse, error) {
	ctx, logger, cleanup := s.setupRequest(ctx, "get-tenant-api-key")
	defer cleanup()

	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	logger.Info("getting tenant API key", zap.String("key_id", req.GetKeyId()))

	apiKey, err := s.db.GetTenantAPIKey(ctx, req.GetTenantId(), req.GetKeyId())
	if err != nil {
		logger.Error("failed to get tenant API key", zap.Error(err))
		if err == database.ErrNotFound {
			return nil, status.Error(codes.NotFound, "tenant API key not found")
		}
		return nil, status.Error(codes.Internal, "failed to get tenant API key")
	}

	return &proto.GetTenantAPIKeyResponse{
		ApiKey: apiKey,
	}, nil
}

// UpdateTenantAPIKey modifies an existing tenant API key's properties.
func (s *Server) UpdateTenantAPIKey(ctx context.Context, req *proto.UpdateTenantAPIKeyRequest) (*proto.UpdateTenantAPIKeyResponse, error) {
	ctx, logger, cleanup := s.setupRequest(ctx, "update-tenant-api-key")
	defer cleanup()

	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	logger.Info("updating tenant API key", zap.String("key_id", req.GetKeyId()))

	apiKey, err := s.db.UpdateTenantAPIKey(ctx, req.GetTenantId(), req.GetKeyId(), req.GetName(), req.GetScopes())
	if err != nil {
		logger.Error("failed to update tenant API key", zap.Error(err))
		if err == database.ErrNotFound {
			return nil, status.Error(codes.NotFound, "tenant API key not found")
		}
		return nil, status.Error(codes.Internal, "failed to update tenant API key")
	}

	return &proto.UpdateTenantAPIKeyResponse{
		ApiKey: apiKey,
	}, nil
}

// DeleteTenantAPIKey permanently removes a tenant API key.
func (s *Server) DeleteTenantAPIKey(ctx context.Context, req *proto.DeleteTenantAPIKeyRequest) (*proto.DeleteTenantAPIKeyResponse, error) {
	ctx, logger, cleanup := s.setupRequest(ctx, "delete-tenant-api-key")
	defer cleanup()

	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	logger.Info("deleting tenant API key", zap.String("key_id", req.GetKeyId()))

	err := s.db.DeleteTenantAPIKey(ctx, req.GetTenantId(), req.GetKeyId())
	if err != nil {
		logger.Error("failed to delete tenant API key", zap.Error(err))
		if err == database.ErrNotFound {
			return nil, status.Error(codes.NotFound, "tenant API key not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete tenant API key")
	}

	return &proto.DeleteTenantAPIKeyResponse{}, nil
}

// ListTenantAPIKeys retrieves all API keys for a tenant.
func (s *Server) ListTenantAPIKeys(ctx context.Context, req *proto.ListTenantAPIKeysRequest) (*proto.ListTenantAPIKeysResponse, error) {
	ctx, logger, cleanup := s.setupRequest(ctx, "list-tenant-api-keys")
	defer cleanup()

	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	logger.Info("listing tenant API keys")

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 50 // Default page size
	}

	pageNumber := req.GetPageNumber()
	if pageNumber < 1 {
		pageNumber = 1
	}
	offset := pageSize * (int(pageNumber) - 1)

	apiKeys, err := s.db.ListTenantAPIKeys(ctx, req.GetTenantId(), pageSize, offset)
	if err != nil {
		logger.Error("failed to list tenant API keys", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list tenant API keys")
	}

	var nextPageNumber int32
	if len(apiKeys) == pageSize {
		nextPageNumber = pageNumber + 1
	}

	return &proto.ListTenantAPIKeysResponse{
		ApiKeys:       apiKeys,
		NextPageNumber: nextPageNumber,
	}, nil
}

// RotateTenantAPIKey generates a new tenant API key while invalidating the old one.
func (s *Server) RotateTenantAPIKey(ctx context.Context, req *proto.RotateTenantAPIKeyRequest) (*proto.RotateTenantAPIKeyResponse, error) {
	ctx, logger, cleanup := s.setupRequest(ctx, "rotate-tenant-api-key")
	defer cleanup()

	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	logger.Info("rotating tenant API key", zap.String("key_id", req.GetKeyId()))

	newApiKey, err := s.db.RotateTenantAPIKey(ctx, req.GetTenantId(), req.GetKeyId())
	if err != nil {
		logger.Error("failed to rotate tenant API key", zap.Error(err))
		if err == database.ErrNotFound {
			return nil, status.Error(codes.NotFound, "tenant API key not found")
		}
		return nil, status.Error(codes.Internal, "failed to rotate tenant API key")
	}

	return &proto.RotateTenantAPIKeyResponse{
		ApiKey: newApiKey,
	}, nil
} 