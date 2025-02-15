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

// UpdateWebhook modifies an existing webhook configuration.
// This can include changing the URL, event types, or other settings.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the webhook ID and fields to update
//
// Returns:
//   - UpdateWebhookResponse: Updated webhook information
//   - error: Any error encountered during update
//
// Required permissions:
//   - update:webhook
//
// Example:
//
//	resp, err := server.UpdateWebhook(ctx, &UpdateWebhookRequest{
//	    Webhook: &WebhookConfig{
//	        Id: 123,
//	        WebhookName: "Updated Webhook",
//	        Events: []string{"lead.*"},
//	    },
//	    WorkspaceId: 456,
//	})
func (s *Server) UpdateWebhook(ctx context.Context, req *proto.UpdateWebhookRequest) (*proto.UpdateWebhookResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "update-webhook")
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

	// Extract webhook details
	webhook := req.GetWebhook()
	if webhook == nil {
		logger.Error("webhook is nil")
		return nil, status.Error(codes.InvalidArgument, "webhook is required")
	}

	logger.Info("updating webhook",
		zap.Uint64("webhook_id", webhook.Id),
		zap.String("webhook_name", webhook.WebhookName))

	// Update the webhook using the database client
	result, err := s.db.UpdateWebhookConfig(ctx, req.Webhook.GetId(), webhook)
	if err != nil {
		logger.Error("failed to update webhook", zap.Error(err))
		if err == database.ErrInvalidInput {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		return nil, status.Error(codes.Internal, "failed to update webhook")
	}

	return &proto.UpdateWebhookResponse{
		Webhook: result,
	}, nil
}
