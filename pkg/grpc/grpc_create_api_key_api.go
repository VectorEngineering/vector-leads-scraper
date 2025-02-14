package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/constants"
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

	// TODO: need a function to check for the existence of tenant and org id in 1 call

	// make sure the workspace of interest exists
	if _, err := s.db.GetWorkspace(ctx, req.GetWorkspaceId()); err != nil {
		logger.Error("failed to get workspace", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get workspace")
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
		RequestsPerSecond:          constants.API_KEY_REQUESTS_PER_SECOND,
		RequestsPerDay:             constants.API_KEY_REQUESTS_PER_DAY,
		ConcurrentRequests:         constants.API_KEY_CONCURRENT_REQUESTS,
		EnforceHttps:               constants.API_KEY_ENFORCE_HTTPS,
		RotationFrequencyDays:      constants.API_KEY_ROTATION_FREQUENCY_DAYS,
		DataClassification:         constants.API_KEY_DEFAULT_DATA_CLASSIFICATION,
	}

	result, err := s.db.CreateAPIKey(ctx, req.GetWorkspaceId(), apiKey)
	if err != nil {
		logger.Error("failed to create API key", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create API key")
	}

	return &proto.CreateAPIKeyResponse{
		ApiKey: result,
	}, nil
}
