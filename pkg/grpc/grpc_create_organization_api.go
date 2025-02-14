package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// Default upper bounds for each organization.
	MaxTenantsPerOrganization = 3
	MaxApiKeysPerOrganization = 10
	MaxUsersPerOrganization   = 100
)

// CreateOrganization establishes a new organization within a tenant.
// Organizations represent business units that can manage their own users,
// have separate billing, and maintain isolated resources.
//
// Required permissions: create:organization
//
// Example:
//
//	resp, err := server.CreateOrganization(ctx, &proto.CreateOrganizationRequest{
//	    Organization: &proto.Organization{
//	        Name: "ACME Corp",
//	        Description: "Leading provider of gadgets",
//	        // ... other fields, if any
//	    },
//	    TenantId: "tenant_123",
//	})
func (s *Server) CreateOrganization(ctx context.Context, req *proto.CreateOrganizationRequest) (*proto.CreateOrganizationResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "create-organization")
	defer cleanup()

	// Check for nil request.
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validate the request.
	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	// Extract organization details.
	org := req.GetOrganization()

	// Set default upper bounds.
	org.MaxTenants = MaxTenantsPerOrganization
	org.MaxApiKeys = MaxApiKeysPerOrganization
	org.MaxUsers = MaxUsersPerOrganization

	logger.Info("creating organization", zap.String("organization_name", org.GetName()))

	// Create the organization using the database client.
	organization, err := s.db.CreateOrganization(ctx, &database.CreateOrganizationInput{
		Organization: org,
	})
	if err != nil {
		logger.Error("failed to create organization", zap.Error(err))
		// Wrapping the error provides additional context.
		return nil, status.Errorf(codes.Internal, "failed to create organization: %s", err.Error())
	}

	return &proto.CreateOrganizationResponse{
		Organization: organization,
	}, nil
}
