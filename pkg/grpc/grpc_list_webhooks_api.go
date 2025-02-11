package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
	offset := (pageNumber - 1) * int32(pageSize)

	// List webhooks using the database client
	webhooks, err := s.db.ListWebhookConfigs(ctx, req.WorkspaceId, pageSize, int(offset))
	if err != nil {
		logger.Error("failed to list webhooks", zap.Error(err))
		if err == database.ErrInvalidInput {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		return nil, status.Error(codes.Internal, "failed to list webhooks")
	}

	return &proto.ListWebhooksResponse{
		Webhooks: webhooks,
	}, nil
} 