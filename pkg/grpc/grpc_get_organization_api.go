package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetOrganization retrieves detailed information about a specific organization,
// including its configuration, members, and resource usage.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the organization ID to retrieve
//
// Returns:
//   - GetOrganizationResponse: Detailed organization information
//   - error: Any error encountered during retrieval
//
// Required permissions:
//   - read:organization
//
// Example:
//
//	resp, err := server.GetOrganization(ctx, &GetOrganizationRequest{
//	    OrganizationId: "org_123abc",
//	})
func (s *Server) GetOrganization(ctx context.Context, req *proto.GetOrganizationRequest) (*proto.GetOrganizationResponse, error) {
	var (
		err error
	)

	// initialize the trace
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "get-organization")
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

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "organization ID is required")
	}

	s.logger.Info("getting organization", zap.Uint64("organization_id", req.GetId()))

	// Get the organization using the database client
	organization, err := s.db.GetOrganization(ctx, &database.GetOrganizationInput{
		ID: req.GetId(),
	})
	if err != nil {
		s.logger.Error("failed to get organization", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get organization")
	}

	return &proto.GetOrganizationResponse{
		Organization: organization,
	}, nil
}
