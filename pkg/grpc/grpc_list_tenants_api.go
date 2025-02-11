package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListTenants retrieves all tenants in the system.
// Results can be filtered and paginated based on request parameters.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains filtering and pagination parameters
//
// Returns:
//   - ListTenantsResponse: List of tenants matching the criteria
//   - error: Any error encountered during listing
//
// Required permissions:
//   - list:tenant (system admin only)
//
// Example:
//
//	resp, err := server.ListTenants(ctx, &ListTenantsRequest{
//	    PageSize: 50,
//	    StatusFilter: []string{"ACTIVE"},
//	})
func (s *Server) ListTenants(ctx context.Context, req *proto.ListTenantsRequest) (*proto.ListTenantsResponse, error) {
	var (
		err    error
		logger = s.logger.With(zap.String("method", "ListTenants"))
	)

	ctx, cancel := context.WithTimeout(ctx, s.config.RpcTimeout)
	defer cancel()

	txn := s.telemetry.GetTraceFromContext(ctx)
	defer txn.StartSegment("grpc-list-tenants").End()

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

	s.logger.Info("listing tenants",
		zap.Int32("page_size", req.GetPageSize()),
		zap.Int32("page_number", req.GetPageNumber()))

	// Use default page size if not specified
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 50 // Default page size
	}

	// Calculate offset based on page number
	pageNumber := req.GetPageNumber()
	if pageNumber < 1 {
		pageNumber = 1
	}
	offset := int(pageSize) * (int(pageNumber) - 1) // Subtract 1 since page numbers start at 1

	// List tenants using the database client
	results, err := s.db.ListTenants(ctx, &database.ListTenantsInput{
		Limit:          int(pageSize),
		Offset:         offset,
		OrganizationID: req.GetOrganizationId(),
	})
	if err != nil {
		s.logger.Error("failed to list tenants", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list tenants")
	}

	// Calculate next page number
	var nextPageNumber int32
	if len(results) == int(pageSize) {
		nextPageNumber = req.GetPageNumber() + 1
	}

	return &proto.ListTenantsResponse{
		Tenants:        results,
		NextPageNumber: nextPageNumber,
	}, nil
} 