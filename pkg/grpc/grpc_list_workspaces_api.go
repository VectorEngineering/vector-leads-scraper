package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListWorkspaces retrieves a list of all workspaces accessible to the user.
// Results can be filtered and paginated.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains filtering and pagination parameters
//
// Returns:
//   - ListWorkspacesResponse: List of workspaces matching the filter criteria
//   - error: Any error encountered during listing
//
// Example:
//
//	resp, err := server.ListWorkspaces(ctx, &ListWorkspacesRequest{
//	    AccountId: 123,
//	    PageSize: 10,
//	    PageNumber: 1,
//	})
func (s *Server) ListWorkspaces(ctx context.Context, req *proto.ListWorkspacesRequest) (*proto.ListWorkspacesResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "list-workspaces")
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

	logger.Info("listing workspaces",
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

	// List workspaces using the database client
	workspaces, err := s.db.ListWorkspaces(ctx, req.AccountId, int(offset), pageSize)
	if err != nil {
		logger.Error("failed to list workspaces", zap.Error(err))
		if err == database.ErrInvalidInput {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		return nil, status.Error(codes.Internal, "failed to list workspaces")
	}

	return &proto.ListWorkspacesResponse{
		Workspaces:     workspaces,
		NextPageNumber: int32(pageSize + 1),
	}, nil
} 