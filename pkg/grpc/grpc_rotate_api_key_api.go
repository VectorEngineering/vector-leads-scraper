package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// import (
// 	"context"

// 	"github.com/Vector/vector-leads-scraper/internal/database"
// 	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
// 	"go.uber.org/zap"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )

// RotateAPIKey generates a new API key while invalidating the old one.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the API key ID to rotate
//
// Returns:
//   - RotateAPIKeyResponse: Contains the new API key
//   - error: Any error encountered during rotation
//
// Required permissions:
//   - rotate:api_key
//
// Example:
//
//	resp, err := server.RotateAPIKey(ctx, &RotateAPIKeyRequest{
//	    WorkspaceId: 123,
//	    KeyId: 456,
//	})
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

	logger.Info("rotating API key", zap.Uint64("key_id", req.GetKeyId()))

	// Get the existing API key first
	existingKey, err := s.db.GetAPIKey(ctx, req.GetKeyId())
	if err != nil {
		logger.Error("failed to get API key", zap.Error(err))
		if err == database.ErrNotFound {
			return nil, status.Error(codes.NotFound, "API key not found")
		}
		return nil, status.Error(codes.Internal, "failed to get API key")
	}

	// Create a new API key with the same properties
	newKey := &proto.APIKey{
		Name:                       existingKey.Name,
		Description:                existingKey.Description,
		Scopes:                     existingKey.Scopes,
		ExpiresAt:                  existingKey.ExpiresAt,
		MaxUses:                    existingKey.MaxUses,
		AllowedIps:                 existingKey.AllowedIps,
		RateLimit:                  existingKey.RateLimit,
		EnforceSigning:             existingKey.EnforceSigning,
		AllowedSignatureAlgorithms: existingKey.AllowedSignatureAlgorithms,
		EnforceMutualTls:           existingKey.EnforceMutualTls,
		AlertEmails:                existingKey.AlertEmails,
		AlertOnQuotaThreshold:      existingKey.AlertOnQuotaThreshold,
		QuotaAlertThreshold:        existingKey.QuotaAlertThreshold,
		RequestsPerSecond:          existingKey.RequestsPerSecond,
		RequestsPerDay:             existingKey.RequestsPerDay,
		ConcurrentRequests:         existingKey.ConcurrentRequests,
		EnforceHttps:               existingKey.EnforceHttps,
		RotationFrequencyDays:      existingKey.RotationFrequencyDays,
		DataClassification:         existingKey.DataClassification,
	}

	// Rotate the API key using the database client
	result, err := s.db.RotateAPIKey(ctx, req.GetWorkspaceId(), req.GetKeyId(), newKey)
	if err != nil {
		logger.Error("failed to rotate API key", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to rotate API key")
	}

	return &proto.RotateAPIKeyResponse{
		NewApiKey: result,
	}, nil
}
