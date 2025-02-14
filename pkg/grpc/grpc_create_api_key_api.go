package grpc

import (
	"context"

	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateAPIKey creates a new API key for accessing the service.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains API key creation details like name and scopes
//
// Returns:
//   - CreateAPIKeyResponse: Contains the created API key
//   - error: Any error encountered during creation
//
// Required permissions:
//   - create:api_key
//
// Example:
//
//	resp, err := server.CreateAPIKey(ctx, &CreateAPIKeyRequest{
//	    WorkspaceId: 123,
//	    Name: "Production API Key",
//	    Scopes: []string{"read:leads", "write:leads"},
//	})
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

	apiKey := &proto.APIKey{
		Name:                       req.GetName(),
		Description:                req.GetDescription(),
		Scopes:                     req.GetScopes(),
		ExpiresAt:                  req.GetExpiresAt(),
		MaxUses:                    req.GetMaxUses(),
		AllowedIps:                 req.GetAllowedIps(),
		RateLimit:                  req.GetRateLimit(),
		EnforceSigning:             req.GetEnforceSigning(),
		AllowedSignatureAlgorithms: req.GetAllowedSignatureAlgorithms(),
		EnforceMutualTls:           req.GetEnforceMutualTls(),
		AlertEmails:                req.GetAlertEmails(),
		AlertOnQuotaThreshold:      req.GetAlertOnQuotaThreshold(),
		QuotaAlertThreshold:        req.GetQuotaAlertThreshold(),
	}

	result, err := s.db.CreateAPIKey(ctx, apiKey)
	if err != nil {
		logger.Error("failed to create API key", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create API key")
	}

	return &proto.CreateAPIKeyResponse{
		ApiKey: result,
	}, nil
}
