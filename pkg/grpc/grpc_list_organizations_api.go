package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListOrganizations retrieves all organizations in a tenant.
// Results can be filtered and paginated based on request parameters.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains filtering and pagination parameters
//
// Returns:
//   - ListOrganizationsResponse: List of organizations matching the criteria
//   - error: Any error encountered during listing
//
// Required permissions:
//   - list:organization
//
// Example:
//
//	resp, err := server.ListOrganizations(ctx, &ListOrganizationsRequest{
//	    TenantId: "tenant_123",
//	    PageSize: 50,
//	    StatusFilter: []string{"ACTIVE"},
//	})
func (s *Server) ListOrganizations(ctx context.Context, req *proto.ListOrganizationsRequest) (*proto.ListOrganizationsResponse, error) {
	var (
		err error
	)

	// initialize the trace
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "list-organizations")
	defer cleanup()

	// ensure the request is not nil
	if req == nil {
		// log the error conditions encountered
		logger.Error("request cannot be nil as it is required")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// validate request and all containing fields
	if err = req.ValidateAll(); err != nil {
		// log the error conditions encountered
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	s.logger.Info("listing organizations",
		zap.Int32("page_size", req.GetPageSize()),
		zap.Int32("page_number", req.GetPageNumber()))

	// Use default page size if not specified
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 50 // Default page size
	}

	// Calculate offset based on page number
	offset := int(pageSize) * int(req.GetPageNumber())

	// List organizations using the database client
	organizations, err := s.db.ListOrganizations(ctx, &database.ListOrganizationsInput{
		Limit:  int(pageSize),
		Offset: offset,
	})
	if err != nil {
		s.logger.Error("failed to list organizations", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list organizations")
	}

	// Calculate next page number
	var nextPageNumber int32
	if len(organizations) == int(pageSize) {
		nextPageNumber = req.GetPageNumber() + 1
	}

	return &proto.ListOrganizationsResponse{
		Organizations:  organizations,
		NextPageNumber: nextPageNumber,
	}, nil
} 