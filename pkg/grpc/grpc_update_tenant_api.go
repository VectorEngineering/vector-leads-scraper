package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateTenant modifies tenant configuration.
// Allows updating various tenant settings including name, description,
// domain configuration, security policies, resource limits, and billing settings.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the tenant ID and fields to update
//
// Returns:
//   - UpdateTenantResponse: Updated tenant information
//   - error: Any error encountered during update
//
// Required permissions:
//   - update:tenant
//
// Example:
//
//	resp, err := server.UpdateTenant(ctx, &UpdateTenantRequest{
//	    TenantId: "tenant_123abc",
//	    Name: "Enterprise Customer Pro",
//	    Settings: &TenantSettings{
//	        MaxOrganizations: 20,
//	        Features: []string{"sso", "audit_logs", "advanced_security"},
//	    },
//	})
func (s *Server) UpdateTenant(ctx context.Context, req *proto.UpdateTenantRequest) (*proto.UpdateTenantResponse, error) {
	var (
		err error
	)

	// initialize the trace
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "update-tenant")
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

	if req.GetTenant() == nil {
		return nil, status.Error(codes.InvalidArgument, "tenant is required")
	}

	s.logger.Info("updating tenant", zap.Uint64("tenant_id", req.GetTenant().GetId()))

	// Update the tenant using the database client
	result, err := s.db.UpdateTenant(ctx, &database.UpdateTenantInput{
		ID:     req.GetTenant().GetId(),
		Tenant: req.GetTenant(),
	})
	if err != nil {
		s.logger.Error("failed to update tenant", zap.Error(err))
		if err == database.ErrTenantDoesNotExist {
			return nil, status.Error(codes.NotFound, "tenant does not exist")
		}
		return nil, status.Error(codes.Internal, "failed to update tenant")
	}

	return &proto.UpdateTenantResponse{
		Tenant: result,
	}, nil
}
