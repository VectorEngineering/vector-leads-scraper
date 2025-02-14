package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateWebhook creates a new webhook configuration for receiving notifications
// about scraping job events and lead updates.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains webhook configuration including URL and event types
//
// Returns:
//   - CreateWebhookResponse: Contains the created webhook's ID and configuration
//   - error: Any error encountered during creation
//
// Required permissions:
//   - create:webhook
//
// Example:
//
//	resp, err := server.CreateWebhook(ctx, &CreateWebhookRequest{
//	    Webhook: &WebhookConfig{
//	        WebhookName: "Lead Updates",
//	        Url: "https://example.com/webhook",
//	        AuthType: "basic",
//	        AuthToken: "test-token",
//	        MaxRetries: 3,
//	        VerifySsl: true,
//	    },
//	    WorkspaceId: 123,
//	})
func (s *Server) CreateWebhook(ctx context.Context, req *proto.CreateWebhookRequest) (*proto.CreateWebhookResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "create-webhook")
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

	logger.Info("creating webhook", zap.String("webhook_name", webhook.WebhookName))

	// Create the webhook using the database client
	result, err := s.db.CreateWebhookConfig(ctx, req.WorkspaceId, webhook)
	if err != nil {
		logger.Error("failed to create webhook", zap.Error(err))
		if err == database.ErrNotFound {
			return nil, status.Error(codes.NotFound, "workspace not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to create webhook: %s", err.Error())
	}

	return &proto.CreateWebhookResponse{
		Webhook: result,
	}, nil
}
