package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeleteWebhook permanently removes a webhook configuration.
// This action cannot be undone.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the webhook ID to delete
//
// Returns:
//   - DeleteWebhookResponse: Confirmation of deletion
//   - error: Any error encountered during deletion
//
// Required permissions:
//   - delete:webhook
//
// Example:
//
//	resp, err := server.DeleteWebhook(ctx, &DeleteWebhookRequest{
//	    WebhookId: 123,
//	    WorkspaceId: 456,
//	})
func (s *Server) DeleteWebhook(ctx context.Context, req *proto.DeleteWebhookRequest) (*proto.DeleteWebhookResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "delete-webhook")
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

	logger.Info("deleting webhook", zap.Uint64("webhook_id", req.WebhookId))

	// Delete the webhook using the database client
	err := s.db.DeleteWebhookConfig(ctx, req.WorkspaceId, req.WebhookId, database.DeletionTypeSoft)
	if err != nil {
		logger.Error("failed to delete webhook", zap.Error(err))
		if err == database.ErrInvalidInput {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		return nil, status.Error(codes.Internal, "failed to delete webhook")
	}

	return &proto.DeleteWebhookResponse{}, nil
} 