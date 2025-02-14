package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetWebhook retrieves detailed information about a specific webhook configuration.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the webhook ID to retrieve
//
// Returns:
//   - GetWebhookResponse: Detailed webhook information
//   - error: Any error encountered during retrieval
//
// Required permissions:
//   - read:webhook
//
// Example:
//
//	resp, err := server.GetWebhook(ctx, &GetWebhookRequest{
//	    WebhookId: 123,
//	    WorkspaceId: 456,
//	})
func (s *Server) GetWebhook(ctx context.Context, req *proto.GetWebhookRequest) (*proto.GetWebhookResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "get-webhook")
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

	logger.Info("getting webhook", zap.Uint64("webhook_id", req.WebhookId))

	// Get the webhook using the database client
	webhook, err := s.db.GetWebhookConfig(ctx, req.WorkspaceId, req.WebhookId)
	if err != nil {
		logger.Error("failed to get webhook", zap.Error(err))
		if err == database.ErrInvalidInput {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		return nil, status.Error(codes.Internal, "failed to get webhook")
	}

	return &proto.GetWebhookResponse{
		Webhook: webhook,
	}, nil
}
