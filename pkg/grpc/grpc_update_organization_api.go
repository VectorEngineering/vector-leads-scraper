package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateOrganization modifies an organization's configuration and settings.
// Only provided fields will be updated.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the organization ID and fields to update
//
// Returns:
//   - UpdateOrganizationResponse: Updated organization information
//   - error: Any error encountered during update
//
// Required permissions:
//   - update:organization
//
// Example:
//
//	resp, err := server.UpdateOrganization(ctx, &UpdateOrganizationRequest{
//	    OrganizationId: "org_123abc",
//	    Name: "ACME Corporation",
//	    Settings: &OrganizationSettings{
//	        MaxUsers: 200,
//	        Features: []string{"advanced_analytics", "custom_workflows"},
//	    },
//	})
func (s *Server) UpdateOrganization(ctx context.Context, req *proto.UpdateOrganizationRequest) (*proto.UpdateOrganizationResponse, error) {
	var (
		err error
	)

	// initialize the trace
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "update-organization")
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

	if req.GetOrganization() == nil {
		return nil, status.Error(codes.InvalidArgument, "organization is required")
	}

	s.logger.Info("updating organization", zap.Uint64("organization_id", req.GetOrganization().GetId()))

	// Update the organization using the database client
	organization, err := s.db.UpdateOrganization(ctx, &database.UpdateOrganizationInput{
		ID:          req.GetOrganization().GetId(),
		Name:        req.GetOrganization().GetName(),
		Description: req.GetOrganization().GetDescription(),
	})
	if err != nil {
		s.logger.Error("failed to update organization", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update organization")
	}

	return &proto.UpdateOrganizationResponse{
		Organization: organization,
	}, nil
} 