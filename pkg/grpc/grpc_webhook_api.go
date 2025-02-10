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
//	    Webhook: &Webhook{
//	        Name: "Lead Updates",
//	        Url: "https://example.com/webhook",
//	        Events: []string{"lead.created", "lead.updated"},
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

	logger.Info("creating webhook", 
		zap.String("webhook_name", webhook.WebhookName),
		zap.String("url", webhook.Url))

	// Create the webhook using the database client
	result, err := s.db.CreateWebhookConfig(ctx, req.WorkspaceId, webhook)
	if err != nil {
		logger.Error("failed to create webhook", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create webhook: %s", err.Error())
	}

	return &proto.CreateWebhookResponse{
		Webhook: result,
	}, nil
}

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
//	    Id: 123,
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
		if err == database.ErrNotFound {
			return nil, status.Error(codes.NotFound, "webhook not found")
		}
		return nil, status.Error(codes.Internal, "failed to get webhook")
	}

	return &proto.GetWebhookResponse{
		Webhook: webhook,
	}, nil
}

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
//	    Webhook: &Webhook{
//	        Id: 123,
//	        Name: "Updated Webhook",
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
	result, err := s.db.UpdateWebhookConfig(ctx, req.WorkspaceId, webhook)
	if err != nil {
		logger.Error("failed to update webhook", zap.Error(err))
		if err == database.ErrNotFound {
			return nil, status.Error(codes.NotFound, "webhook not found")
		}
		return nil, status.Error(codes.Internal, "failed to update webhook")
	}

	return &proto.UpdateWebhookResponse{
		Webhook: result,
	}, nil
}

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
//	    Id: 123,
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
		if err == database.ErrNotFound {
			return nil, status.Error(codes.NotFound, "webhook not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete webhook")
	}

	return &proto.DeleteWebhookResponse{}, nil
}

// ListWebhooks retrieves a list of all webhook configurations in a workspace.
// Results can be filtered and paginated.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains filtering and pagination parameters
//
// Returns:
//   - ListWebhooksResponse: List of webhooks matching the filter criteria
//   - error: Any error encountered during listing
//
// Required permissions:
//   - list:webhook
//
// Example:
//
//	resp, err := server.ListWebhooks(ctx, &ListWebhooksRequest{
//	    WorkspaceId: 123,
//	    PageSize: 10,
//	    PageNumber: 1,
//	})
func (s *Server) ListWebhooks(ctx context.Context, req *proto.ListWebhooksRequest) (*proto.ListWebhooksResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "list-webhooks")
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

	logger.Info("listing webhooks",
		zap.Int32("page_size", req.PageSize),
		zap.Int32("page_number", req.PageNumber))

	// Use default page size if not specified
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 50 // Default page size
	}

	// Calculate offset based on page number
	pageNumber := req.PageNumber
	if pageNumber < 1 {
		pageNumber = 1
	}
	offset := pageSize * (int(pageNumber) - 1)

	// List webhooks using the database client
	webhooks, err := s.db.ListWebhookConfigs(ctx, req.WorkspaceId, pageSize, offset)
	if err != nil {
		logger.Error("failed to list webhooks", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list webhooks")
	}

	// Calculate next page number
	var nextPageNumber int32
	if len(webhooks) == pageSize {
		nextPageNumber = pageNumber + 1
	}

	return &proto.ListWebhooksResponse{
		Webhooks:       webhooks,
		NextPageNumber: nextPageNumber,
	}, nil
} 