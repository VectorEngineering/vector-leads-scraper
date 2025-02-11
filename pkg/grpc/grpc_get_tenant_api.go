package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetTenant retrieves detailed information about a tenant.
// Returns comprehensive information including basic metadata,
// resource utilization, organization list, configuration settings,
// and billing status.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the tenant ID to retrieve
//
// Returns:
//   - GetTenantResponse: Detailed tenant information
//   - error: Any error encountered during retrieval
//
// Required permissions:
//   - read:tenant
//
// Example:
//
//	resp, err := server.GetTenant(ctx, &GetTenantRequest{
//	    TenantId: "tenant_123abc",
//	})
func (s *Server) GetTenant(ctx context.Context, req *proto.GetTenantRequest) (*proto.GetTenantResponse, error) {
	var (
		err error
	)

	// initialize the trace
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "get-tenant")
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

	if req.GetTenantId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	s.logger.Info("getting tenant", zap.Uint64("tenant_id", req.GetTenantId()))

	// Get the tenant using the database client
	result, err := s.db.GetTenant(ctx, &database.GetTenantInput{
		ID: req.GetTenantId(),
	})
	if err != nil {
		s.logger.Error("failed to get tenant", zap.Error(err))
		if err == database.ErrTenantDoesNotExist {
			return nil, status.Error(codes.NotFound, "tenant does not exist")
		}
		return nil, status.Error(codes.Internal, "failed to get tenant")
	}

	return &proto.GetTenantResponse{
		Tenant: result,
	}, nil
} 