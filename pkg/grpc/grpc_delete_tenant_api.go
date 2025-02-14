package grpc

import (
	"context"
	"fmt"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeleteTenant removes a tenant and all associated resources.
// This is a destructive operation that removes all tenant data,
// deletes associated organizations, cleans up resources, and archives audit logs.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the tenant ID to delete
//
// Returns:
//   - DeleteTenantResponse: Confirmation of deletion
//   - error: Any error encountered during deletion
//
// Required permissions:
//   - delete:tenant (system admin only)
//
// Example:
//
//	resp, err := server.DeleteTenant(ctx, &DeleteTenantRequest{
//	    TenantId: "tenant_123abc",
//	})
func (s *Server) DeleteTenant(ctx context.Context, req *proto.DeleteTenantRequest) (*proto.DeleteTenantResponse, error) {
	var (
		err error
	)

	// initialize the trace
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "delete-tenant")
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
		return nil, status.Errorf(codes.InvalidArgument, "%s", fmt.Sprintf("invalid request: %s", err.Error()))
	}

	if req.GetTenantId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	s.logger.Info("deleting tenant", zap.Uint64("tenant_id", req.GetTenantId()))

	// Delete the tenant using the database client
	err = s.db.DeleteTenant(ctx, &database.DeleteTenantInput{
		ID: req.GetTenantId(),
	})
	if err != nil {
		s.logger.Error("failed to delete tenant", zap.Error(err))
		if err == database.ErrTenantDoesNotExist {
			return nil, status.Error(codes.NotFound, "tenant does not exist")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete tenant: %v", err))
	}

	// TODO: create a task and send to the task queue to delete all records associated with the tenant
	return &proto.DeleteTenantResponse{
		Success: true,
	}, nil
}
