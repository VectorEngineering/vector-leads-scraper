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

// DeleteOrganization permanently removes an organization and its resources.
// This operation removes organization data, deletes member associations,
// cleans up resources, and archives audit logs.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the organization ID to delete
//
// Returns:
//   - DeleteOrganizationResponse: Confirmation of deletion
//   - error: Any error encountered during deletion
//
// Required permissions:
//   - delete:organization
//
// Example:
//
//	resp, err := server.DeleteOrganization(ctx, &DeleteOrganizationRequest{
//	    OrganizationId: "org_123abc",
//	})
func (s *Server) DeleteOrganization(ctx context.Context, req *proto.DeleteOrganizationRequest) (*proto.DeleteOrganizationResponse, error) {
	var (
		err error
	)

	// initialize the trace
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "delete-organization")
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

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "organization ID is required")
	}

	s.logger.Info("deleting organization", zap.Uint64("organization_id", req.GetId()))

	// Delete the organization using the database client
	err = s.db.DeleteOrganization(ctx, &database.DeleteOrganizationInput{
		ID: req.GetId(),
	})
	if err != nil {
		s.logger.Error("failed to delete organization", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete organization")
	}

	// TODO: create a task and send to the task queue to delete all records associated with the organization
	return &proto.DeleteOrganizationResponse{
		Success: true,
	}, nil
} 